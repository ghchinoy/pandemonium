// Package ui implements the terminal and web interfaces for Pandemonium.
package ui

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/a2aproject/a2a-go/a2a"
	"github.com/a2aproject/a2a-go/a2asrv"
	"github.com/a2aproject/a2a-go/a2asrv/eventqueue"
	"github.com/ghchinoy/pandemonium/pkg/types"
)

// EntityExecutor wraps a Pandemonium Entity to fulfill the A2A AgentExecutor interface.
type EntityExecutor struct {
	Entity types.Entity
	Bus    *types.EventBus
}

// Execute handles incoming A2A requests by delegating to the Pandemonium Entity.
func (e *EntityExecutor) Execute(ctx context.Context, reqCtx *a2asrv.RequestContext, q eventqueue.Queue) error {
	var input string
	if reqCtx.Message != nil {
		for _, part := range reqCtx.Message.Parts {
			if tp, ok := part.(a2a.TextPart); ok {
				input = tp.Text
				break
			}
		}
	}

	// Determine if this is an Observation or an Action request.
	// In Pandemonium, an Observation is "World -> Agent" (User role in A2A)
	// An Action is "GM -> Agent: What do you do?"
	
	// Heuristic: If it looks like a question/prompt, it's an Act call.
	// Otherwise, it's an Observation.
	// A more robust way would be using A2A Metadata or specific Skills.
	
	isActionRequest := reqCtx.Message.Role == a2a.MessageRoleUser

	if isActionRequest {
		// 1. Mark task as working
		if err := q.Write(ctx, a2a.NewStatusUpdateEvent(reqCtx, a2a.TaskStateWorking, nil)); err != nil {
			log.Printf("Failed to write working status: %v", err)
		}

		// 2. Subscribe to the local EventBus BEFORE acting
		var reasoningGroup sync.WaitGroup
		reasoning := make(map[string]string)
		var mu sync.Mutex

		var action string
		var err error

		if e.Bus != nil {
			sub := e.Bus.Subscribe()
			done := make(chan struct{})
			reasoningGroup.Add(1)
			go func() {
				defer reasoningGroup.Done()
				for {
					select {
					case <-done:
						return
					case event, ok := <-sub:
						if !ok {
							return
						}
						if event.Actor == e.Entity.Name() && event.Type == types.EventThought {
							if contexts, ok := event.Data.(map[string]string); ok {
								mu.Lock()
								for name, val := range contexts {
									reasoning[name] = val
									art := &a2a.Artifact{
										ID:    a2a.NewArtifactID(),
										Name:  name,
										Parts: []a2a.Part{a2a.TextPart{Text: val}},
									}
									// Only write if the context is still active
									select {
									case <-ctx.Done():
										mu.Unlock()
										return
									default:
										if err := q.Write(ctx, &a2a.TaskArtifactUpdateEvent{
											ContextID: reqCtx.ContextID,
											TaskID:    reqCtx.TaskID,
											Artifact:  art,
										}); err != nil {
											log.Printf("Failed to write artifact: %v", err)
										}
									}
								}
								mu.Unlock()
							}
						}
					}
				}
			}()

			// 3. Call Pandemonium Entity Act
			spec := types.ActionSpec{
				CallToAction: input,
				OutputType:   types.OutputFree,
			}
			action, err = e.Entity.Act(ctx, spec)
			close(done) // Signal goroutine to stop
			reasoningGroup.Wait()
		} else {
			// No bus, just act
			spec := types.ActionSpec{
				CallToAction: input,
				OutputType:   types.OutputFree,
			}
			action, err = e.Entity.Act(ctx, spec)
		}

		if err != nil {
			return q.Write(ctx, a2a.NewStatusUpdateEvent(reqCtx, a2a.TaskStateFailed, 
				a2a.NewMessageForTask(a2a.MessageRoleAgent, reqCtx, a2a.TextPart{Text: err.Error()})))
		}

		// 4. Return final result
		response := a2a.NewMessageForTask(a2a.MessageRoleAgent, reqCtx, a2a.TextPart{Text: action})
		return q.Write(ctx, response)
	}

	// It's an observation
	err := e.Entity.Observe(ctx, input)
	if err != nil {
		return err
	}
	
	return q.Write(ctx, a2a.NewMessageForTask(a2a.MessageRoleAgent, reqCtx, a2a.TextPart{Text: "Observation received."}))
}

// Cancel is a no-op for EntityExecutor.
func (e *EntityExecutor) Cancel(_ context.Context, _ *a2asrv.RequestContext, q eventqueue.Queue) error {
	return nil
}

// StartA2AServer hosts a Pandemonium Entity as an A2A service.
func StartA2AServer(entity types.Entity, bus *types.EventBus, port int) {
	mux := http.NewServeMux()
	
	executor := &EntityExecutor{Entity: entity, Bus: bus}
	handler := a2asrv.NewHandler(executor)
	
mux.Handle("/invoke", a2asrv.NewJSONRPCHandler(handler))
	
	card := &a2a.AgentCard{
		Name:        entity.Name(),
		Description: fmt.Sprintf("Pandemonium Agent: %s", entity.Name()),
		URL:         fmt.Sprintf("http://localhost:%d/invoke", port),
		PreferredTransport: a2a.TransportProtocolJSONRPC,
		Capabilities: a2a.AgentCapabilities{Streaming: true},
	}
	mux.Handle(a2asrv.WellKnownAgentCardPath, a2asrv.NewStaticAgentCardHandler(card))

	fmt.Printf("[A2A] Hosting agent %s at http://localhost:%d\n", entity.Name(), port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}
