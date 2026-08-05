// Package store provides persistence for simulation state and scenarios.
package store

import (
	"context"
	"github.com/ghchinoy/pandemonium/pkg/types"
)

// StateStore defines the operations for non-vector persistence.
type StateStore interface {
	// Scenarios
	SaveScenario(ctx context.Context, s types.Scenario) error
	GetScenario(ctx context.Context, name string) (types.Scenario, error)
	ListScenarios(ctx context.Context) ([]string, error)

	// Sessions
	SaveEvent(ctx context.Context, sessionID string, event types.Event) error
	GetSessionEvents(ctx context.Context, sessionID string) ([]types.Event, error)
}
