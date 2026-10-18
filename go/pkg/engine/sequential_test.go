package engine_test

import (
	"context"
	"path/filepath"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ghchinoy/pandemonium/pkg/engine"
	"github.com/ghchinoy/pandemonium/pkg/prefab"
	"github.com/ghchinoy/pandemonium/pkg/types"
	"github.com/ghchinoy/pandemonium/pkg/vector"
)

func TestSequential_Run(t *testing.T) {
	ctx := t.Context()

	tests := []struct {
		name          string
		strategy      types.TurnTakingStrategy
		minSteps      int
		maxSteps      int
		gmNextResp    string
		gmTerminate   string
		wantActCounts map[string]int
	}{
		{
			name:        "sequential_strategy_with_gm_substring_fallback",
			strategy:    types.StrategySequential,
			minSteps:    5,
			maxSteps:    2,
			gmNextResp:  "Next should be Bob.",
			gmTerminate: "No",
			wantActCounts: map[string]int{
				"Alice": 0,
				"Bob":   2,
			},
		},
		{
			name:        "rotation_strategy_alternates_players",
			strategy:    types.StrategyRotation,
			minSteps:    5,
			maxSteps:    3,
			gmTerminate: "No",
			wantActCounts: map[string]int{
				"Alice": 2,
				"Bob":   1,
			},
		},
		{
			name:        "terminates_early_at_min_steps",
			strategy:    types.StrategyRotation,
			minSteps:    1,
			maxSteps:    5,
			gmTerminate: "Yes",
			wantActCounts: map[string]int{
				"Alice": 1,
				"Bob":   0,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bus := types.NewEventBus()
			defer bus.Close()

			alice := &scriptedPlayer{name: "Alice"}
			bob := &scriptedPlayer{name: "Bob"}
			gm := &scriptedGM{
				nextActorResp: tc.gmNextResp,
				terminateResp: tc.gmTerminate,
			}

			seq := &engine.Sequential{
				Bus:      bus,
				MinSteps: tc.minSteps,
				Strategy: tc.strategy,
			}

			err := seq.Run(ctx, gm, []types.Entity{alice, bob}, "Test premise", tc.maxSteps)
			if err != nil {
				t.Fatalf("seq.Run(ctx, gm, players, \"Test premise\", %d) returned unexpected error: %v", tc.maxSteps, err)
			}

			gotCounts := map[string]int{
				"Alice": alice.TotalActs(),
				"Bob":   bob.TotalActs(),
			}
			if diff := cmp.Diff(tc.wantActCounts, gotCounts); diff != "" {
				t.Errorf("seq.Run() player act counts diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSequential_CheckpointAndResume(t *testing.T) {
	ctx := t.Context()
	cpPath := filepath.Join(t.TempDir(), "checkpoint.json")

	model := &stubLLM{}
	embedder := &stubEmbedder{}
	alice1 := prefab.NewThreeKeyQuestionsAgentWithGoal(ctx, "Alice", "Optimistic baker", "Share the cookie", model, embedder, vector.NewMemoryStore())
	bob1 := prefab.NewThreeKeyQuestionsAgentWithGoal(ctx, "Bob", "Skeptical journalist", "Investigate the storm", model, embedder, vector.NewMemoryStore())
	gm1 := &scriptedGM{terminateResp: "No"}

	seq1 := &engine.Sequential{
		MinSteps:       5,
		Strategy:       types.StrategyRotation,
		CheckpointPath: cpPath,
	}
	if err := seq1.Run(ctx, gm1, []types.Entity{alice1, bob1}, "Snowed in at the pub.", 2); err != nil {
		t.Fatalf("seq1.Run(maxSteps=2) returned unexpected error: %v", err)
	}

	cp, err := engine.LoadCheckpoint(cpPath)
	if err != nil {
		t.Fatalf("LoadCheckpoint(%q) returned unexpected error: %v", cpPath, err)
	}
	if cp.CompletedSteps != 2 {
		t.Errorf("LoadCheckpoint(%q).CompletedSteps = %d, want 2", cpPath, cp.CompletedSteps)
	}
	if cp.LastActor != "Bob" {
		t.Errorf("LoadCheckpoint(%q).LastActor = %q, want %q", cpPath, cp.LastActor, "Bob")
	}

	// Resume from step 2 to step 3 in fresh entity instances
	alice2 := prefab.NewThreeKeyQuestionsAgentWithGoal(ctx, "Alice", "Optimistic baker", "Share the cookie", model, embedder, vector.NewMemoryStore())
	bob2 := prefab.NewThreeKeyQuestionsAgentWithGoal(ctx, "Bob", "Skeptical journalist", "Investigate the storm", model, embedder, vector.NewMemoryStore())
	gm2 := &scriptedGM{terminateResp: "No"}

	seq2 := &engine.Sequential{
		MinSteps: 5,
		Strategy: types.StrategyRotation,
	}
	seq2.Restore(cp, gm2, []types.Entity{alice2, bob2})

	if err := seq2.Run(ctx, gm2, []types.Entity{alice2, bob2}, "Snowed in at the pub.", 3); err != nil {
		t.Fatalf("seq2.Run(maxSteps=3) after Restore returned unexpected error: %v", err)
	}
	if seq2.LastActor != "Alice" {
		t.Errorf("seq2.LastActor after step 3 = %q, want %q", seq2.LastActor, "Alice")
	}
}

func TestSequential_SkipStepAndGMSwitcher(t *testing.T) {
	ctx := t.Context()
	bus := types.NewEventBus()
	defer bus.Close()

	alice := &scriptedPlayer{name: "Alice"}
	gm1 := &scriptedGM{
		name:          "ExplorationGM",
		nextActorResp: string(types.OutputSkipThisStep),
		nextGMResp:    "CombatGM",
		terminateResp: "No",
	}
	gm2 := &scriptedGM{
		name:          "CombatGM",
		nextActorResp: "Alice",
		nextGMResp:    "CombatGM",
		terminateResp: "No",
	}

	seq := &engine.Sequential{
		Bus:      bus,
		MinSteps: 5,
		Strategy: types.StrategySequential,
		GMs:      []types.Entity{gm1, gm2},
	}

	// Step 1: gm1 switches to gm2; gm2 returns Alice; wait: let's test skip step directly when gm2 also skips on step 1 or gm1 is sole GM
	if err := seq.Run(ctx, gm1, []types.Entity{alice}, "Premise", 1); err != nil {
		t.Fatalf("seq.Run() with GM switch returned unexpected error: %v", err)
	}
	if alice.TotalActs() != 1 {
		t.Errorf("alice.TotalActs() after GM switch to CombatGM = %d, want 1", alice.TotalActs())
	}

	// Test OutputSkipThisStep and verify checkpoint is written on skipped steps (Concordia PR #411)
	skipCPPath := filepath.Join(t.TempDir(), "skip_cp.json")
	skipSeq := &engine.Sequential{
		Bus:            bus,
		MinSteps:       5,
		Strategy:       types.StrategySequential,
		CheckpointPath: skipCPPath,
	}
	if err := skipSeq.Run(ctx, gm1, []types.Entity{alice}, "Premise", 2); err != nil {
		t.Fatalf("skipSeq.Run() with OutputSkipThisStep returned unexpected error: %v", err)
	}
	if alice.TotalActs() != 1 {
		t.Errorf("alice.TotalActs() after skipped step = %d, want 1 (unchanged)", alice.TotalActs())
	}
	skipCP, err := engine.LoadCheckpoint(skipCPPath)
	if err != nil {
		t.Fatalf("LoadCheckpoint(%q) after skipped steps returned unexpected error: %v", skipCPPath, err)
	}
	if skipCP.CompletedSteps != 2 {
		t.Errorf("LoadCheckpoint(%q).CompletedSteps after 2 skipped steps = %d, want 2", skipCPPath, skipCP.CompletedSteps)
	}
}

func TestSimultaneous_Run(t *testing.T) {
	ctx := t.Context()
	bus := types.NewEventBus()
	defer bus.Close()

	alice := &scriptedPlayer{name: "Alice"}
	bob := &scriptedPlayer{name: "Bob"}
	gm := &scriptedGM{terminateResp: "No"}

	sim := &engine.Simultaneous{
		Bus:      bus,
		MinSteps: 5,
	}
	if err := sim.Run(ctx, gm, []types.Entity{alice, bob}, "Simultaneous premise", 2); err != nil {
		t.Fatalf("sim.Run(ctx, gm, players, \"Simultaneous premise\", 2) returned unexpected error: %v", err)
	}

	gotCounts := map[string]int{
		"Alice": alice.TotalActs(),
		"Bob":   bob.TotalActs(),
	}
	wantCounts := map[string]int{
		"Alice": 2,
		"Bob":   2,
	}
	if diff := cmp.Diff(wantCounts, gotCounts); diff != "" {
		t.Errorf("sim.Run() player act counts diff (-want +got):\n%s", diff)
	}
}

func TestAsynchronous_Run(t *testing.T) {
	ctx := t.Context()
	bus := types.NewEventBus()
	defer bus.Close()

	alice := &scriptedPlayer{name: "Alice"}
	bob := &scriptedPlayer{name: "Bob"}
	gm := &scriptedGM{terminateResp: "No"}

	asyncEng := &engine.Asynchronous{
		Bus:      bus,
		MinSteps: 10,
	}
	if err := asyncEng.Run(ctx, gm, []types.Entity{alice, bob}, "Async premise", 2); err != nil {
		t.Fatalf("asyncEng.Run(ctx, gm, players, \"Async premise\", 2) returned unexpected error: %v", err)
	}

	gotCounts := map[string]int{
		"Alice": alice.TotalActs(),
		"Bob":   bob.TotalActs(),
	}
	wantCounts := map[string]int{
		"Alice": 2,
		"Bob":   2,
	}
	if diff := cmp.Diff(wantCounts, gotCounts); diff != "" {
		t.Errorf("asyncEng.Run() player act counts diff (-want +got):\n%s", diff)
	}
}

func TestRulebookGM_SequentialRun(t *testing.T) {
	ctx := t.Context()
	rules := &counterRulebook{targetSteps: 2}
	gm := engine.NewRulebookGM("Referee", rules)

	alice := &scriptedPlayer{name: "Alice"}
	bob := &scriptedPlayer{name: "Bob"}

	seq := &engine.Sequential{
		MinSteps: 1,
		Strategy: types.StrategySequential,
	}
	if err := seq.Run(ctx, gm, []types.Entity{alice, bob}, "Race to two points.", 5); err != nil {
		t.Fatalf("seq.Run() with RulebookGM returned unexpected error: %v", err)
	}

	if rules.resolutions != 2 {
		t.Errorf("rules.resolutions = %d, want 2 (stopped at targetSteps)", rules.resolutions)
	}
	if len(gm.History()) != 3 { // 1 premise + 2 resolved events
		t.Errorf("len(gm.History()) = %d, want 3", len(gm.History()))
	}
}

type counterRulebook struct {
	turn        int
	resolutions int
	targetSteps int
}

func (c *counterRulebook) NextActing(_ context.Context, options []string) (string, error) {
	actor := options[c.turn%len(options)]
	c.turn++
	return actor, nil
}

func (c *counterRulebook) MakeObservation(_ context.Context, _ string) (string, error) {
	return "Scoreboard shows current progress.", nil
}

func (c *counterRulebook) Resolve(_ context.Context, _ string) (string, error) {
	c.resolutions++
	return "Point scored.", nil
}

func (c *counterRulebook) ShouldTerminate(_ context.Context) (bool, error) {
	return c.resolutions >= c.targetSteps, nil
}

type stubLLM struct{}

func (s *stubLLM) Generate(_ context.Context, _ string) (string, error) {
	return "thoughtful response", nil
}

func (s *stubLLM) GenerateWithSystem(_ context.Context, _, _ string) (string, error) {
	return "thoughtful response", nil
}

type stubEmbedder struct{}

func (s *stubEmbedder) Embed(_ context.Context, _ string) ([]float64, error) {
	return []float64{1, 0, 0}, nil
}

type scriptedPlayer struct {
	mu       sync.Mutex
	name     string
	actCount int
}

func (p *scriptedPlayer) Name() string { return p.name }

func (p *scriptedPlayer) Observe(_ context.Context, _ string) error { return nil }

func (p *scriptedPlayer) Act(_ context.Context, _ types.ActionSpec) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.actCount++
	return p.name + " acts", nil
}

func (p *scriptedPlayer) TotalActs() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.actCount
}

type scriptedGM struct {
	name          string
	nextActorResp string
	nextGMResp    string
	terminateResp string
}

func (g *scriptedGM) Name() string {
	if g.name != "" {
		return g.name
	}
	return "GM"
}

func (g *scriptedGM) Observe(_ context.Context, _ string) error { return nil }

func (g *scriptedGM) Act(_ context.Context, spec types.ActionSpec) (string, error) {
	switch spec.OutputType {
	case types.OutputTerminate:
		return g.terminateResp, nil
	case types.OutputNextActing:
		return g.nextActorResp, nil
	case types.OutputNextGameMaster:
		return g.nextGMResp, nil
	case types.OutputMakeObservation:
		return "The room is quiet.", nil
	case types.OutputResolve:
		return "Nothing changes.", nil
	default:
		return "ok", nil
	}
}
