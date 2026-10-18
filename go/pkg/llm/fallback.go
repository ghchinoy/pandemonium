package llm

import (
	"context"
	"sync"
)

// FallbackModel wraps a Primary Model and delegates to a Fallback Model when
// the Primary model returns an error or exceeds MaxPrimaryCalls (if > 0),
// ported from Concordia PR #419's FallbackLanguageModel.
type FallbackModel struct {
	Primary         Model
	Fallback        Model
	MaxPrimaryCalls int

	mu           sync.Mutex
	primaryCalls int
	fallbacks    int
	lastError    error
}

// Generate calls Primary.Generate and falls back to Fallback.Generate on error or budget exhaustion.
func (f *FallbackModel) Generate(ctx context.Context, prompt string) (string, error) {
	return f.GenerateWithSystem(ctx, "", prompt)
}

// GenerateWithSystem calls Primary.GenerateWithSystem and falls back to Fallback.GenerateWithSystem
// on error or when MaxPrimaryCalls is exhausted.
func (f *FallbackModel) GenerateWithSystem(ctx context.Context, system, prompt string) (string, error) {
	f.mu.Lock()
	budgetExhausted := f.MaxPrimaryCalls > 0 && f.primaryCalls >= f.MaxPrimaryCalls
	if !budgetExhausted && f.Primary != nil {
		f.primaryCalls++
		f.mu.Unlock()

		res, err := f.Primary.GenerateWithSystem(ctx, system, prompt)
		if err == nil {
			return res, nil
		}

		f.mu.Lock()
		f.lastError = err
		f.fallbacks++
		f.mu.Unlock()
		return f.Fallback.GenerateWithSystem(ctx, system, prompt)
	}

	f.fallbacks++
	f.mu.Unlock()
	return f.Fallback.GenerateWithSystem(ctx, system, prompt)
}

// Stats returns the number of primary calls, fallback calls, and the last error from Primary.
func (f *FallbackModel) Stats() (primaryCalls, fallbacks int, lastErr error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.primaryCalls, f.fallbacks, f.lastError
}
