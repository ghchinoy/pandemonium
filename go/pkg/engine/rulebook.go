package engine

import (
	"context"
	"slices"
	"sync"

	"github.com/ghchinoy/pandemonium/pkg/types"
)

// Rulebook defines deterministic Go game rules for a Game Master that does not
// require LLM calls (ported from Concordia PR #419's Rulebook protocol).
type Rulebook interface {
	// NextActing returns the name of the next player to act (or "skip_this_step").
	NextActing(ctx context.Context, options []string) (string, error)
	// MakeObservation returns the observation string for the active player.
	MakeObservation(ctx context.Context, callToAction string) (string, error)
	// Resolve resolves a putative player action into a world event outcome.
	Resolve(ctx context.Context, callToAction string) (string, error)
	// ShouldTerminate reports whether the simulation has reached a terminal state.
	ShouldTerminate(ctx context.Context) (bool, error)
}

// RulebookGM wraps a deterministic Rulebook as a Pandemonium Entity so any
// Engine (Sequential, Simultaneous, Asynchronous) can orchestrate it with zero LLM calls.
type RulebookGM struct {
	GMName   string
	Rules    Rulebook
	mu       sync.RWMutex
	timeline []string
}

// NewRulebookGM creates a new deterministic RulebookGM entity.
func NewRulebookGM(name string, rules Rulebook) *RulebookGM {
	if name == "" {
		name = "RulebookGM"
	}
	return &RulebookGM{
		GMName: name,
		Rules:  rules,
	}
}

// Name returns the name of the Game Master entity.
func (r *RulebookGM) Name() string {
	if r.GMName != "" {
		return r.GMName
	}
	return "RulebookGM"
}

// Observe records world premises and resolved events in the rulebook GM's history.
func (r *RulebookGM) Observe(_ context.Context, observation string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.timeline = append(r.timeline, observation)
	return nil
}

// History returns a copy of all observations recorded by the RulebookGM.
func (r *RulebookGM) History() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return slices.Clone(r.timeline)
}

// Act dispatches the engine's ActionSpec to the underlying deterministic Rulebook method.
func (r *RulebookGM) Act(ctx context.Context, spec types.ActionSpec) (string, error) {
	switch spec.OutputType {
	case types.OutputTerminate:
		done, err := r.Rules.ShouldTerminate(ctx)
		if err != nil {
			return "", err
		}
		if done {
			return "Yes", nil
		}
		return "No", nil
	case types.OutputNextActing:
		return r.Rules.NextActing(ctx, spec.Options)
	case types.OutputMakeObservation:
		return r.Rules.MakeObservation(ctx, spec.CallToAction)
	case types.OutputResolve:
		return r.Rules.Resolve(ctx, spec.CallToAction)
	default:
		return "", nil
	}
}
