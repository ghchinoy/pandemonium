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
			fmt.Fprintf(&user, "\n## %s\n%s\n", name, val)
		}
	}
	fmt.Fprintf(&user, "\nTask: %s", spec.CallToAction)
	if len(spec.Options) > 0 {
		fmt.Fprintf(&user, "\nAllowed options (reply with one): %s", strings.Join(spec.Options, ", "))
	}
	user.WriteString("\nResponse:")

	return c.Model.GenerateWithSystem(ctx, system, user.String())
}

// NewThreeKeyQuestionsAgent builds the standard reasoning agent.
func NewThreeKeyQuestionsAgent(ctx context.Context, name string, persona string, model llm.Model, embedder vector.Embedder, store vector.Store) types.Entity {
	return NewThreeKeyQuestionsAgentWithGoal(ctx, name, persona, "", model, embedder, store)
}

// NewThreeKeyQuestionsAgentWithGoal builds a reasoning agent with an optional overarching goal
// and all three Concordia key questions (SituationPerception, SelfPerception, PersonBySituation).
func NewThreeKeyQuestionsAgentWithGoal(_ context.Context, name string, persona string, goal string, model llm.Model, embedder vector.Embedder, store vector.Store) types.Entity {
	mem := &memory.AssociativeMemory{
		Store:    store,
		Embedder: embedder,
		K:        5,
	}
	timeline := memory.NewTimelineMemory("Timeline", 10)
	reflection := &memory.ReflectionComponent{
		CompName: "Reflection",
		Timeline: timeline,
		Store:    store,
		Embedder: embedder,
		Model:    model,
		Interval: 3,
	}

	perceptionSources := []component.ContextComponent{timeline}
	var goalComp *component.Constant
	if goal != "" {
		goalComp = &component.Constant{CompName: "Goal", Value: goal}
		perceptionSources = append([]component.ContextComponent{goalComp}, perceptionSources...)
	}

	selfPerception := &component.DynamicContextComponent{
		CompName:       "SelfPerception",
		Prompt:         fmt.Sprintf("Question: Based on %s's persona and recent history, what kind of person is %s? Persona: %s\nAnswer in one sentence.", name, name, persona),
		Model:          model,
		ContextSources: perceptionSources,
	}
	situationPerception := &component.DynamicContextComponent{
		CompName:       "SituationPerception",
		Prompt:         fmt.Sprintf("Question: Based on recent observations, what kind of situation is %s in right now?\nAnswer in one sentence.", name),
		Model:          model,
		ContextSources: perceptionSources,
	}
	personBySituation := &component.DynamicContextComponent{
		CompName:       "PersonBySituation",
		Prompt:         fmt.Sprintf("Question: What would a person like %s (%s) do in a situation like this?\nAnswer in one sentence.", name, persona),
		Model:          model,
		ContextSources: perceptionSources,
	}

	components := map[string]component.ContextComponent{
		"Instructions":        &component.Constant{CompName: "Instructions", Value: persona},
		"Memory":              mem,
		"Timeline":            timeline,
		"Reflection":          reflection,
		"SelfPerception":      selfPerception,
		"SituationPerception": situationPerception,
		"PersonBySituation":   personBySituation,
	}

	order := []string{"Instructions"}
	if goalComp != nil {
		components["Goal"] = goalComp
		order = append(order, "Goal")
	}
	order = append(order, "Memory", "Timeline", "Reflection", "SituationPerception", "SelfPerception", "PersonBySituation")

	return &entity.BaseEntity{
		EntityName: name,
		Components: components,
		ActComponent: &ConcatActComponent{
			Model:          model,
			ComponentOrder: order,
		},
	}
}
