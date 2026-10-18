package store_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ghchinoy/pandemonium/pkg/store"
	"github.com/ghchinoy/pandemonium/pkg/types"
)

func TestLocalStore_Scenarios(t *testing.T) {
	ctx := t.Context()
	baseDir := t.TempDir()

	ls, err := store.NewLocalStore(baseDir)
	if err != nil {
		t.Fatalf("NewLocalStore(%q) returned unexpected error: %v", baseDir, err)
	}

	wantScenario := types.Scenario{
		Name:        "pub",
		Description: "A pub scenario",
		Premise:     "Snowed in.",
		WorldFacts:  []string{"Fact 1"},
		Agents: []types.AgentConfig{
			{Name: "Alice", Persona: "Baker"},
		},
		Settings: types.EngineSettings{
			MinSteps: 2,
			MaxSteps: 5,
			Strategy: types.StrategyRotation,
		},
	}

	if err := ls.SaveScenario(ctx, wantScenario); err != nil {
		t.Fatalf("ls.SaveScenario(ctx, %+v) returned unexpected error: %v", wantScenario, err)
	}

	gotScenario, err := ls.GetScenario(ctx, "pub")
	if err != nil {
		t.Fatalf("ls.GetScenario(ctx, \"pub\") returned unexpected error: %v", err)
	}
	if diff := cmp.Diff(wantScenario, gotScenario); diff != "" {
		t.Errorf("ls.GetScenario(ctx, \"pub\") returned unexpected diff (-want +got):\n%s", diff)
	}

	gotList, err := ls.ListScenarios(ctx)
	if err != nil {
		t.Fatalf("ls.ListScenarios(ctx) returned unexpected error: %v", err)
	}
	wantList := []string{"pub"}
	if diff := cmp.Diff(wantList, gotList); diff != "" {
		t.Errorf("ls.ListScenarios(ctx) returned unexpected diff (-want +got):\n%s", diff)
	}
}

func TestLocalStore_SessionEvents(t *testing.T) {
	ctx := t.Context()
	baseDir := t.TempDir()

	ls, err := store.NewLocalStore(baseDir)
	if err != nil {
		t.Fatalf("NewLocalStore(%q) returned unexpected error: %v", baseDir, err)
	}

	// Non-existent session should return nil slice and nil error.
	gotMissing, err := ls.GetSessionEvents(ctx, "missing-session")
	if err != nil {
		t.Fatalf("ls.GetSessionEvents(ctx, \"missing-session\") returned unexpected error: %v", err)
	}
	if len(gotMissing) != 0 {
		t.Errorf("ls.GetSessionEvents(ctx, \"missing-session\") = %+v, want empty slice", gotMissing)
	}

	wantEvents := []types.Event{
		{Type: types.EventSimStart, Message: "Simulation initializing..."},
		{Type: types.EventAction, Actor: "Alice", Message: "waves hello"},
	}

	for _, ev := range wantEvents {
		if err := ls.SaveEvent(ctx, "sess-1", ev); err != nil {
			t.Fatalf("ls.SaveEvent(ctx, \"sess-1\", %+v) returned unexpected error: %v", ev, err)
		}
	}

	gotEvents, err := ls.GetSessionEvents(ctx, "sess-1")
	if err != nil {
		t.Fatalf("ls.GetSessionEvents(ctx, \"sess-1\") returned unexpected error: %v", err)
	}
	if diff := cmp.Diff(wantEvents, gotEvents); diff != "" {
		t.Errorf("ls.GetSessionEvents(ctx, \"sess-1\") returned unexpected diff (-want +got):\n%s", diff)
	}
}
