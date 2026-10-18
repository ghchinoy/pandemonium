package memory

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/ghchinoy/pandemonium/pkg/types"
)

// TimelineMemory maintains a rolling chronological buffer of recent events.
type TimelineMemory struct {
	CompName string
	Capacity int

	mu     sync.RWMutex
	events []string
}

// NewTimelineMemory initializes a new chronological buffer component.
func NewTimelineMemory(name string, capacity int) *TimelineMemory {
	return &TimelineMemory{
		CompName: name,
		Capacity: capacity,
		events:   make([]string, 0, capacity),
	}
}

// Name returns the name of the component.
func (t *TimelineMemory) Name() string { return t.CompName }

// PreAct returns the recent history as a formatted string.
func (t *TimelineMemory) PreAct(_ context.Context, _ types.ActionSpec) (string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.events) == 0 {
		return "No recent events recorded.", nil
	}

	var sb strings.Builder
	sb.WriteString("Recent Timeline:\n")
	for i, event := range t.events {
		fmt.Fprintf(&sb, "%d. %s\n", i+1, event)
	}
	return sb.String(), nil
}

// PostAct adds the agent's own action to the timeline.
func (t *TimelineMemory) PostAct(_ context.Context, action string) (string, error) {
	t.addEvent(fmt.Sprintf("I decided to: %s", action))
	return "", nil
}

// PreObserve adds an incoming observation to the timeline.
func (t *TimelineMemory) PreObserve(_ context.Context, observation string) (string, error) {
	t.addEvent(observation)
	return "", nil
}

// PostObserve is a no-op for TimelineMemory.
func (t *TimelineMemory) PostObserve(_ context.Context) (string, error) {
	return "", nil
}

// Update is a no-op for TimelineMemory.
func (t *TimelineMemory) Update(_ context.Context) error { return nil }

// Events returns a copy of the current timeline buffer.
func (t *TimelineMemory) Events() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return slices.Clone(t.events)
}

// SetEvents replaces the timeline buffer contents.
func (t *TimelineMemory) SetEvents(events []string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(events) > t.Capacity && t.Capacity > 0 {
		events = events[len(events)-t.Capacity:]
	}
	t.events = slices.Clone(events)
}

func (t *TimelineMemory) addEvent(text string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.events) >= t.Capacity {
		// Shift left to remove oldest
		t.events = t.events[1:]
	}
	t.events = append(t.events, text)
}
