// Package engine implements simulation loops and orchestration logic.
package engine

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/a2aproject/a2a-go/v2/a2a"

	"github.com/ghchinoy/pandemonium/pkg/types"
)

// Sequential implements a turn-based simulation loop.
type Sequential struct {
	Bus       *types.EventBus
	LastActor string
	MinSteps  int
	ContextID string
	Strategy  types.TurnTakingStrategy

	lastRotationIndex int
}

// Run executes the simulation loop.
func (s *Sequential) Run(ctx context.Context, gm types.Entity, players []types.Entity, premise string, maxSteps int) error {
	// Generate a unique session ID for A2A correlation
	s.ContextID = a2a.NewContextID()

	fmt.Printf("\n--- Simulation Start ---\n")
	s.emit(types.Event{Type: types.EventSimStart, Message: "Simulation initializing...", Data: premise})

	// 1. Initial Handshake & Context Sync
	allEntities := append([]types.Entity{gm}, players...)
	for _, e := range allEntities {
		if a2aEnt, ok := e.(types.A2AEntity); ok {
			a2aEnt.SetContextID(s.ContextID)
		}
	}

	// 2. Scenario Setup Phase (Step 0)
	// Mirroring Concordia's INITIALIZER: Seed agent memories with the premise.
	fmt.Printf("[World] Setting the Scene: %s\n", premise)
	s.emit(types.Event{Type: types.EventStepStart, Message: "Setup: Setting the Scene", Data: 0})

	if premise != "" {
		if err := gm.Observe(ctx, fmt.Sprintf("[Premise] %s", premise)); err != nil {
			return fmt.Errorf("setup observation for GM: %w", err)
		}
		// Synchronously seed all players with the world context
		for _, p := range players {
			setupMsg := fmt.Sprintf("System: Scenario Initialization. Premise: %s", premise)
			if err := p.Observe(ctx, setupMsg); err != nil {
				return fmt.Errorf("setup observation for %q: %w", p.Name(), err)
			}
			s.emit(types.Event{Type: types.EventObservation, Actor: p.Name(), Message: "Received world context."})
		}
	}

	// 3. Main Simulation Loop
	for step := 0; step < maxSteps; step++ {
		fmt.Printf("\n--- Step %d ---\n", step+1)
		s.emit(types.Event{Type: types.EventStepStart, Message: fmt.Sprintf("Step %d", step+1), Data: step + 1})

		// Sync step ID with all A2A-capable entities
		for _, e := range allEntities {
			if a2aEnt, ok := e.(types.A2AEntity); ok {
				a2aEnt.SetMetadata("step_id", step+1)
			}
		}

		// 2. Check Termination (only after MinSteps)
		if step >= s.MinSteps {
			terminateSpec := types.ActionSpec{
				CallToAction: "System: You are the Game Master. The objective of this simulation is to explore the social dynamics of the situation. Is the simulation logically finished? Have the agents reached a stable state or resolution? Reply ONLY with 'Yes' or 'No'.",
				OutputType:   types.OutputTerminate,
				Options:      []string{"Yes", "No"},
			}
			res, err := gm.Act(ctx, terminateSpec)
			if err != nil {
				return fmt.Errorf("checking termination: %w", err)
			}
			cleanRes := strings.TrimSpace(strings.ToLower(res))
			s.emit(types.Event{Type: types.EventThought, Actor: "GM", Message: fmt.Sprintf("Is simulation finished? GM says: %s", res)})

			if strings.HasPrefix(cleanRes, "yes") {
				fmt.Println("[World] Termination signal received.")
				s.emit(types.Event{Type: types.EventResolution, Message: "The Game Master has concluded the simulation."})
				break
			}
		}

		// 3. Select Next Actor
		var nextActorName string
		playerNames := make([]string, len(players))
		playerMap := make(map[string]types.Entity)
		for i, p := range players {
			playerNames[i] = p.Name()
			playerMap[p.Name()] = p
		}

		switch s.Strategy {
		case types.StrategyRotation:
			nextActorName = playerNames[s.lastRotationIndex%len(players)]
			s.lastRotationIndex++
			s.emit(types.Event{Type: types.EventThought, Actor: "Engine", Message: fmt.Sprintf("Strategy [Rotation]: selected %s", nextActorName)})

		case types.StrategyRandom:
			nextActorName = playerNames[rand.Intn(len(players))]
			s.emit(types.Event{Type: types.EventThought, Actor: "Engine", Message: fmt.Sprintf("Strategy [Random]: selected %s", nextActorName)})

		default:
			// Default: StrategySequential (GM Decided)
			actingSpec := types.ActionSpec{
				CallToAction: fmt.Sprintf("System: You are the Game Master. Available players: %s. The last player to act was: %s. Based on the narrative so far, who should act next? Please ensure all players get a turn. Reply ONLY with the name of the player.", strings.Join(playerNames, ", "), s.LastActor),
				OutputType:   types.OutputNextActing,
				Options:      playerNames,
			}
			var err error
			nextActorName, err = gm.Act(ctx, actingSpec)
			if err != nil {
				return fmt.Errorf("selecting next actor: %w", err)
			}
			nextActorName = strings.TrimSpace(nextActorName)
			s.emit(types.Event{Type: types.EventThought, Actor: "GM", Message: fmt.Sprintf("Strategy [Sequential/GM]: GM chose %s (last was %s)", nextActorName, s.LastActor)})
		}

		s.LastActor = nextActorName
		fmt.Printf("[GM] Next to act: %s\n", nextActorName)

		nextActor, ok := playerMap[nextActorName]
		if !ok {
			// Fallback: If GM gives a long response, try to find the name in it
			found := false
			for _, name := range playerNames {
				if strings.Contains(nextActorName, name) {
					nextActorName = name
					nextActor = playerMap[name]
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("GM selected unknown actor %q", nextActorName)
			}
		}

		// 4. Generate Observation for Actor
		obsSpec := types.ActionSpec{
			CallToAction: fmt.Sprintf("System: You are the Game Master. Describe the physical environment and external events that %s sees and hears right now. CRITICAL: Do NOT describe %s's own thoughts, feelings, or internal reactions. Do NOT describe %s performing any actions.", nextActorName, nextActorName, nextActorName),
			OutputType:   types.OutputMakeObservation,
		}
		observation, err := gm.Act(ctx, obsSpec)
		if err != nil {
			return fmt.Errorf("generating observation: %w", err)
		}

		if observation != "" {
			fmt.Printf("[%s] Observed: %s\n", nextActorName, observation)
			s.emit(types.Event{Type: types.EventObservation, Actor: nextActorName, Message: observation})
			if err := nextActor.Observe(ctx, observation); err != nil {
				return fmt.Errorf("actor observation: %w", err)
			}
		}

		// 5. Get Actor Action
		actionSpec := types.ActionSpec{
			CallToAction: fmt.Sprintf("What does %s do next?", nextActorName),
			OutputType:   types.OutputFree,
		}
		action, err := nextActor.Act(ctx, actionSpec)
		if err != nil {
			return fmt.Errorf("actor action: %w", err)
		}
		fmt.Printf("[%s] Action: %s\n", nextActorName, action)
		s.emit(types.Event{Type: types.EventAction, Actor: nextActorName, Message: action})

		// 6. Resolve Action
		resolveSpec := types.ActionSpec{
			CallToAction: fmt.Sprintf("System: You are the Game Master. %s tried to: %s. Resolve this action by describing the physical and environmental outcome. CRITICAL: Do NOT write dialogue or reactions for ANY player. Do NOT describe what other players do in response. Only describe the change in the world.", nextActorName, action),
			OutputType:   types.OutputResolve,
		}
		resolution, err := gm.Act(ctx, resolveSpec)
		if err != nil {
			return fmt.Errorf("resolving action: %w", err)
		}
		fmt.Printf("[World] Resolution: %s\n", resolution)
		s.emit(types.Event{Type: types.EventResolution, Message: resolution})

		// 7. Inform GM of resolution (to update its world state)
		if err := gm.Observe(ctx, fmt.Sprintf("[Event] %s", resolution)); err != nil {
			return fmt.Errorf("updating GM world state: %w", err)
		}
	}

	fmt.Printf("\n--- Simulation Complete ---\n")
	s.emit(types.Event{Type: types.EventSimComplete, Message: "Simulation completed"})
	return nil
}

func (s *Sequential) emit(e types.Event) {
	if s.Bus != nil {
		s.Bus.Publish(e)
	}
}
