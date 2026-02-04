// Package main implements the standalone agent node for distributed simulations.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/ghchinoy/pandemonium/pkg/entity"
	"github.com/ghchinoy/pandemonium/pkg/llm"
	"github.com/ghchinoy/pandemonium/pkg/prefab"
	"github.com/ghchinoy/pandemonium/pkg/types"
	"github.com/ghchinoy/pandemonium/pkg/ui"
	"github.com/joho/godotenv"
)

// MockEmbedder implements a simple identity embedder.
type MockEmbedder struct{}

func (m *MockEmbedder) Embed(_ context.Context, text string) ([]float64, error) {
	return []float64{float64(len(text)), 0, 0}, nil
}

func main() {
	port := flag.Int("port", 9001, "Port to host the A2A agent")
	name := flag.String("name", "Bob", "Name of the agent")
	persona := flag.String("persona", "A skeptical journalist investigating the storm.", "Agent persona")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	ctx := context.Background()

	model, err := llm.NewGeminiModel(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize model: %v", err)
	}
	embedder, err := llm.NewGeminiEmbedder(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize embedder: %v", err)
	}

		// 1. Initialize the local agent

		bus := types.NewEventBus()

		agent := prefab.NewThreeKeyQuestionsAgent(ctx, *name, *persona, model, embedder, "")

		

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
