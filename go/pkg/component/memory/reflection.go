package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/ghchinoy/pandemonium/pkg/llm"
	"github.com/ghchinoy/pandemonium/pkg/types"
	"github.com/ghchinoy/pandemonium/pkg/vector"
)

// ReflectionComponent periodically synthesizes recent timeline events into
// high-level semantic reflections stored in the vector Store.
// When Async is true, reflection synthesis runs in the background during Update
// (mirroring Concordia v2.4's pubsub Publisher/Subscriber pattern) while PreAct
// immediately returns the latest completed reflection with its age in turns.
type ReflectionComponent struct {
	CompName string
	Timeline *TimelineMemory
	Store    vector.Store
	Embedder vector.Embedder
	Model    llm.Model
	Interval int
	Async    bool

	mu               sync.RWMutex
	stepCount        int
	latestReflection string
	reflectionStep   int
	busy             bool
	wg               sync.WaitGroup
}

// Name returns the name of the component.
func (r *ReflectionComponent) Name() string {
	if r.CompName != "" {
		return r.CompName
	}
	return "Reflection"
}

// PreAct returns the most recent synthesized reflection and its age in turns.
func (r *ReflectionComponent) PreAct(_ context.Context, _ types.ActionSpec) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.latestReflection == "" {
		return "", nil
	}
	age := max(0, r.stepCount-r.reflectionStep)
	if age == 0 {
		return fmt.Sprintf("Recent reflection [fresh]: %s", r.latestReflection), nil
	}
	return fmt.Sprintf("Recent reflection [%d turns old]: %s", age, r.latestReflection), nil
}

// PostAct is a no-op for ReflectionComponent.
func (r *ReflectionComponent) PostAct(_ context.Context, _ string) (string, error) {
	return "", nil
}

// PreObserve is a no-op for ReflectionComponent.
func (r *ReflectionComponent) PreObserve(_ context.Context, _ string) (string, error) {
	return "", nil
}

// PostObserve is a no-op for ReflectionComponent.
func (r *ReflectionComponent) PostObserve(_ context.Context) (string, error) {
	return "", nil
}

// Update increments the turn counter and triggers reflection synthesis every Interval turns.
func (r *ReflectionComponent) Update(ctx context.Context) error {
	r.mu.Lock()
	r.stepCount++
	interval := r.Interval
	if interval <= 0 {
		interval = 3
	}
	shouldReflect := r.stepCount%interval == 0
	if !shouldReflect || r.Timeline == nil || r.Model == nil {
		r.mu.Unlock()
		return nil
	}
	events := r.Timeline.Events()
	if len(events) == 0 {
		r.mu.Unlock()
		return nil
	}
	step := r.stepCount

	if r.Async {
		if r.busy {
			r.mu.Unlock()
			return nil
		}
		r.busy = true
		r.wg.Add(1)
		r.mu.Unlock()

		go func() {
			defer r.wg.Done()
			_ = r.synthesize(ctx, events, step)
			r.mu.Lock()
			r.busy = false
			r.mu.Unlock()
		}()
		return nil
	}
	r.mu.Unlock()

	return r.synthesize(ctx, events, step)
}

// Wait blocks until any in-flight asynchronous reflection synthesis completes.
func (r *ReflectionComponent) Wait() {
	r.wg.Wait()
}

func (r *ReflectionComponent) synthesize(ctx context.Context, events []string, step int) error {
	prompt := fmt.Sprintf(
		"Review these recent events and synthesize one high-level insight or reflection in a single sentence:\n%s",
		strings.Join(events, "\n"),
	)
	insight, err := r.Model.Generate(ctx, prompt)
	if err != nil {
		return fmt.Errorf("generating reflection: %w", err)
	}
	insight = strings.TrimSpace(insight)
	if insight == "" {
		return nil
	}

	if r.Store != nil && r.Embedder != nil {
		text := fmt.Sprintf("[Reflection] %s", insight)
		vec, err := r.Embedder.Embed(ctx, text)
		if err != nil {
			return fmt.Errorf("embedding reflection: %w", err)
		}
		if err := r.Store.Add(ctx, vector.Item{Text: text, Embedding: vec}); err != nil {
			return fmt.Errorf("storing reflection: %w", err)
		}
	}

	r.mu.Lock()
	r.latestReflection = insight
	r.reflectionStep = step
	r.mu.Unlock()
	return nil
}
