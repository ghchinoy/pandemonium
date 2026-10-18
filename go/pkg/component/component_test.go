package component_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ghchinoy/pandemonium/pkg/component"
	"github.com/ghchinoy/pandemonium/pkg/types"
)

func TestConstant(t *testing.T) {
	ctx := t.Context()
	c := &component.Constant{
		CompName: "WorldFacts",
		Value:    "It is snowing heavily.",
	}

	if got := c.Name(); got != "WorldFacts" {
		t.Errorf("c.Name() = %q, want %q", got, "WorldFacts")
	}

	spec := types.ActionSpec{CallToAction: "What next?", OutputType: types.OutputFree}
	gotPreAct, err := c.PreAct(ctx, spec)
	if err != nil {
		t.Fatalf("c.PreAct(ctx, %+v) returned unexpected error: %v", spec, err)
	}
	if gotPreAct != c.Value {
		t.Errorf("c.PreAct(ctx, %+v) = %q, want %q", spec, gotPreAct, c.Value)
	}

	if got, err := c.PostAct(ctx, "waited"); err != nil || got != "" {
		t.Errorf("c.PostAct(ctx, %q) = (%q, %v), want (\"\", nil)", "waited", got, err)
	}
	if got, err := c.PreObserve(ctx, "wind howls"); err != nil || got != "" {
		t.Errorf("c.PreObserve(ctx, %q) = (%q, %v), want (\"\", nil)", "wind howls", got, err)
	}
	if got, err := c.PostObserve(ctx); err != nil || got != "" {
		t.Errorf("c.PostObserve(ctx) = (%q, %v), want (\"\", nil)", got, err)
	}
	if err := c.Update(ctx); err != nil {
		t.Errorf("c.Update(ctx) = %v, want nil", err)
	}
}

func TestDynamicContextComponent_PreAct(t *testing.T) {
	ctx := t.Context()
	errBoom := errors.New("llm unavailable")

	tests := []struct {
		name          string
		sources       []component.ContextComponent
		response      string
		err           error
		want          string
		wantPromptHas string
		wantErr       error
	}{
		{
			name:     "trims_whitespace",
			response: "  Calm and observant. \n",
			want:     "Calm and observant.",
		},
		{
			name: "includes_context_sources",
			sources: []component.ContextComponent{
				&component.Constant{CompName: "Goal", Value: "Find the brass key"},
			},
			response:      "Focused on the brass key.",
			want:          "Focused on the brass key.",
			wantPromptHas: "[Goal]\nFind the brass key",
		},
		{
			name:    "propagates_error",
			err:     errBoom,
			wantErr: errBoom,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stub := &stubModel{response: tc.response, err: tc.err}
			d := &component.DynamicContextComponent{
				CompName:       "SelfPerception",
				Prompt:         "Who are you?",
				Model:          stub,
				ContextSources: tc.sources,
			}
			spec := types.ActionSpec{CallToAction: "Act"}
			got, err := d.PreAct(ctx, spec)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("d.PreAct(ctx, %+v) error = %v, want %v", spec, err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("d.PreAct(ctx, %+v) = %q, want %q", spec, got, tc.want)
			}
			if tc.wantPromptHas != "" && stub.lastPrompt == "" {
				t.Errorf("d.PreAct(ctx, %+v) did not pass prompt to model", spec)
			}
		})
	}
}

type stubModel struct {
	response   string
	err        error
	lastPrompt string
}

func (s *stubModel) Generate(_ context.Context, prompt string) (string, error) {
	s.lastPrompt = prompt
	return s.response, s.err
}

func (s *stubModel) GenerateWithSystem(_ context.Context, _, prompt string) (string, error) {
	s.lastPrompt = prompt
	return s.response, s.err
}
