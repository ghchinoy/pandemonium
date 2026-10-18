package component

import (
	"context"

	"github.com/ghchinoy/pandemonium/pkg/types"
)

// Constant component provides fixed text context that never changes.
type Constant struct {
	CompName string
	Value    string
}

// Name returns the name of the component.
func (c *Constant) Name() string { return c.CompName }

// PreAct returns the constant value as context for an action.
func (c *Constant) PreAct(_ context.Context, _ types.ActionSpec) (string, error) {
	return c.Value, nil
}

// PostAct is a no-op for Constant.
func (c *Constant) PostAct(ctx context.Context, actionAttempt string) (string, error) {
	return "", nil
}

// PreObserve is a no-op for Constant.
func (c *Constant) PreObserve(ctx context.Context, observation string) (string, error) {
	return "", nil
}

// PostObserve is a no-op for Constant.
func (c *Constant) PostObserve(ctx context.Context) (string, error) {
	return "", nil
}

// Update is a no-op for Constant.
func (c *Constant) Update(ctx context.Context) error { return nil }
