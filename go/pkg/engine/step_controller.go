package engine

import (
	"context"
	"sync"
)

// StepController provides thread-safe play, pause, single-step, and stop control
// over simulation execution, modeled after Concordia v2.4's StepController.
type StepController struct {
	mu            sync.Mutex
	cond          *sync.Cond
	running       bool
	stepRequested bool
	stopRequested bool
}

// NewStepController initializes a StepController. If startPaused is true,
// the simulation waits before Step 1 until Play() or Step() is called.
func NewStepController(startPaused bool) *StepController {
	sc := &StepController{
		running: !startPaused,
	}
	sc.cond = sync.NewCond(&sc.mu)
	return sc
}

// Play resumes continuous simulation execution.
func (sc *StepController) Play() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.running = true
	sc.stepRequested = false
	sc.cond.Broadcast()
}

// Pause pauses simulation execution at the next step boundary.
func (sc *StepController) Pause() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.running = false
	sc.stepRequested = false
}

// Step advances a paused simulation by a single step and returns to the paused state.
func (sc *StepController) Step() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	if sc.running {
		return
	}
	sc.stepRequested = true
	sc.cond.Broadcast()
}

// Stop requests the simulation to stop at the next step boundary.
func (sc *StepController) Stop() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.stopRequested = true
	sc.running = true
	sc.cond.Broadcast()
}

// IsRunning reports whether the controller is in continuous running mode.
func (sc *StepController) IsRunning() bool {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.running && !sc.stopRequested
}

// IsPaused reports whether the controller is currently paused.
func (sc *StepController) IsPaused() bool {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return !sc.running && !sc.stopRequested
}

// ShouldStop reports whether Stop() has been requested.
func (sc *StepController) ShouldStop() bool {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.stopRequested
}

// WaitForStepPermission blocks until the simulation is permitted to execute the next step.
// Returns false if Stop() was requested or ctx was canceled.
func (sc *StepController) WaitForStepPermission(ctx context.Context) bool {
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			sc.mu.Lock()
			sc.cond.Broadcast()
			sc.mu.Unlock()
		case <-done:
		}
	}()

	sc.mu.Lock()
	defer sc.mu.Unlock()

	for !sc.running && !sc.stepRequested && !sc.stopRequested && ctx.Err() == nil {
		sc.cond.Wait()
	}

	if sc.stopRequested || ctx.Err() != nil {
		return false
	}

	if sc.stepRequested {
		sc.stepRequested = false
	}

	return true
}
