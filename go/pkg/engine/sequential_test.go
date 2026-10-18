package engine_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ghchinoy/pandemonium/pkg/engine"
	"github.com/ghchinoy/pandemonium/pkg/types"
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
				"Alice": alice.actCount,
				"Bob":   bob.actCount,
			}
			if diff := cmp.Diff(tc.wantActCounts, gotCounts); diff != "" {
				t.Errorf("seq.Run() player act counts diff (-want +got):\n%s", diff)
			}
		})
	}
}

type scriptedPlayer struct {
	name     string
	actCount int
}

func (p *scriptedPlayer) Name() string { return p.name }

func (p *scriptedPlayer) Observe(_ context.Context, _ string) error { return nil }

func (p *scriptedPlayer) Act(_ context.Context, _ types.ActionSpec) (string, error) {
	p.actCount++
	return p.name + " acts", nil
}

type scriptedGM struct {
	nextActorResp string
	terminateResp string
}

func (g *scriptedGM) Name() string { return "GM" }

func (g *scriptedGM) Observe(_ context.Context, _ string) error { return nil }

func (g *scriptedGM) Act(_ context.Context, spec types.ActionSpec) (string, error) {
	switch spec.OutputType {
	case types.OutputTerminate:
		return g.terminateResp, nil
	case types.OutputNextActing:
		return g.nextActorResp, nil
	case types.OutputMakeObservation:
		return "The room is quiet.", nil
	case types.OutputResolve:
		return "Nothing changes.", nil
	default:
		return "ok", nil
	}
}
