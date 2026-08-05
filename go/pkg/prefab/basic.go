// Package prefab provides pre-configured agent recipes and simulation loaders.
package prefab

import (
	"context"
	"fmt"
	"strings"

	"github.com/ghchinoy/pandemonium/pkg/component"
	"github.com/ghchinoy/pandemonium/pkg/component/memory"
	"github.com/ghchinoy/pandemonium/pkg/entity"
	"github.com/ghchinoy/pandemonium/pkg/llm"
	"github.com/ghchinoy/pandemonium/pkg/types"
	"github.com/ghchinoy/pandemonium/pkg/vector"
)

// ConcatActComponent builds a prompt by concatenating all component outputs.
type ConcatActComponent struct {
	Model          llm.Model
	ComponentOrder []string
}

// Name returns the name of the component.
func (c *ConcatActComponent) Name() string { return "ConcatAct" }

// GetActionAttempt synthesizes the provided context into a final action.
func (c *ConcatActComponent) GetActionAttempt(ctx context.Context, contexts map[string]string, spec types.ActionSpec) (string, error) {
	var system string
	var user strings.Builder

	// Convention: 'Instructions' is the System prompt
	if val, ok := contexts["Instructions"]; ok {
		system = val
	}

	for _, name := range c.ComponentOrder {
		if name == "Instructions" {
			continue
		}
		if val, ok := contexts[name]; ok && val != "" {
			user.WriteString(fmt.Sprintf("\n## %s\n%s\n", name, val))
		}
	}
	user.WriteString(fmt.Sprintf("\nTask: %s\nResponse:", spec.CallToAction))

	return c.Model.GenerateWithSystem(ctx, system, user.String())
}

// SimpleContextComponent is a helper for fixed context.
type SimpleContextComponent struct {
	CompName string
	Value    string
}

// Name returns the name of the component.
func (s *SimpleContextComponent) Name() string { return s.CompName }

// PreAct returns the fixed value.
func (s *SimpleContextComponent) PreAct(ctx context.Context, spec types.ActionSpec) (string, error) {
	return s.Value, nil
}

// PostAct is a no-op.
func (s *SimpleContextComponent) PostAct(ctx context.Context, action string) (string, error) {
	return "", nil
}

// PreObserve is a no-op.
func (s *SimpleContextComponent) PreObserve(ctx context.Context, observation string) (string, error) {
	return "", nil
}

// PostObserve is a no-op.
func (s *SimpleContextComponent) PostObserve(ctx context.Context) (string, error) {
	return "", nil
}

// Update is a no-op.
func (s *SimpleContextComponent) Update(ctx context.Context) error { return nil }

// NewThreeKeyQuestionsAgent builds the standard reasoning agent.
func NewThreeKeyQuestionsAgent(ctx context.Context, name string, persona string, model llm.Model, embedder vector.Embedder, store vector.Store) types.Entity {
	mem := &memory.AssociativeMemory{
		Store:    store,
		Embedder: embedder,
		K:        5,
	}
	timeline := memory.NewTimelineMemory("Timeline", 10)

	components := map[string]component.ContextComponent{
		"Instructions": &SimpleContextComponent{CompName: "Instructions", Value: persona},
		"Memory":       mem,
		"Timeline":     timeline,
		"SelfPerception": &component.DynamicContextComponent{
			CompName: "SelfPerception",
			Prompt:   fmt.Sprintf("Based on your persona and recent memories, who are you? Persona: %s", persona),
			Model:    model,
		},
		"SituationPerception": &component.DynamicContextComponent{
			CompName: "SituationPerception",
			Prompt:   "Based on recent observations, what kind of situation are you in right now? Describe it in one sentence.",
			Model:    model,
		},
	}

	// Decision is the final step. Hybrid approach: Instructions -> Memory -> Timeline -> Reflections
	order := []string{"Instructions", "Memory", "Timeline", "SelfPerception", "SituationPerception"}

	return &entity.BaseEntity{
		EntityName: name,
		Components: components,
		ActComponent: &ConcatActComponent{
			Model:          model,
			ComponentOrder: order,
		},
	}
}
