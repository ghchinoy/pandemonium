// Package main implements the Pandemonium simulation orchestrator.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/ghchinoy/pandemonium/pkg/engine"
	"github.com/ghchinoy/pandemonium/pkg/llm"
	"github.com/ghchinoy/pandemonium/pkg/prefab"
	"github.com/ghchinoy/pandemonium/pkg/types"
	"github.com/ghchinoy/pandemonium/pkg/ui"
)

func main() {
	logFile := flag.String("log", "", "path to file for simulation event logging (JSONL format)")
	scenarioPath := flag.String("scenario", "pub.yaml", "path to scenario YAML file or builtin name")
	webPath := flag.String("web", "../web/dist", "path to web frontend directory")
	flag.Parse()

	// Load environment variables from .env if present
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	ctx := context.Background()
	model, err := llm.NewGeminiModel(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize Gemini model: %v", err)
	}
	embedder, err := llm.NewGeminiEmbedder(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize Gemini embedder: %v", err)
	}

	// 1. Initialize Event Bus
	bus := types.NewEventBus()

	// 2. Load Scenario
	simData, err := prefab.LoadScenario(ctx, *scenarioPath, model, embedder, bus)
	if err != nil {
		log.Fatalf("Failed to load scenario: %v", err)
	}

	// 3. Initialize Engine
	sim := &engine.Sequential{
		Bus:      bus,
		MinSteps: simData.Config.Settings.MinSteps,
		Strategy: simData.Config.Settings.Strategy,
	}

	// 4. Set up subscribers
	tuiEvents := bus.Subscribe()
	webEvents := bus.Subscribe()

	if *logFile != "" {
		fileEvents := bus.Subscribe()
		go func(path string, ch chan types.Event) {
			f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Printf("Failed to open log file: %v", err)
				return
			}
			defer func() {
				if err := f.Close(); err != nil {
					log.Printf("Error closing log file: %v", err)
				}
			}()
			encoder := json.NewEncoder(f)
			for e := range ch {
				if err := encoder.Encode(e); err != nil {
					log.Printf("Failed to log event: %v", err)
				}
			}
		}(*logFile, fileEvents)
	}

	// 4. Start X-Ray Server
	server := ui.NewServer(webEvents, *webPath)
	go server.Start(8080)

	// 5. Register Agents in UI immediately
	for _, p := range simData.Config.Agents {
		isRemote := p.URL != ""
		bus.Publish(types.Event{
			Type:    types.EventObservation,
			Actor:   p.Name,
			Message: "Registered",
			Data:    map[string]bool{"isRemote": isRemote},
		})
	}

	// 6. Run Simulation in a goroutine
	go func() {
		err := sim.Run(ctx, simData.GM, simData.Players, simData.Config.Premise, simData.Config.Settings.MaxSteps)
		if err != nil {
			bus.Publish(types.Event{Type: types.EventSimComplete, Message: fmt.Sprintf("Error: %v", err)})
		}
	}()

	// 7. Start TUI
	p := ui.NewTUI(tuiEvents)
	if _, err := p.Run(); err != nil {
		log.Fatalf("TUI error: %v", err)
	}
}
