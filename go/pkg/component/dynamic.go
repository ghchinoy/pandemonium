package component

import (
	"context"
	"fmt"
	"strings"

	"github.com/ghchinoy/pandemonium/pkg/llm"
	"github.com/ghchinoy/pandemonium/pkg/types"
)

// DynamicContextComponent queries an LLM during PreAct to generate context.
type DynamicContextComponent struct {
	CompName string
	Prompt   string
	Model    llm.Model
}

// Name returns the name of the component.
func (d *DynamicContextComponent) Name() string { return d.CompName }

// PreAct queries the LLM to generate dynamic context.
func (d *DynamicContextComponent) PreAct(ctx context.Context, _ types.ActionSpec) (string, error) {
	// Gather other component info if needed?
	// For now, simple prompt based on internal state.

	// We might want to pass the agent's recent history here.
	// For Phase 2, let's keep it simple: just the question.
	res, err := d.Model.Generate(ctx, d.Prompt)
	if err != nil {
		return "", fmt.Errorf("dynamic component %q: %w", d.CompName, err)
	}
	return strings.TrimSpace(res), nil
}

// PostAct is a no-op for DynamicContextComponent.
func (d *DynamicContextComponent) PostAct(_ context.Context, actionAttempt string) (string, error) {
	return "", nil
}

// PreObserve is a no-op for DynamicContextComponent.
func (d *DynamicContextComponent) PreObserve(_ context.Context, observation string) (string, error) {
	return "", nil
}

// PostObserve is a no-op for DynamicContextComponent.
func (d *DynamicContextComponent) PostObserve(_ context.Context) (string, error) {
	return "", nil
}

// Update is a no-op for DynamicContextComponent.
func (d *DynamicContextComponent) Update(ctx context.Context) error { return nil }
