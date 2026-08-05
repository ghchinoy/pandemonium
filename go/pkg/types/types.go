// Package types provides core interfaces and simulation primitives.
package types

import (
	"context"
	"sync"
)

// OutputType defines the format of the entity's response.
type OutputType string

const (
	// OutputFree represents a free-form text response.
	OutputFree OutputType = "free"
	// OutputChoice represents a multiple-choice response.
	OutputChoice OutputType = "choice"
	// OutputFloat represents a floating-point number response.
	OutputFloat OutputType = "float"

	// OutputMakeObservation instructs the entity to generate an observation.
	OutputMakeObservation OutputType = "make_observation"
	// OutputNextActing instructs the GM to select the next player to act.
	OutputNextActing OutputType = "next_acting"
	// OutputNextActionSpec defines the format for the next action.
	OutputNextActionSpec OutputType = "next_action_spec"
	// OutputResolve instructs the GM to resolve a player action.
	OutputResolve OutputType = "resolve"
	// OutputTerminate checks if the simulation should end.
	OutputTerminate OutputType = "terminate"
	// OutputNextGameMaster selects the next GM if multiple are present.
	OutputNextGameMaster OutputType = "next_game_master"
	// OutputSkipThisStep skips the current simulation step.
	OutputSkipThisStep OutputType = "skip_this_step"
)

// ActionSpec specifies what the entity is being asked to do.
type ActionSpec struct {
	CallToAction string     `json:"call_to_action"`
	OutputType   OutputType `json:"output_type"`
	Options      []string   `json:"options,omitempty"`
	Tag          string     `json:"tag,omitempty"`
}

// Entity is the base interface for all actors (Agents and Game Masters).
type Entity interface {
	Name() string
	Act(ctx context.Context, spec ActionSpec) (string, error)
	Observe(ctx context.Context, observation string) error
}

// A2AEntity is an optional interface for remote agents that support session correlation.
type A2AEntity interface {
	Entity
	SetContextID(id string)
	SetMetadata(key string, val any)
}

// EventType categorizes simulation events.
type EventType string

const (
	// EventSimStart occurs when a simulation begins.
	EventSimStart EventType = "sim_start"
	// EventSimComplete occurs when a simulation finishes.
	EventSimComplete EventType = "sim_complete"
	// EventStepStart marks the beginning of a new turn.
	EventStepStart EventType = "step_start"
	// EventObservation represents what an agent senses.
	EventObservation EventType = "observation"
	// EventAction represents an agent's intended action.
	EventAction EventType = "action"
	// EventResolution represents the world's response to an action.
	EventResolution EventType = "resolution"
	// EventThought represents internal agent reasoning.
	EventThought EventType = "thought" // Internal reasoning
)

// Event represents a single simulation update.
type Event struct {
	Type    EventType `json:"type"`
	Actor   string    `json:"actor,omitempty"`
	Message string    `json:"message"`
	Data    any       `json:"data,omitempty"`
}

// EventBus handles broadcasting simulation events to multiple listeners.
type EventBus struct {
	mu        sync.RWMutex
	listeners []chan Event
	closed    bool
}

// NewEventBus initializes a new multi-consumer event bus.
func NewEventBus() *EventBus {
	return &EventBus{
		listeners: make([]chan Event, 0),
	}
}

// Subscribe adds a new listener and returns its channel.
func (b *EventBus) Subscribe() chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan Event, 100)
	b.listeners = append(b.listeners, ch)
	return ch
}

// Publish sends an event to all subscribers.
func (b *EventBus) Publish(e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return
	}

	for _, ch := range b.listeners {
		// Non-blocking send to avoid hanging the simulation
		select {
		case ch <- e:
		default:
		}
	}
}

// Close closes all listener channels.
func (b *EventBus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}
	b.closed = true
	for _, ch := range b.listeners {
		close(ch)
	}
}
