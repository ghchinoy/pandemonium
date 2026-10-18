package ui_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/a2aproject/a2a-go/v2/a2aclient"
	"github.com/a2aproject/a2a-go/v2/a2asrv"
	"github.com/google/go-cmp/cmp"

	"github.com/ghchinoy/pandemonium/pkg/engine"
	"github.com/ghchinoy/pandemonium/pkg/entity"
	"github.com/ghchinoy/pandemonium/pkg/types"
	"github.com/ghchinoy/pandemonium/pkg/ui"
)

func TestA2AServerAndProxyInteraction(t *testing.T) {
	bus := types.NewEventBus()
	defer bus.Close()
	mock := &mockEntity{name: "TestBot", act: "I inspect the room."}

	executor := &ui.EntityExecutor{Entity: mock, Bus: bus}
	handler := a2asrv.NewHandler(executor)

	mux := http.NewServeMux()
	mux.Handle("/invoke", a2asrv.NewJSONRPCHandler(handler))

	card := &a2a.AgentCard{
		Name:         mock.Name(),
		Description:  fmt.Sprintf("Pandemonium Agent: %s", mock.Name()),
		Capabilities: a2a.AgentCapabilities{Streaming: true},
	}
	mux.Handle(a2asrv.WellKnownAgentCardPath, a2asrv.NewStaticAgentCardHandler(card))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	// Verify agent card can be fetched
	checkURL := fmt.Sprintf("%s/.well-known/agent-card.json", ts.URL)
	resp, err := http.Get(checkURL)
	if err != nil {
		t.Fatalf("http.Get(%q) returned unexpected error: %v", checkURL, err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("http.Get(%q) status = %d, want %d", checkURL, resp.StatusCode, http.StatusOK)
	}

	// Create A2AProxyEntity using NewA2AProxyEntity
	invokeURL := fmt.Sprintf("%s/invoke", ts.URL)
	proxy, err := entity.NewA2AProxyEntity(ctx, "TestBot", invokeURL, bus)
	if err != nil {
		t.Fatalf("NewA2AProxyEntity(ctx, %q, %q, bus) returned unexpected error: %v", "TestBot", invokeURL, err)
	}

	// Test Observe
	obsInput := "The sky is blue."
	if err := proxy.Observe(ctx, obsInput); err != nil {
		t.Fatalf("proxy.Observe(ctx, %q) returned unexpected error: %v", obsInput, err)
	}
	wantObs := []string{obsInput}
	if diff := cmp.Diff(wantObs, mock.obs); diff != "" {
		t.Errorf("proxy.Observe(ctx, %q) mock.obs diff (-want +got):\n%s", obsInput, diff)
	}

	// Test Act
	spec := types.ActionSpec{CallToAction: "What do you do?"}
	gotAction, err := proxy.Act(ctx, spec)
	if err != nil {
		t.Fatalf("proxy.Act(ctx, %+v) returned unexpected error: %v", spec, err)
	}
	wantAction := "I inspect the room."
	if gotAction != wantAction {
		t.Errorf("proxy.Act(ctx, %+v) = %q, want %q", spec, gotAction, wantAction)
	}
}

func TestA2ACLIIntegration(t *testing.T) {
	cliPath := resolveA2ACLI(t)

	bus := types.NewEventBus()
	defer bus.Close()
	mock := &mockEntity{name: "PandemoniumCLIEntity", act: "I am responding to a2acli."}

	executor := &ui.EntityExecutor{Entity: mock, Bus: bus}
	handler := a2asrv.NewHandler(executor)

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen(\"tcp\", \"127.0.0.1:0\") returned unexpected error: %v", err)
	}
	serviceURL := fmt.Sprintf("http://%s", l.Addr().String())

	mux := http.NewServeMux()
	mux.Handle("/", a2asrv.NewJSONRPCHandler(handler))

	card := &a2a.AgentCard{
		Name:        mock.Name(),
		Description: fmt.Sprintf("Pandemonium Agent: %s", mock.Name()),
		SupportedInterfaces: []*a2a.AgentInterface{
			a2a.NewAgentInterface(serviceURL+"/", a2a.TransportProtocolJSONRPC),
		},
		Capabilities: a2a.AgentCapabilities{Streaming: true},
	}
	mux.Handle(a2asrv.WellKnownAgentCardPath, a2asrv.NewStaticAgentCardHandler(card))

	srv := &http.Server{Handler: mux}
	go func() {
		if err := srv.Serve(l); err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Errorf("srv.Serve(%q) returned unexpected error: %v", l.Addr(), err)
		}
	}()
	defer func() {
		_ = srv.Close()
	}()

	// 1. Test a2acli discover
	cmdDiscover := exec.Command(cliPath, "discover", "--service-url", serviceURL, "--output", "json")
	outDiscover, err := cmdDiscover.CombinedOutput()
	if err != nil {
		t.Fatalf("a2acli discover --service-url %q failed: %v\nOutput: %s", serviceURL, err, string(outDiscover))
	}
	gotDiscover := string(outDiscover)
	if !strings.Contains(gotDiscover, "PandemoniumCLIEntity") {
		t.Errorf("a2acli discover --service-url %q = %q, want substring %q", serviceURL, gotDiscover, "PandemoniumCLIEntity")
	}

	// 2. Test a2acli send
	prompt := "Hello from a2acli!"
	cmdSend := exec.Command(cliPath, "send", prompt, "--service-url", serviceURL, "--output", "json")
	outSend, err := cmdSend.CombinedOutput()
	if err != nil {
		t.Fatalf("a2acli send %q --service-url %q failed: %v\nOutput: %s", prompt, serviceURL, err, string(outSend))
	}
	gotSend := string(outSend)
	wantSendSubstr := "I am responding to a2acli."
	if !strings.Contains(gotSend, wantSendSubstr) {
		t.Errorf("a2acli send %q --service-url %q = %q, want substring %q", prompt, serviceURL, gotSend, wantSendSubstr)
	}

	// Ensure a2aclient package reference remains exercised if needed by callers
	_ = a2aclient.WithJSONRPCTransport
}

func TestServerControlAndScenariosAPI(t *testing.T) {
	bus := types.NewEventBus()
	defer bus.Close()

	ctrl := engine.NewStepController(true) // start paused
	srv := ui.NewServer(bus.Subscribe(), "")
	srv.Controller = ctrl

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	// 1. Check initial paused status
	statusResp, err := http.Get(ts.URL + "/api/status")
	if err != nil {
		t.Fatalf("http.Get(%q) returned unexpected error: %v", ts.URL+"/api/status", err)
	}
	_ = statusResp.Body.Close()
	if statusResp.StatusCode != http.StatusOK {
		t.Errorf("GET /api/status status = %d, want %d", statusResp.StatusCode, http.StatusOK)
	}

	// 2. Test play control action
	playResp, err := http.Post(ts.URL+"/api/control", "application/json", strings.NewReader(`{"action":"play"}`))
	if err != nil {
		t.Fatalf("http.Post(%q, play) returned unexpected error: %v", ts.URL+"/api/control", err)
	}
	_ = playResp.Body.Close()
	if !ctrl.IsRunning() {
		t.Errorf("ctrl.IsRunning() after POST /api/control play = false, want true")
	}

	// 3. Test pause control action
	pauseResp, err := http.Post(ts.URL+"/api/control", "application/json", strings.NewReader(`{"action":"pause"}`))
	if err != nil {
		t.Fatalf("http.Post(%q, pause) returned unexpected error: %v", ts.URL+"/api/control", err)
	}
	_ = pauseResp.Body.Close()
	if !ctrl.IsPaused() {
		t.Errorf("ctrl.IsPaused() after POST /api/control pause = false, want true")
	}

	// 4. Test scenarios listing
	scenResp, err := http.Get(ts.URL + "/api/scenarios")
	if err != nil {
		t.Fatalf("http.Get(%q) returned unexpected error: %v", ts.URL+"/api/scenarios", err)
	}
	_ = scenResp.Body.Close()
	if scenResp.StatusCode != http.StatusOK {
		t.Errorf("GET /api/scenarios status = %d, want %d", scenResp.StatusCode, http.StatusOK)
	}
}

func TestA2AProxyEntity_ATETargetActorHeader(t *testing.T) {
	ctx := t.Context()
	bus := types.NewEventBus()
	defer bus.Close()

	mock := &mockEntity{name: "AXBot", act: "Running inside AX sandbox."}
	executor := &ui.EntityExecutor{Entity: mock, Bus: bus}
	handler := a2asrv.NewHandler(executor)

	innerMux := http.NewServeMux()
	innerMux.Handle("/invoke", a2asrv.NewJSONRPCHandler(handler))
	card := &a2a.AgentCard{
		Name:         mock.Name(),
		Description:  "AX Sandboxed Agent",
		Capabilities: a2a.AgentCapabilities{Streaming: true},
	}
	innerMux.Handle(a2asrv.WellKnownAgentCardPath, a2asrv.NewStaticAgentCardHandler(card))

	var receivedTargets []string
	routerMiddleware := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target := r.Header.Get("ate-target-actor")
		if target == "" {
			http.Error(w, "missing ate-target-actor header", http.StatusBadGateway)
			return
		}
		receivedTargets = append(receivedTargets, target)
		innerMux.ServeHTTP(w, r)
	})

	ts := httptest.NewServer(routerMiddleware)
	defer ts.Close()

	invokeURL := fmt.Sprintf("%s/invoke", ts.URL)
	headers := map[string]string{"ate-target-actor": "default/agent-bob"}
	proxy, err := entity.NewA2AProxyEntityWithHeaders(ctx, "AXBot", invokeURL, headers, bus)
	if err != nil {
		t.Fatalf("NewA2AProxyEntityWithHeaders(ctx, \"AXBot\", %q, %v, bus) returned unexpected error: %v", invokeURL, headers, err)
	}

	if err := proxy.Observe(ctx, "Sandbox initialized."); err != nil {
		t.Fatalf("proxy.Observe(ctx, \"Sandbox initialized.\") returned unexpected error: %v", err)
	}

	gotAct, err := proxy.Act(ctx, types.ActionSpec{CallToAction: "Report status."})
	if err != nil {
		t.Fatalf("proxy.Act(ctx, spec) returned unexpected error: %v", err)
	}
	if gotAct != "Running inside AX sandbox." {
		t.Errorf("proxy.Act(ctx, spec) = %q, want %q", gotAct, "Running inside AX sandbox.")
	}
	if len(receivedTargets) < 3 {
		t.Errorf("len(receivedTargets) = %d, want >= 3 (card discovery + observe + act)", len(receivedTargets))
	}
	for i, got := range receivedTargets {
		if got != "default/agent-bob" {
			t.Errorf("receivedTargets[%d] = %q, want %q", i, got, "default/agent-bob")
		}
	}
}

func resolveA2ACLI(t *testing.T) string {
	t.Helper()
	if cliPath, err := exec.LookPath("a2acli"); err == nil {
		return cliPath
	}
	absPath, err := filepath.Abs("../../../../a2acli/bin/a2acli")
	if err == nil {
		if cliPath, err := exec.LookPath(absPath); err == nil {
			return cliPath
		}
	}
	t.Skip("a2acli binary not found in PATH or ../../../../a2acli/bin/a2acli, skipping CLI integration test")
	return ""
}

type mockEntity struct {
	name string
	obs  []string
	act  string
}

func (m *mockEntity) Name() string { return m.name }

func (m *mockEntity) Observe(_ context.Context, observation string) error {
	m.obs = append(m.obs, observation)
	return nil
}

func (m *mockEntity) Act(_ context.Context, _ types.ActionSpec) (string, error) {
	return m.act, nil
}
