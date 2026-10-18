package engine

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"sync"

	"github.com/ghchinoy/pandemonium/pkg/llm"
	"github.com/ghchinoy/pandemonium/pkg/prefab"
	"github.com/ghchinoy/pandemonium/pkg/store"
	"github.com/ghchinoy/pandemonium/pkg/types"
	"github.com/ghchinoy/pandemonium/pkg/vector"
)

// SessionInfo summarizes the runtime state of a managed simulation session.
type SessionInfo struct {
	ID           string `json:"id"`
	ScenarioName string `json:"scenario_name"`
	Running      bool   `json:"running"`
	Paused       bool   `json:"paused"`
	Stopped      bool   `json:"stopped"`
}

// SimulationManager manages concurrent simulation sessions, StepControllers, and StateStore persistence.
type SimulationManager struct {
	Store    store.StateStore
	Model    llm.Model
	Embedder vector.Embedder
	Bus      *types.EventBus

	mu       sync.RWMutex
	sessions map[string]*managedSession
}

type managedSession struct {
	id           string
	scenarioName string
	controller   *StepController
}

// NewManager initializes a SimulationManager.
func NewManager(stateStore store.StateStore, model llm.Model, embedder vector.Embedder, bus *types.EventBus) *SimulationManager {
	return &SimulationManager{
		Store:    stateStore,
		Model:    model,
		Embedder: embedder,
		Bus:      bus,
		sessions: make(map[string]*managedSession),
	}
}

// RegisterSession registers an existing StepController under the given session ID.
func (m *SimulationManager) RegisterSession(id, scenarioName string, ctrl *StepController) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[id] = &managedSession{
		id:           id,
		scenarioName: scenarioName,
		controller:   ctrl,
	}
}

// StartSession loads a scenario and launches a managed simulation session in the background.
func (m *SimulationManager) StartSession(ctx context.Context, id, scenarioPath string, startPaused bool) (*StepController, error) {
	bus := m.Bus
	if bus == nil {
		bus = types.NewEventBus()
	}

	simData, err := prefab.LoadScenario(ctx, scenarioPath, m.Model, m.Embedder, bus)
	if err != nil {
		return nil, fmt.Errorf("loading scenario %q for session %q: %w", scenarioPath, id, err)
	}

	ctrl := NewStepController(startPaused)
	m.RegisterSession(id, simData.Config.Name, ctrl)

	sim := New(simData.Config.Settings, bus, "")
	AttachStepController(sim, ctrl)

	if m.Store != nil {
		sub := bus.Subscribe()
		go func() {
			defer bus.Unsubscribe(sub)
			for ev := range sub {
				_ = m.Store.SaveEvent(ctx, id, ev)
				if ev.Type == types.EventSimComplete {
					return
				}
			}
		}()
	}

	go func() {
		err := sim.Run(ctx, simData.GM, simData.Players, simData.Config.Premise, simData.Config.Settings.MaxSteps)
		if err != nil {
			bus.Publish(types.Event{Type: types.EventSimComplete, Message: fmt.Sprintf("Error: %v", err)})
		}
		ctrl.Stop()
	}()

	return ctrl, nil
}

// Controller returns the StepController for a session ID (or "default" if id is empty).
func (m *SimulationManager) Controller(id string) (*StepController, bool) {
	if id == "" {
		id = "default"
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	sess, ok := m.sessions[id]
	if !ok || sess.controller == nil {
		return nil, false
	}
	return sess.controller, true
}

// ListSessions returns snapshots of all registered sessions.
func (m *SimulationManager) ListSessions() []SessionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []SessionInfo
	for _, s := range m.sessions {
		var running, paused, stopped bool
		if s.controller != nil {
			running = s.controller.IsRunning()
			paused = s.controller.IsPaused()
			stopped = s.controller.ShouldStop()
		}
		list = append(list, SessionInfo{
			ID:           s.id,
			ScenarioName: s.scenarioName,
			Running:      running,
			Paused:       paused,
			Stopped:      stopped,
		})
	}
	slices.SortFunc(list, func(a, b SessionInfo) int {
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})
	return list
}

// ListScenarios returns the union of built-in scenario templates and StateStore scenarios.
func (m *SimulationManager) ListScenarios(ctx context.Context) ([]string, error) {
	seen := make(map[string]bool)
	var names []string

	entries, err := fs.ReadDir(prefab.BuiltinScenarios, "scenarios")
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() && path.Ext(e.Name()) == ".yaml" {
				seen[e.Name()] = true
				names = append(names, e.Name())
			}
		}
	}

	if m.Store != nil {
		stored, err := m.Store.ListScenarios(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing stored scenarios: %w", err)
		}
		for _, s := range stored {
			if !seen[s] {
				seen[s] = true
				names = append(names, s)
			}
		}
	}

	slices.Sort(names)
	return names, nil
}

// AttachStepController attaches ctrl to any supported Engine implementation.
func AttachStepController(eng Engine, ctrl *StepController) {
	switch e := eng.(type) {
	case *Sequential:
		e.Controller = ctrl
	case *Simultaneous:
		e.Controller = ctrl
	case *Asynchronous:
		e.Controller = ctrl
	}
}
