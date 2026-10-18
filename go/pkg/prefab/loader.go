package prefab

import (
	"context"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/ghchinoy/pandemonium/pkg/component"
	"github.com/ghchinoy/pandemonium/pkg/component/memory"
	"github.com/ghchinoy/pandemonium/pkg/entity"
	"github.com/ghchinoy/pandemonium/pkg/llm"
	"github.com/ghchinoy/pandemonium/pkg/types"
	"github.com/ghchinoy/pandemonium/pkg/vector"
)

// LoadedSimulation contains the initialized parts of a simulation.
type LoadedSimulation struct {
	GM      types.Entity
	Players []types.Entity
	Config  types.Scenario
}

// LoadScenario parses a YAML file and initializes the simulation environment.
func LoadScenario(ctx context.Context, path string, model llm.Model, embedder vector.Embedder, bus *types.EventBus) (*LoadedSimulation, error) {
	var data []byte
	var err error

	// 1. Try to load from disk first
	data, err = os.ReadFile(path)
	if err != nil {
		// 2. Fallback to embedded scenarios
		embeddedPath := path
		if !strings.HasPrefix(path, "scenarios/") {
			embeddedPath = "scenarios/" + path
		}
		data, err = BuiltinScenarios.ReadFile(embeddedPath)
		if err != nil {
			return nil, fmt.Errorf("could not find scenario file %q on disk or in builtin templates: %w", path, err)
		}
	}

	var scenario types.Scenario
	if err := yaml.Unmarshal(data, &scenario); err != nil {
		return nil, fmt.Errorf("parsing scenario YAML: %w", err)
	}

	// 1. Initialize Players
	players := make([]types.Entity, len(scenario.Agents))
	for i, ac := range scenario.Agents {
		var agent types.Entity
		if ac.URL != "" {
			// Remote A2A Agent
			agent, err = entity.NewA2AProxyEntity(ctx, ac.Name, ac.URL, bus)
			if err != nil {
				return nil, fmt.Errorf("loading remote agent %q: %w", ac.Name, err)
			}
		} else {
			// Local Agent with pluggable storage
			store, err := NewStore(ctx, ac)
			if err != nil {
				return nil, fmt.Errorf("initializing store for agent %q: %w", ac.Name, err)
			}
			agent = NewThreeKeyQuestionsAgentWithGoal(ctx, ac.Name, ac.Persona, ac.Goal, model, embedder, store)

			// Seed Formative Memories if present
			for _, m := range ac.FormativeMemories {
				vec, err := embedder.Embed(ctx, m)
				if err != nil {
					return nil, fmt.Errorf("embedding formative memory for %q: %w", ac.Name, err)
				}
				if err := store.Add(ctx, vector.Item{Text: m, Embedding: vec}); err != nil {
					return nil, fmt.Errorf("storing formative memory for %q: %w", ac.Name, err)
				}
			}

			// Set the Bus for Trace emission
			if base, ok := agent.(*entity.BaseEntity); ok {
				base.Bus = bus
			}
		}
		players[i] = agent
	}

	// 2. Initialize Game Master
	gmStore := vector.NewMemoryStore()
	gmMem := &memory.AssociativeMemory{
		Store:    gmStore,
		Embedder: embedder,
		K:        10,
	}

	worldFacts := strings.Join(scenario.WorldFacts, " ")
	gm := &entity.BaseEntity{
		EntityName: "GM",
		Bus:        bus,
		Components: map[string]component.ContextComponent{
			"Instructions": &component.Constant{
				CompName: "Instructions",
				Value:    "You are the Game Master of a generative social simulation. Your role is to mediate interactions, provide sensory observations for players, and resolve their actions into environmental outcomes. Maintain a consistent world state and ensure all players have an opportunity to participate.",
			},
			"WorldMemory": gmMem,
			"WorldFacts": &component.Constant{
				CompName: "WorldFacts",
				Value:    worldFacts,
			},
		},
		ActComponent: &ConcatActComponent{
			Model:          model,
			ComponentOrder: []string{"Instructions", "WorldFacts", "WorldMemory"},
		},
	}

	return &LoadedSimulation{
		GM:      gm,
		Players: players,
		Config:  scenario,
	}, nil
}

// NewStore creates a vector.Store based on the agent's storage configuration.
func NewStore(ctx context.Context, ac types.AgentConfig) (vector.Store, error) {
	store := vector.NewMemoryStore()

	// Legacy memory_file flag
	if ac.MemoryFile != "" {
		return vector.NewFileStore(ctx, ac.MemoryFile, store)
	}

	// Modern Storage block
	switch ac.Storage.Type {
	case "file":
		return vector.NewFileStore(ctx, ac.Storage.Path, store)
	case "markdown":
		return vector.NewMarkdownStore(ac.Storage.Path)
	default:
		return store, nil
	}
}
