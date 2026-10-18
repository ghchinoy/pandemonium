package llm_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ghchinoy/pandemonium/pkg/llm"
)

func TestFallbackModel(t *testing.T) {
	ctx := t.Context()
	errRateLimit := errors.New("rate limit exceeded")

	t.Run("uses_primary_within_budget", func(t *testing.T) {
		fm := &llm.FallbackModel{
			Primary:         &fixedModel{out: "primary-out"},
			Fallback:        &fixedModel{out: "fallback-out"},
			MaxPrimaryCalls: 1,
		}

		got1, err := fm.Generate(ctx, "hello")
		if err != nil {
			t.Fatalf("fm.Generate(ctx, \"hello\") call 1 returned unexpected error: %v", err)
		}
		if got1 != "primary-out" {
			t.Errorf("fm.Generate(ctx, \"hello\") call 1 = %q, want %q", got1, "primary-out")
		}

		// Call 2 exceeds MaxPrimaryCalls=1 and should route to Fallback
		got2, err := fm.Generate(ctx, "hello again")
		if err != nil {
			t.Fatalf("fm.Generate(ctx, \"hello again\") call 2 returned unexpected error: %v", err)
		}
		if got2 != "fallback-out" {
			t.Errorf("fm.Generate(ctx, \"hello again\") call 2 = %q, want %q", got2, "fallback-out")
		}

		pCalls, fCalls, _ := fm.Stats()
		if pCalls != 1 || fCalls != 1 {
			t.Errorf("fm.Stats() = (%d, %d), want (1, 1)", pCalls, fCalls)
		}
	})

	t.Run("falls_back_on_primary_error", func(t *testing.T) {
		fm := &llm.FallbackModel{
			Primary:  &fixedModel{err: errRateLimit},
			Fallback: &fixedModel{out: "recovered-out"},
		}

		got, err := fm.Generate(ctx, "prompt")
		if err != nil {
			t.Fatalf("fm.Generate(ctx, \"prompt\") returned unexpected error: %v", err)
		}
		if got != "recovered-out" {
			t.Errorf("fm.Generate(ctx, \"prompt\") = %q, want %q", got, "recovered-out")
		}

		pCalls, fCalls, lastErr := fm.Stats()
		if pCalls != 1 || fCalls != 1 || !errors.Is(lastErr, errRateLimit) {
			t.Errorf("fm.Stats() = (%d, %d, %v), want (1, 1, %v)", pCalls, fCalls, lastErr, errRateLimit)
		}
	})
}

type fixedModel struct {
	out string
	err error
}

func (m *fixedModel) Generate(_ context.Context, _ string) (string, error) {
	return m.out, m.err
}

func (m *fixedModel) GenerateWithSystem(_ context.Context, _, _ string) (string, error) {
	return m.out, m.err
}
