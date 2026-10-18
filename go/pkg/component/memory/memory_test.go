package memory_test

import (
	"context"
	"strings"
	"testing"

	"github.com/ghchinoy/pandemonium/pkg/component/memory"
	"github.com/ghchinoy/pandemonium/pkg/types"
	"github.com/ghchinoy/pandemonium/pkg/vector"
)

func TestTimelineMemory(t *testing.T) {
	ctx := t.Context()
	tm := memory.NewTimelineMemory("Timeline", 2)
	spec := types.ActionSpec{CallToAction: "What next?"}

	gotEmpty, err := tm.PreAct(ctx, spec)
	if err != nil {
		t.Fatalf("tm.PreAct(ctx, %+v) on empty timeline returned unexpected error: %v", spec, err)
	}
	wantEmpty := "No recent events recorded."
	if gotEmpty != wantEmpty {
		t.Errorf("tm.PreAct(ctx, %+v) = %q, want %q", spec, gotEmpty, wantEmpty)
	}

	if _, err := tm.PreObserve(ctx, "Event 1"); err != nil {
		t.Fatalf("tm.PreObserve(ctx, \"Event 1\") returned unexpected error: %v", err)
	}
	if _, err := tm.PostAct(ctx, "Action 2"); err != nil {
		t.Fatalf("tm.PostAct(ctx, \"Action 2\") returned unexpected error: %v", err)
	}
	if _, err := tm.PreObserve(ctx, "Event 3"); err != nil {
		t.Fatalf("tm.PreObserve(ctx, \"Event 3\") returned unexpected error: %v", err)
	}

	got, err := tm.PreAct(ctx, spec)
	if err != nil {
		t.Fatalf("tm.PreAct(ctx, %+v) returned unexpected error: %v", spec, err)
	}
	want := "Recent Timeline:\n1. I decided to: Action 2\n2. Event 3\n"
	if got != want {
		t.Errorf("tm.PreAct(ctx, %+v) = %q, want %q", spec, got, want)
	}
}

func TestAssociativeMemory(t *testing.T) {
	ctx := t.Context()
	am := &memory.AssociativeMemory{
		Store:    vector.NewMemoryStore(),
		Embedder: &lengthEmbedder{},
		K:        2,
	}

	if _, err := am.PreObserve(ctx, "The fireplace is cold."); err != nil {
		t.Fatalf("am.PreObserve(ctx, obs) returned unexpected error: %v", err)
	}
	if _, err := am.PostAct(ctx, "light the fireplace"); err != nil {
		t.Fatalf("am.PostAct(ctx, action) returned unexpected error: %v", err)
	}

	spec := types.ActionSpec{CallToAction: "What do you see?"}
	got, err := am.PreAct(ctx, spec)
	if err != nil {
		t.Fatalf("am.PreAct(ctx, %+v) returned unexpected error: %v", spec, err)
	}

	for _, wantSubstr := range []string{"Relevant memories:", "The fireplace is cold.", "I decided to: light the fireplace"} {
		if !strings.Contains(got, wantSubstr) {
			t.Errorf("am.PreAct(ctx, %+v) = %q, want substring %q", spec, got, wantSubstr)
		}
	}
}

type lengthEmbedder struct{}

func (e *lengthEmbedder) Embed(_ context.Context, text string) ([]float64, error) {
	return []float64{float64(len(text)), 1, 0}, nil
}
