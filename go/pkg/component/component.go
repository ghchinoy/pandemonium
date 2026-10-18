// Package component defines the foundational interfaces for Pandemonium entities.
package component

import (
	"context"

	"github.com/ghchinoy/pandemonium/pkg/types"
)

// Component is the base interface for all entity components.
type Component interface {
	Name() string
}

// ContextComponent provides context for an entity's actions or observations.
type ContextComponent interface {
	Component
	// PreAct is called before an action to gather context.
	PreAct(ctx context.Context, spec types.ActionSpec) (string, error)
	// PostAct is called after an action attempt.
	PostAct(ctx context.Context, actionAttempt string) (string, error)
	// PreObserve is called when an observation is received.
	PreObserve(ctx context.Context, observation string) (string, error)
	// PostObserve is called after an observation has been processed.
	PostObserve(ctx context.Context) (string, error)
	// Update allows the component to refresh its internal state.
	Update(ctx context.Context) error
}

// ActingComponent is a privileged component that decides the final action.
type ActingComponent interface {
	Component
	// GetActionAttempt synthesizes the provided context into a final action.
	GetActionAttempt(ctx context.Context, contexts map[string]string, spec types.ActionSpec) (string, error)
}
