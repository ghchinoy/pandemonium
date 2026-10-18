package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/a2aproject/a2a-go/v2/a2a"

	"github.com/ghchinoy/pandemonium/pkg/types"
)

// Simultaneous implements a synchronous multi-actor simulation loop where all
// active players observe and act concurrently within each step before the Game Master
// resolves the aggregated actions into a single world outcome.
type Simultaneous struct {
	Bus            *types.EventBus
	MinSteps       int
	StartStep      int
	ContextID      string
	CheckpointPath string
	GMs            []types.Entity
	Controller     *StepController
}

// Run executes the simultaneous simulation loop.
func (s *Simultaneous) Run(ctx context.Context, gm types.Entity, players []types.Entity, premise string, maxSteps int) error {
	if s.ContextID == "" {
		s.ContextID = a2a.NewContextID()
	}

	s.emit(types.Event{Type: types.EventSimStart, Message: "Simultaneous simulation initializing...", Data: premise})

	allEntities := append([]types.Entity{gm}, players...)
	for _, e := range allEntities {
		if a2aEnt, ok := e.(types.A2AEntity); ok {
			a2aEnt.SetContextID(s.ContextID)
		}
	}

	if s.StartStep == 0 && premise != "" {
		s.emit(types.Event{Type: types.EventStepStart, Message: "Setup: Setting the Scene", Data: 0})
		if err := gm.Observe(ctx, fmt.Sprintf("[Premise] %s", premise)); err != nil {
			return fmt.Errorf("setup observation for GM: %w", err)
		}
		for _, p := range players {
			setupMsg := fmt.Sprintf("System: Scenario Initialization. Premise: %s", premise)
			if err := p.Observe(ctx, setupMsg); err != nil {
				return fmt.Errorf("setup observation for %q: %w", p.Name(), err)
			}
			s.emit(types.Event{Type: types.EventObservation, Actor: p.Name(), Message: "Received world context."})
		}
	}

	for step := s.StartStep; step < maxSteps; step++ {
		if s.Controller != nil && !s.Controller.WaitForStepPermission(ctx) {
			break
		}
		s.emit(types.Event{Type: types.EventStepStart, Message: fmt.Sprintf("Step %d", step+1), Data: step + 1})

		for _, e := range allEntities {
			if a2aEnt, ok := e.(types.A2AEntity); ok {
				a2aEnt.SetMetadata("step_id", step+1)
			}
		}

		// 1. Check Termination (only after MinSteps)
		if step >= s.MinSteps {
			terminateSpec := types.ActionSpec{
				CallToAction: "System: You are the Game Master. Is the simulation logically finished? Reply ONLY with 'Yes' or 'No'.",
				OutputType:   types.OutputTerminate,
				Options:      []string{"Yes", "No"},
			}
			res, err := gm.Act(ctx, terminateSpec)
			if err != nil {
				return fmt.Errorf("checking termination: %w", err)
			}
			cleanRes := strings.TrimSpace(strings.ToLower(res))
			s.emit(types.Event{Type: types.EventThought, Actor: gm.Name(), Message: fmt.Sprintf("Is simulation finished? GM says: %s", res)})
			if strings.HasPrefix(cleanRes, "yes") {
				s.emit(types.Event{Type: types.EventResolution, Message: "The Game Master has concluded the simulation."})
				break
			}
		}

		// 2. Switch Game Master if multiple rule sets are configured
		if len(s.GMs) > 1 {
			nextGM, err := SelectNextGM(ctx, gm, s.GMs)
			if err != nil {
				return err
			}
			if nextGM.Name() != gm.Name() {
				s.emit(types.Event{Type: types.EventThought, Actor: "Engine", Message: fmt.Sprintf("Switched Game Master from %s to %s", gm.Name(), nextGM.Name())})
				gm = nextGM
			}
		}

		// 3. Generate observations for each player and collect actions concurrently
		observations := make(map[string]string, len(players))
		for _, p := range players {
			obsSpec := types.ActionSpec{
				CallToAction: fmt.Sprintf("System: You are the Game Master. Describe what %s sees and hears right now.", p.Name()),
				OutputType:   types.OutputMakeObservation,
			}
			obs, err := gm.Act(ctx, obsSpec)
			if err != nil {
				return fmt.Errorf("generating observation for %q: %w", p.Name(), err)
			}
			observations[p.Name()] = obs
		}

		type playerAction struct {
			name   string
			action string
			err    error
		}
		results := make([]playerAction, len(players))
		var wg sync.WaitGroup

		for i, p := range players {
			wg.Add(1)
			go func(idx int, player types.Entity, obs string) {
				defer wg.Done()
				if obs != "" {
					s.emit(types.Event{Type: types.EventObservation, Actor: player.Name(), Message: obs})
					if err := player.Observe(ctx, obs); err != nil {
						results[idx] = playerAction{name: player.Name(), err: fmt.Errorf("actor %q observation: %w", player.Name(), err)}
						return
					}
				}
				actionSpec := types.ActionSpec{
					CallToAction: fmt.Sprintf("What does %s do next?", player.Name()),
					OutputType:   types.OutputFree,
				}
				act, err := player.Act(ctx, actionSpec)
				if err != nil {
					results[idx] = playerAction{name: player.Name(), err: fmt.Errorf("actor %q action: %w", player.Name(), err)}
					return
				}
				s.emit(types.Event{Type: types.EventAction, Actor: player.Name(), Message: act})
				results[idx] = playerAction{name: player.Name(), action: act}
			}(i, p, observations[p.Name()])
		}
		wg.Wait()

		var putativeEvents []string
		for _, r := range results {
			if r.err != nil {
				return r.err
			}
			putativeEvents = append(putativeEvents, fmt.Sprintf("- %s: %s", r.name, r.action))
		}

		// 4. Resolve aggregated actions
		resolveSpec := types.ActionSpec{
			CallToAction: fmt.Sprintf(
				"System: You are the Game Master. The players simultaneously attempted the following actions:\n%s\nResolve all of these actions into a single coherent environmental outcome.",
				strings.Join(putativeEvents, "\n"),
			),
			OutputType: types.OutputResolve,
		}
		resolution, err := gm.Act(ctx, resolveSpec)
		if err != nil {
			return fmt.Errorf("resolving simultaneous actions: %w", err)
		}
		s.emit(types.Event{Type: types.EventResolution, Message: resolution})

		if err := gm.Observe(ctx, fmt.Sprintf("[Event] %s", resolution)); err != nil {
			return fmt.Errorf("updating GM world state: %w", err)
		}

		if s.CheckpointPath != "" {
			seq := &Sequential{ContextID: s.ContextID}
			cp := seq.Snapshot(step+1, gm, players)
			if err := SaveCheckpoint(s.CheckpointPath, cp); err != nil {
				return fmt.Errorf("saving checkpoint at step %d: %w", step+1, err)
			}
		}
	}

	s.emit(types.Event{Type: types.EventSimComplete, Message: "Simulation completed"})
	return nil
}

func (s *Simultaneous) emit(e types.Event) {
	if s.Bus != nil {
		s.Bus.Publish(e)
	}
}
