package ui_test

import (
	"context"
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
	"github.com/ghchinoy/pandemonium/pkg/entity"
	"github.com/ghchinoy/pandemonium/pkg/types"
	"github.com/ghchinoy/pandemonium/pkg/ui"
)

type mockEntity struct {
	name string
	obs  []string
	act  string
}

func (m *mockEntity) Name() string { return m.name }
func (m *mockEntity) Observe(ctx context.Context, observation string) error {
	m.obs = append(m.obs, observation)
	return nil
}
func (m *mockEntity) Act(ctx context.Context, spec types.ActionSpec) (string, error) {
	return m.act, nil
}

func TestA2AServerAndProxyInteraction(t *testing.T) {
	bus := types.NewEventBus()
	mock := &mockEntity{name: "TestBot", act: "I inspect the room."}

	executor := &ui.EntityExecutor{Entity: mock, Bus: bus}
	handler := a2asrv.NewHandler(executor)

	mux := http.NewServeMux()
	mux.Handle("/invoke", a2asrv.NewJSONRPCHandler(handler))

	card := &a2a.AgentCard{
		Name:        mock.Name(),
		Description: fmt.Sprintf("Pandemonium Agent: %s", mock.Name()),
		Capabilities: a2a.AgentCapabilities{Streaming: true},
	}
	mux.Handle(a2asrv.WellKnownAgentCardPath, a2asrv.NewStaticAgentCardHandler(card))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify agent card can be fetched
	checkURL := fmt.Sprintf("%s/.well-known/agent-card.json", ts.URL)
	resp, err := http.Get(checkURL)
	if err != nil {
		t.Fatalf("Failed to fetch agent card: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Agent card returned status %d", resp.StatusCode)
	}

	// Create A2AProxyEntity using client factory directly with test server URL
	factory := a2aclient.NewFactory(a2aclient.WithJSONRPCTransport(nil))
	invokeURL := fmt.Sprintf("%s/invoke", ts.URL)
	endpoints := []*a2a.AgentInterface{
		a2a.NewAgentInterface(invokeURL, a2a.TransportProtocolJSONRPC),
	}
	client, err := factory.CreateFromEndpoints(ctx, endpoints)
	if err != nil {
		t.Fatalf("Failed to create client from endpoints: %v", err)
	}

	proxy := &entity.A2AProxyEntity{
		EntityName: "TestBot",
		URL:        invokeURL,
		Client:     client,
		Bus:        bus,
	}

	// Test Observe
	err = proxy.Observe(ctx, "The sky is blue.")
	if err != nil {
		t.Fatalf("Observe failed: %v", err)
	}
	if len(mock.obs) != 1 || mock.obs[0] != "The sky is blue." {
		t.Fatalf("Unexpected observations in mock: %v", mock.obs)
	}

	// Test Act
	action, err := proxy.Act(ctx, types.ActionSpec{CallToAction: "What do you do?"})
	if err != nil {
		t.Fatalf("Act failed: %v", err)
	}
	if action != "I inspect the room." {
		t.Fatalf("Unexpected action returned: %q", action)
	}
}

func TestA2ACLIIntegration(t *testing.T) {
	absPath, err := filepath.Abs("../../../../a2acli/bin/a2acli")
	if err != nil {
		t.Skip("Failed to resolve a2acli path")
	}
	cliPath, err := exec.LookPath(absPath)
	if err != nil {
		t.Skipf("a2acli binary not found at %s, skipping CLI integration test", absPath)
	}

	bus := types.NewEventBus()
	mock := &mockEntity{name: "PandemoniumCLIEntity", act: "I am responding to a2acli."}

	executor := &ui.EntityExecutor{Entity: mock, Bus: bus}
	handler := a2asrv.NewHandler(executor)

	mux := http.NewServeMux()
	mux.Handle("/", a2asrv.NewJSONRPCHandler(handler))

	card := &a2a.AgentCard{
		Name:        mock.Name(),
		Description: fmt.Sprintf("Pandemonium Agent: %s", mock.Name()),
		SupportedInterfaces: []*a2a.AgentInterface{
			a2a.NewAgentInterface("http://127.0.0.1:9099/", a2a.TransportProtocolJSONRPC),
		},
		Capabilities: a2a.AgentCapabilities{Streaming: true},
	}
	mux.Handle(a2asrv.WellKnownAgentCardPath, a2asrv.NewStaticAgentCardHandler(card))

	l, err := net.Listen("tcp", "127.0.0.1:9099")
	if err != nil {
		t.Fatalf("Failed to listen on 9099: %v", err)
	}
	srv := &http.Server{Handler: mux}
	go srv.Serve(l)
	defer srv.Close()

	// 1. Test a2acli discover
	cmdDiscover := exec.Command(cliPath, "discover", "--service-url", "http://127.0.0.1:9099", "--output", "json")
	outDiscover, err := cmdDiscover.CombinedOutput()
	if err != nil {
		t.Fatalf("a2acli discover failed: %v\nOutput: %s", err, string(outDiscover))
	}
	if !strings.Contains(string(outDiscover), "PandemoniumCLIEntity") {
		t.Fatalf("a2acli discover output missing entity name: %s", string(outDiscover))
	}

	// 2. Test a2acli send
	cmdSend := exec.Command(cliPath, "send", "Hello from a2acli!", "--service-url", "http://127.0.0.1:9099", "--wait", "--output", "json")
	outSend, err := cmdSend.CombinedOutput()
	if err != nil {
		t.Fatalf("a2acli send failed: %v\nOutput: %s", err, string(outSend))
	}
	if !strings.Contains(string(outSend), "I am responding to a2acli.") {
		t.Fatalf("a2acli send output missing expected response: %s", string(outSend))
	}
}
