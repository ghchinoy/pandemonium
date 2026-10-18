package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/a2aproject/a2a-go/v2/a2a"

	"github.com/ghchinoy/pandemonium/pkg/types"
)

// Asynchronous implements Concordia v2.4's fully asynchronous simulation engine:
// each player entity runs its own independent observe-act loop concurrently in a
// dedicated goroutine at its own pace against a synchronized Game Master.
type Asynchronous struct {
	Bus        *types.EventBus
	MinSteps   int
	ContextID  string
	GMs        []types.Entity
	Controller *StepController
}

// Run executes the asynchronous per-player simulation loops until maxSteps per player
// or a Game Master termination signal is reached.
func (a *Asynchronous) Run(ctx context.Context, gm types.Entity, players []types.Entity, premise string, maxSteps int) error {
	if a.ContextID == "" {
		a.ContextID = a2a.NewContextID()
	}

	a.emit(types.Event{Type: types.EventSimStart, Message: "Asynchronous simulation initializing...", Data: premise})

	allEntities := append([]types.Entity{gm}, players...)
	for _, e := range allEntities {
		if a2aEnt, ok := e.(types.A2AEntity); ok {
			a2aEnt.SetContextID(a.ContextID)
		}
	}

	if premise != "" {
		a.emit(types.Event{Type: types.EventStepStart, Message: "Setup: Setting the Scene", Data: 0})
		if err := gm.Observe(ctx, fmt.Sprintf("[Premise] %s", premise)); err != nil {
			return fmt.Errorf("setup observation for GM: %w", err)
		}
		for _, p := range players {
			setupMsg := fmt.Sprintf("System: Scenario Initialization. Premise: %s", premise)
			if err := p.Observe(ctx, setupMsg); err != nil {
				return fmt.Errorf("setup observation for %q: %w", p.Name(), err)
			}
			a.emit(types.Event{Type: types.EventObservation, Actor: p.Name(), Message: "Received world context."})
		}
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var gmMu sync.Mutex
	activeGM := gm
	var globalSteps atomic.Int64
	var terminated atomic.Bool

	errChan := make(chan error, len(players))
	var wg sync.WaitGroup

	for _, p := range players {
		wg.Add(1)
		go func(player types.Entity) {
			defer wg.Done()
			for iter := 0; iter < maxSteps; iter++ {
				if runCtx.Err() != nil || terminated.Load() {
					return
				}
				if a.Controller != nil && !a.Controller.WaitForStepPermission(runCtx) {
					return
				}

				stepNum := int(globalSteps.Add(1))
				a.emit(types.Event{Type: types.EventStepStart, Actor: player.Name(), Message: fmt.Sprintf("Async Step %d (%s)", stepNum, player.Name()), Data: stepNum})

				// 1. Check termination & GM switching under GM lock
				gmMu.Lock()
				if stepNum > a.MinSteps && !terminated.Load() {
					terminateSpec := types.ActionSpec{
						CallToAction: "System: You are the Game Master. Is the simulation logically finished? Reply ONLY with 'Yes' or 'No'.",
						OutputType:   types.OutputTerminate,
						Options:      []string{"Yes", "No"},
					}
					res, err := activeGM.Act(runCtx, terminateSpec)
					if err != nil {
						gmMu.Unlock()
						if runCtx.Err() == nil {
							errChan <- fmt.Errorf("checking termination: %w", err)
							cancel()
						}
						return
					}
					if strings.HasPrefix(strings.ToLower(strings.TrimSpace(res)), "yes") {
						terminated.Store(true)
						a.emit(types.Event{Type: types.EventResolution, Message: "The Game Master has concluded the simulation."})
						gmMu.Unlock()
						cancel()
						return
					}
				}

				if len(a.GMs) > 1 {
					nextGM, err := SelectNextGM(runCtx, activeGM, a.GMs)
					if err != nil {
						gmMu.Unlock()
						if runCtx.Err() == nil {
							errChan <- err
							cancel()
						}
						return
					}
					activeGM = nextGM
				}

				obsSpec := types.ActionSpec{
					CallToAction: fmt.Sprintf("System: You are the Game Master. Describe what %s sees and hears right now.", player.Name()),
					OutputType:   types.OutputMakeObservation,
				}
				obs, err := activeGM.Act(runCtx, obsSpec)
				gmMu.Unlock()
				if err != nil {
					if runCtx.Err() == nil {
						errChan <- fmt.Errorf("generating observation for %q: %w", player.Name(), err)
						cancel()
					}
					return
				}

				// 2. Player observes and acts independently in its own goroutine
				if obs != "" {
					a.emit(types.Event{Type: types.EventObservation, Actor: player.Name(), Message: obs})
					if err := player.Observe(runCtx, obs); err != nil {
						if runCtx.Err() == nil {
							errChan <- fmt.Errorf("actor %q observation: %w", player.Name(), err)
							cancel()
						}
						return
					}
				}

				actionSpec := types.ActionSpec{
					CallToAction: fmt.Sprintf("What does %s do next?", player.Name()),
					OutputType:   types.OutputFree,
				}
				action, err := player.Act(runCtx, actionSpec)
				if err != nil {
					if runCtx.Err() == nil {
						errChan <- fmt.Errorf("actor %q action: %w", player.Name(), err)
						cancel()
					}
					return
				}
				a.emit(types.Event{Type: types.EventAction, Actor: player.Name(), Message: action})

				// 3. Synchronize with GM to resolve action and update world memory
				gmMu.Lock()
				if terminated.Load() {
					gmMu.Unlock()
					return
				}
				resolveSpec := types.ActionSpec{
					CallToAction: fmt.Sprintf("System: You are the Game Master. %s tried to: %s. Resolve this action into the physical and environmental outcome.", player.Name(), action),
					OutputType:   types.OutputResolve,
				}
				resolution, err := activeGM.Act(runCtx, resolveSpec)
				if err != nil {
					gmMu.Unlock()
					if runCtx.Err() == nil {
						errChan <- fmt.Errorf("resolving action for %q: %w", player.Name(), err)
						cancel()
					}
					return
				}
				a.emit(types.Event{Type: types.EventResolution, Actor: player.Name(), Message: resolution})

				if err := activeGM.Observe(runCtx, fmt.Sprintf("[Event] %s", resolution)); err != nil {
					gmMu.Unlock()
					if runCtx.Err() == nil {
						errChan <- fmt.Errorf("updating GM world state: %w", err)
						cancel()
					}
					return
				}
				gmMu.Unlock()
			}
		}(p)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	a.emit(types.Event{Type: types.EventSimComplete, Message: "Simulation completed"})
	return nil
}

func (a *Asynchronous) emit(e types.Event) {
	if a.Bus != nil {
		a.Bus.Publish(e)
	}
}
