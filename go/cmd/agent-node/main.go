// Package main implements the standalone agent node for distributed simulations.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"

	"github.com/ghchinoy/pandemonium/pkg/entity"
	"github.com/ghchinoy/pandemonium/pkg/llm"
	"github.com/ghchinoy/pandemonium/pkg/prefab"
	"github.com/ghchinoy/pandemonium/pkg/types"
	"github.com/ghchinoy/pandemonium/pkg/ui"
)

func main() {
	port := flag.Int("port", 9001, "Port to host the A2A agent")
	name := flag.String("name", "Bob", "Name of the agent")
	persona := flag.String("persona", "A skeptical journalist investigating the storm.", "Agent persona")
	configPath := flag.String("config", "", "Path to agent YAML config (overrides flags)")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	ctx := context.Background()

	var ac types.AgentConfig
	ac.Name = *name
	ac.Persona = *persona

	// Load from YAML if provided
	if *configPath != "" {
		data, err := os.ReadFile(*configPath)
		if err != nil {
			log.Fatalf("Failed to read config: %v", err)
		}
		if err := yaml.Unmarshal(data, &ac); err != nil {
			log.Fatalf("Failed to parse config: %v", err)
		}
	}

	model, err := llm.NewGeminiModel(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize model: %v", err)
	}
	embedder, err := llm.NewGeminiEmbedder(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize embedder: %v", err)
	}

	// 1. Initialize the local agent with factory-created store
	bus := types.NewEventBus()
	store, err := prefab.NewStore(ctx, ac)
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}

	agent := prefab.NewThreeKeyQuestionsAgent(ctx, ac.Name, ac.Persona, model, embedder, store)

	// CRITICAL: Set the Bus so thoughts are emitted locally
	if base, ok := agent.(*entity.BaseEntity); ok {
		base.Bus = bus
	}

	// 2. Add a terminal logger for the Node
	logEvents := bus.Subscribe()
	go func() {
		for e := range logEvents {
			switch e.Type {
			case types.EventThought:
				fmt.Printf("[%s Thinking] %s\n", e.Actor, e.Message)
			case types.EventObservation:
				fmt.Printf("[%s Incoming] %s\n", e.Actor, e.Message)
			case types.EventAction:
				fmt.Printf("[%s Outgoing] %s\n", e.Actor, e.Message)
			}
		}
	}()

	fmt.Printf("Starting Pandemonium Agent Node [%s]...\n", *name)

	// 3. Wrap and host as A2A service
	ui.StartA2AServer(agent, bus, *port)
}
