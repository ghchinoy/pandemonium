// Package ui implements the terminal and web interfaces for Pandemonium.
package ui

import (
	"context"
	"fmt"
	"iter"
	"log"
	"net/http"
	"sync"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/a2aproject/a2a-go/v2/a2asrv"
	"github.com/ghchinoy/pandemonium/pkg/types"
)

// EntityExecutor wraps a Pandemonium Entity to fulfill the A2A AgentExecutor interface.
type EntityExecutor struct {
	Entity types.Entity
	Bus    *types.EventBus
}

// Execute handles incoming A2A requests by delegating to the Pandemonium Entity.
func (e *EntityExecutor) Execute(ctx context.Context, execCtx *a2asrv.ExecutorContext) iter.Seq2[a2a.Event, error] {
	return func(yield func(a2a.Event, error) bool) {
		if execCtx.Message == nil {
			return
		}

		var input string
		for _, part := range execCtx.Message.Parts {
			if text := part.Text(); text != "" {
				input = text
				break
			}
		}

		isObservation := false
		if execCtx.Metadata != nil {
			if t, ok := execCtx.Metadata["type"].(string); ok && t == "observation" {
				isObservation = true
			}
		}

		if isObservation {
			if execCtx.StoredTask == nil {
				if !yield(a2a.NewSubmittedTask(execCtx, execCtx.Message), nil) {
					return
				}
			}
			err := e.Entity.Observe(ctx, input)
			if err != nil {
				errMsg := a2a.NewMessageForTask(a2a.MessageRoleAgent, execCtx, a2a.NewTextPart(err.Error()))
				yield(a2a.NewStatusUpdateEvent(execCtx, a2a.TaskStateFailed, errMsg), nil)
				return
			}

			resp := a2a.NewMessageForTask(a2a.MessageRoleAgent, execCtx, a2a.NewTextPart("Observation received."))
			yield(a2a.NewStatusUpdateEvent(execCtx, a2a.TaskStateCompleted, resp), nil)
			return
		}

		// Action Request
		if execCtx.StoredTask == nil {
			if !yield(a2a.NewSubmittedTask(execCtx, execCtx.Message), nil) {
				return
			}
		}

		if !yield(a2a.NewStatusUpdateEvent(execCtx, a2a.TaskStateWorking, nil), nil) {
			return
		}

		eventsChan := make(chan a2a.Event, 100)

		if e.Bus != nil {
			sub := e.Bus.Subscribe()
			done := make(chan struct{})
			var reasoningGroup sync.WaitGroup
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
								for name, val := range contexts {
									if val == "" {
										continue
									}
									artEvent := a2a.NewArtifactEvent(execCtx, a2a.NewTextPart(val))
									if artEvent.Artifact != nil {
										artEvent.Artifact.Name = name
									}
									select {
									case <-ctx.Done():
										return
									case eventsChan <- artEvent:
									}
								}
							}
						}
					}
				}
			}()

			type actResult struct {
				action string
				err    error
			}
			actDone := make(chan actResult, 1)

			go func() {
				spec := types.ActionSpec{
					CallToAction: input,
					OutputType:   types.OutputFree,
				}
				act, err := e.Entity.Act(ctx, spec)
				actDone <- actResult{action: act, err: err}
			}()

			var action string
			var actErr error
		loop:
			for {
				select {
				case artEvent := <-eventsChan:
					if !yield(artEvent, nil) {
						close(done)
						return
					}
				case res := <-actDone:
					action = res.action
					actErr = res.err
					close(done)
					break loop
				case <-ctx.Done():
					close(done)
					yield(a2a.NewStatusUpdateEvent(execCtx, a2a.TaskStateCanceled, nil), nil)
					return
				}
			}
			reasoningGroup.Wait()

			close(eventsChan)
			for artEvent := range eventsChan {
				if !yield(artEvent, nil) {
					return
				}
			}

			if actErr != nil {
				errMsg := a2a.NewMessageForTask(a2a.MessageRoleAgent, execCtx, a2a.NewTextPart(actErr.Error()))
				yield(a2a.NewStatusUpdateEvent(execCtx, a2a.TaskStateFailed, errMsg), nil)
				return
			}

			responseMsg := a2a.NewMessageForTask(a2a.MessageRoleAgent, execCtx, a2a.NewTextPart(action))
			yield(a2a.NewStatusUpdateEvent(execCtx, a2a.TaskStateCompleted, responseMsg), nil)
			return
		} else {
			spec := types.ActionSpec{
				CallToAction: input,
				OutputType:   types.OutputFree,
			}
			action, err := e.Entity.Act(ctx, spec)
			if err != nil {
				errMsg := a2a.NewMessageForTask(a2a.MessageRoleAgent, execCtx, a2a.NewTextPart(err.Error()))
				yield(a2a.NewStatusUpdateEvent(execCtx, a2a.TaskStateFailed, errMsg), nil)
				return
			}

			responseMsg := a2a.NewMessageForTask(a2a.MessageRoleAgent, execCtx, a2a.NewTextPart(action))
			yield(a2a.NewStatusUpdateEvent(execCtx, a2a.TaskStateCompleted, responseMsg), nil)
			return
		}
	}
}

// Cancel handles cancellation requests for EntityExecutor.
func (e *EntityExecutor) Cancel(_ context.Context, execCtx *a2asrv.ExecutorContext) iter.Seq2[a2a.Event, error] {
	return func(yield func(a2a.Event, error) bool) {
		yield(a2a.NewStatusUpdateEvent(execCtx, a2a.TaskStateCanceled, nil), nil)
	}
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
		SupportedInterfaces: []*a2a.AgentInterface{
			a2a.NewAgentInterface(fmt.Sprintf("http://localhost:%d/invoke", port), a2a.TransportProtocolJSONRPC),
		},
		Capabilities: a2a.AgentCapabilities{Streaming: true},
	}
	mux.Handle(a2asrv.WellKnownAgentCardPath, a2asrv.NewStaticAgentCardHandler(card))

	fmt.Printf("[A2A] Hosting agent %s at http://localhost:%d\n", entity.Name(), port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}
