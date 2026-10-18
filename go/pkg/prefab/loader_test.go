package prefab_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghchinoy/pandemonium/pkg/prefab"
	"github.com/ghchinoy/pandemonium/pkg/types"
)

func TestLoadScenario_BuiltinAndCustom(t *testing.T) {
	ctx := t.Context()
	bus := types.NewEventBus()
	defer bus.Close()

	model := &recordingModel{response: "test response"}
	embedder := &fixedEmbedder{}

	// 1. Load built-in pub.yaml
	sim, err := prefab.LoadScenario(ctx, "pub.yaml", model, embedder, bus)
	if err != nil {
		t.Fatalf("LoadScenario(ctx, \"pub.yaml\", ...) returned unexpected error: %v", err)
	}
	if sim.Config.Name != "The Snowed-in Pub" {
		t.Errorf("LoadScenario(ctx, \"pub.yaml\").Config.Name = %q, want %q", sim.Config.Name, "The Snowed-in Pub")
	}
	if len(sim.Players) != 3 {
		t.Errorf("len(LoadScenario(ctx, \"pub.yaml\").Players) = %d, want 3", len(sim.Players))
	}

	// 2. Load custom YAML with FormativeMemories and file storage
	tmpDir := t.TempDir()
	memPath := filepath.Join(tmpDir, "alice_mem.jsonl")
	yamlPath := filepath.Join(tmpDir, "custom.yaml")
	yamlContent := `name: "Custom Scenario"
premise: "A quiet library."
world_facts:
  - "Books line the walls."
agents:
  - name: "Alice"
    persona: "A curious archivist."
    storage:
      type: "file"
      path: "` + memPath + `"
    formative_memories:
      - "Alice studied ancient maps."
settings:
  min_steps: 1
  max_steps: 2
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("os.WriteFile(%q) returned unexpected error: %v", yamlPath, err)
	}

	customSim, err := prefab.LoadScenario(ctx, yamlPath, model, embedder, bus)
	if err != nil {
		t.Fatalf("LoadScenario(ctx, %q, ...) returned unexpected error: %v", yamlPath, err)
	}
	if len(customSim.Players) != 1 || customSim.Players[0].Name() != "Alice" {
		t.Errorf("LoadScenario(ctx, %q) players = %+v, want 1 player named \"Alice\"", yamlPath, customSim.Players)
	}
}

func TestConcatActComponent_GetActionAttempt(t *testing.T) {
	ctx := t.Context()
	model := &recordingModel{response: "I open the book."}

	comp := &prefab.ConcatActComponent{
		Model:          model,
		ComponentOrder: []string{"Instructions", "Memory", "Timeline"},
	}

	contexts := map[string]string{
		"Instructions": "You are Alice.",
		"Memory":       "Saw a locked chest.",
		"Timeline":     "1. Entered room.",
	}
	spec := types.ActionSpec{
		CallToAction: "What does Alice do?",
		OutputType:   types.OutputChoice,
		Options:      []string{"open chest", "leave room"},
	}

	got, err := comp.GetActionAttempt(ctx, contexts, spec)
	if err != nil {
		t.Fatalf("comp.GetActionAttempt(ctx, contexts, %+v) returned unexpected error: %v", spec, err)
	}
	if got != "I open the book." {
		t.Errorf("comp.GetActionAttempt(ctx, contexts, %+v) = %q, want %q", spec, got, "I open the book.")
	}
	if model.lastSystem != "You are Alice." {
		t.Errorf("comp.GetActionAttempt() system prompt = %q, want %q", model.lastSystem, "You are Alice.")
	}
	for _, section := range []string{"## Memory\nSaw a locked chest.", "## Timeline\n1. Entered room.", "Task: What does Alice do?", "Allowed options (reply with one): open chest, leave room"} {
		if !strings.Contains(model.lastPrompt, section) {
			t.Errorf("comp.GetActionAttempt() prompt = %q, want substring %q", model.lastPrompt, section)
		}
	}
}

type recordingModel struct {
	response   string
	lastSystem string
	lastPrompt string
}

func (m *recordingModel) Generate(ctx context.Context, prompt string) (string, error) {
	return m.GenerateWithSystem(ctx, "", prompt)
}

func (m *recordingModel) GenerateWithSystem(_ context.Context, system, prompt string) (string, error) {
	m.lastSystem = system
	m.lastPrompt = prompt
	return m.response, nil
}

type fixedEmbedder struct{}

func (e *fixedEmbedder) Embed(_ context.Context, _ string) ([]float64, error) {
	return []float64{1, 0, 0}, nil
}
