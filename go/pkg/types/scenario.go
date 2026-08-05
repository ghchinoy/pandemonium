// Package types provides shared interfaces and configuration structures for Pandemonium.
package types

// Scenario defines the complete configuration for a simulation.
type Scenario struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Premise     string         `yaml:"premise"`
	WorldFacts  []string       `yaml:"world_facts"`
	Agents      []AgentConfig  `yaml:"agents"`
	Settings    EngineSettings `yaml:"settings"`
}

// AgentConfig defines an agent's starting state and storage rules.

type AgentConfig struct {

	Name              string        `yaml:"name"`

	Persona           string        `yaml:"persona"`

	Goal              string        `yaml:"goal,omitempty"`

	URL               string        `yaml:"url,omitempty"`                // A2A endpoint URL

	MemoryFile        string        `yaml:"memory_file,omitempty"`        // Legacy path support

	Storage           StorageConfig `yaml:"storage,omitempty"`            // Modern strategy block

	FormativeMemories []string      `yaml:"formative_memories,omitempty"` // Initial background knowledge

}



// StorageConfig defines the persistence strategy for an agent (Memory, File, or Markdown).

type StorageConfig struct {

	Type string `yaml:"type"` // "memory", "file", "markdown"

	Path string `yaml:"path,omitempty"`

}





// TurnTakingStrategy defines how the next actor is selected.
type TurnTakingStrategy string

const (
	// StrategySequential uses the GM to decide (current default).
	StrategySequential TurnTakingStrategy = "sequential"
	// StrategyRotation strictly alternates between all players.
	StrategyRotation TurnTakingStrategy = "rotation"
	// StrategyRandom picks a player at random.
	StrategyRandom TurnTakingStrategy = "random"
)

// EngineSettings configures the simulation loop.
type EngineSettings struct {
	MinSteps int                `yaml:"min_steps"`
	MaxSteps int                `yaml:"max_steps"`
	Strategy TurnTakingStrategy `yaml:"strategy,omitempty"`
}
