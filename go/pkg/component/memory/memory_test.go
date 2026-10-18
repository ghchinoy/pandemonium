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

func TestReflectionComponent_SyncAndAsync(t *testing.T) {
	ctx := t.Context()
	spec := types.ActionSpec{CallToAction: "What next?"}

	for _, async := range []bool{false, true} {
		name := "sync"
		if async {
			name = "async"
		}
		t.Run(name, func(t *testing.T) {
			tm := memory.NewTimelineMemory("Timeline", 5)
			_, _ = tm.PreObserve(ctx, "The blizzard intensifies.")
			_, _ = tm.PostAct(ctx, "check the firewood")

			store := vector.NewMemoryStore()
			rc := &memory.ReflectionComponent{
				CompName: "Reflection",
				Timeline: tm,
				Store:    store,
				Embedder: &lengthEmbedder{},
				Model:    &fixedReflectionModel{insight: "We must conserve heat."},
				Interval: 2,
				Async:    async,
			}

			// Step 1: below interval (1 % 2 != 0)
			if err := rc.Update(ctx); err != nil {
				t.Fatalf("rc.Update(ctx) step 1 returned unexpected error: %v", err)
			}
			if got, _ := rc.PreAct(ctx, spec); got != "" {
				t.Errorf("rc.PreAct(ctx, %+v) at step 1 = %q, want empty", spec, got)
			}

			// Step 2: triggers reflection synthesis (2 % 2 == 0)
			if err := rc.Update(ctx); err != nil {
				t.Fatalf("rc.Update(ctx) step 2 returned unexpected error: %v", err)
			}
			rc.Wait()

			gotFresh, err := rc.PreAct(ctx, spec)
			if err != nil {
				t.Fatalf("rc.PreAct(ctx, %+v) at step 2 returned unexpected error: %v", spec, err)
			}
			wantFresh := "Recent reflection [fresh]: We must conserve heat."
			if gotFresh != wantFresh {
				t.Errorf("rc.PreAct(ctx, %+v) at step 2 = %q, want %q", spec, gotFresh, wantFresh)
			}

			// Step 3: ages by 1 turn without new reflection
			if err := rc.Update(ctx); err != nil {
				t.Fatalf("rc.Update(ctx) step 3 returned unexpected error: %v", err)
			}
			gotAged, _ := rc.PreAct(ctx, spec)
			wantAged := "Recent reflection [1 turns old]: We must conserve heat."
			if gotAged != wantAged {
				t.Errorf("rc.PreAct(ctx, %+v) at step 3 = %q, want %q", spec, gotAged, wantAged)
			}

			if len(store.Items()) != 1 {
				t.Errorf("len(store.Items()) = %d, want 1", len(store.Items()))
			}
		})
	}
}

type lengthEmbedder struct{}

func (e *lengthEmbedder) Embed(_ context.Context, text string) ([]float64, error) {
	return []float64{float64(len(text)), 1, 0}, nil
}

type fixedReflectionModel struct {
	insight string
}

func (m *fixedReflectionModel) Generate(_ context.Context, _ string) (string, error) {
	return m.insight, nil
}

func (m *fixedReflectionModel) GenerateWithSystem(_ context.Context, _, _ string) (string, error) {
	return m.insight, nil
}
