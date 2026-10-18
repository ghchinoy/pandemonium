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
	CompName       string
	Prompt         string
	Model          llm.Model
	ContextSources []ContextComponent
}

// Name returns the name of the component.
func (d *DynamicContextComponent) Name() string { return d.CompName }

// PreAct queries the LLM to generate dynamic context.
func (d *DynamicContextComponent) PreAct(ctx context.Context, spec types.ActionSpec) (string, error) {
	prompt := d.Prompt
	if len(d.ContextSources) > 0 {
		var sb strings.Builder
		for _, src := range d.ContextSources {
			if src == nil {
				continue
			}
			val, err := src.PreAct(ctx, spec)
			if err != nil {
				return "", fmt.Errorf("context source %q for %q: %w", src.Name(), d.CompName, err)
			}
			if val != "" {
				fmt.Fprintf(&sb, "[%s]\n%s\n\n", src.Name(), val)
			}
		}
		sb.WriteString(d.Prompt)
		prompt = sb.String()
	}

	res, err := d.Model.Generate(ctx, prompt)
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
