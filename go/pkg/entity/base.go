package entity

import (
	"context"
	"fmt"
	"sync"

	"github.com/ghchinoy/pandemonium/pkg/component"
	"github.com/ghchinoy/pandemonium/pkg/types"
)

// BaseEntity implements the Entity interface and manages a set of components.
type BaseEntity struct {
	EntityName string
	Lock       sync.Mutex

	ActComponent component.ActingComponent
	Components   map[string]component.ContextComponent
	Bus          *types.EventBus
}

// Name returns the name of the entity.
func (e *BaseEntity) Name() string {
	return e.EntityName
}

// Act implements the Entity interface.
func (e *BaseEntity) Act(ctx context.Context, spec types.ActionSpec) (string, error) {
	e.Lock.Lock()
	defer e.Lock.Unlock()

	// 1. PRE_ACT
	contexts, err := e.runPhase(ctx, func(c component.ContextComponent) (string, error) {
		return c.PreAct(ctx, spec)
	})
	if err != nil {
		return "", err
	}

	if e.Bus != nil {
		e.Bus.Publish(types.Event{
			Type:    types.EventThought,
			Actor:   e.EntityName,
			Message: fmt.Sprintf("%s is considering their action...", e.EntityName),
			Data:    contexts,
		})
	}

	// 2. GET_ACTION
	action, err := e.ActComponent.GetActionAttempt(ctx, contexts, spec)
	if err != nil {
		return "", err
	}

	// 3. POST_ACT
	_, err = e.runPhase(ctx, func(c component.ContextComponent) (string, error) {
		return c.PostAct(ctx, action)
	})
	if err != nil {
		return "", err
	}

	// 4. UPDATE
	err = e.update(ctx)
	if err != nil {
		return "", err
	}

	return action, nil
}

// Observe implements the Entity interface.
func (e *BaseEntity) Observe(ctx context.Context, observation string) error {
	e.Lock.Lock()
	defer e.Lock.Unlock()

	// 1. PRE_OBSERVE
	_, err := e.runPhase(ctx, func(c component.ContextComponent) (string, error) {
		return c.PreObserve(ctx, observation)
	})
	if err != nil {
		return err
	}

	// 2. POST_OBSERVE
	_, err = e.runPhase(ctx, func(c component.ContextComponent) (string, error) {
		return c.PostObserve(ctx)
	})
	if err != nil {
		return err
	}

	// 3. UPDATE
	return e.update(ctx)
}

func (e *BaseEntity) runPhase(_ context.Context, call func(component.ContextComponent) (string, error)) (map[string]string, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	contexts := make(map[string]string)
	var firstErr error

	for name, comp := range e.Components {
		wg.Add(1)
		go func(n string, c component.ContextComponent) {
			defer wg.Done()
			val, err := call(c)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if firstErr == nil {
					firstErr = fmt.Errorf("component %q: %w", n, err)
				}
				return
			}
			contexts[n] = val
		}(name, comp)
	}

	wg.Wait()
	return contexts, firstErr
}

func (e *BaseEntity) update(ctx context.Context) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(e.Components))

	for _, c := range e.Components {
		wg.Add(1)
		go func(comp component.ContextComponent) {
			defer wg.Done()
			if err := comp.Update(ctx); err != nil {
				errChan <- err
			}
		}(c)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}
	return nil
}
