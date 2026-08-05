package entity

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/a2aproject/a2a-go/a2a"
	"github.com/a2aproject/a2a-go/a2aclient"
	"github.com/ghchinoy/pandemonium/pkg/types"
)

// A2AProxyEntity implements the Entity interface by delegating to a remote service.
type A2AProxyEntity struct {
	EntityName string
	URL        string
	Client     *a2aclient.Client
	Bus        *types.EventBus
	ContextID  string
	Metadata   map[string]any
}

// NewA2AProxyEntity initializes a remote agent proxy.
func NewA2AProxyEntity(ctx context.Context, name, url string, bus *types.EventBus) (*A2AProxyEntity, error) {
	baseURL := strings.TrimSuffix(url, "/invoke")

	// Connectivity Check: Fetch the public Agent Card via standard HTTP
	// This ensures the server is up and identifies as an A2A agent.
	checkURL := fmt.Sprintf("%s/.well-known/agent-card.json", baseURL)
	httpClient := &http.Client{Timeout: 5 * time.Second}
	resp, err := httpClient.Get(checkURL)
	if err != nil {
		return nil, fmt.Errorf("remote agent %s at %s not responding: %w", name, baseURL, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remote agent %s at %s returned status %d", name, checkURL, resp.StatusCode)
	}

	// Create client with JSON-RPC transport
	factory := a2aclient.NewFactory(a2aclient.WithJSONRPCTransport(nil))
	endpoints := []*a2a.AgentInterface{
		a2a.NewAgentInterface(url, a2a.TransportProtocolJSONRPC),
	}

	client, err := factory.CreateFromEndpoints(ctx, endpoints)
	if err != nil {
		return nil, fmt.Errorf("initializing A2A client for %s: %w", name, err)
	}

	return &A2AProxyEntity{
		EntityName: name,
		URL:        url,
		Client:     client,
		Bus:        bus,
		Metadata:   make(map[string]any),
	}, nil
}

// SetContextID allows the engine to group all agent tasks under one simulation session.
func (e *A2AProxyEntity) SetContextID(id string) {
	e.ContextID = id
}

// SetMetadata allows the engine to pass sim-clock or step info.
func (e *A2AProxyEntity) SetMetadata(key string, val any) {
	e.Metadata[key] = val
}

// Name returns the name of the entity.
func (e *A2AProxyEntity) Name() string { return e.EntityName }

// Observe sends a message to the remote agent.
func (e *A2AProxyEntity) Observe(ctx context.Context, observation string) error {
	msg := a2a.NewMessage(a2a.MessageRoleUser, a2a.NewTextPart(observation))
	msg.ContextID = e.ContextID

	meta := make(map[string]any)
	for k, v := range e.Metadata {
		meta[k] = v
	}
	meta["type"] = "observation"

	req := a2a.SendMessageRequest{
		Message:  msg,
		Metadata: meta,
	}

	_, err := e.Client.SendMessage(ctx, &req)
	if err != nil {
		return fmt.Errorf("sending observation to %s: %w", e.EntityName, err)
	}
	return nil
}

// Act triggers a task-based action on the remote agent.
func (e *A2AProxyEntity) Act(ctx context.Context, spec types.ActionSpec) (string, error) {
	msg := a2a.NewMessage(a2a.MessageRoleUser, a2a.NewTextPart(spec.CallToAction))
	msg.ContextID = e.ContextID

	req := a2a.SendMessageRequest{
		Message:  msg,
		Metadata: e.Metadata,
	}

	// Use SendStreamingMessage for stateful/task-based interactions
	stream := e.Client.SendStreamingMessage(ctx, &req)

	var finalAction string
	reasoning := make(map[string]string)

	// Process the event stream using Go iterator
	for ev, err := range stream {
		if err != nil {
			return "", fmt.Errorf("A2A stream error: %w", err)
		}

		switch te := ev.(type) {
		case *a2a.TaskStatusUpdateEvent:
			if te.Status.Message != nil {
				for _, part := range te.Status.Message.Parts {
					if t := part.Text(); t != "" {
						finalAction = t
					}
				}
			}
		case *a2a.TaskArtifactUpdateEvent:
			if te.Artifact != nil {
				for _, part := range te.Artifact.Parts {
					if t := part.Text(); t != "" {
						reasoning[te.Artifact.Name] = t
					}
				}
			}
		case *a2a.Message:
			for _, part := range te.Parts {
				if t := part.Text(); t != "" {
					finalAction = t
				}
			}
		case *a2a.Task:
			if te.Status.Message != nil {
				for _, part := range te.Status.Message.Parts {
					if t := part.Text(); t != "" {
						finalAction = t
					}
				}
			}
			for _, art := range te.Artifacts {
				for _, part := range art.Parts {
					if t := part.Text(); t != "" {
						reasoning[art.Name] = t
					}
				}
			}
		}
	}

	if len(reasoning) > 0 && e.Bus != nil {
		e.Bus.Publish(types.Event{
			Type:    types.EventThought,
			Actor:   e.EntityName,
			Message: fmt.Sprintf("%s has finished thinking.", e.EntityName),
			Data:    reasoning,
		})
	}

	if finalAction == "" {
		return "", fmt.Errorf("remote agent %s returned no action", e.EntityName)
	}

	return finalAction, nil
}

func (e *A2AProxyEntity) emitThought(msg string) {
	if e.Bus != nil {
		e.Bus.Publish(types.Event{
			Type:    types.EventThought,
			Actor:   e.EntityName,
			Message: msg,
		})
	}
}
