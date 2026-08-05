package ui_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/a2aproject/a2a-go/a2a"
	"github.com/a2aproject/a2a-go/a2aclient"
	"github.com/a2aproject/a2a-go/a2asrv"
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

	server := httptest.Server{
		Config: &http.Server{Handler: mux},
	}
	l, err := server.Config.ConnState == nil, error(nil)
	_ = l
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
