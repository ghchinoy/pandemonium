// Package memory implements memory components for Pandemonium entities.
package memory

import (
	"context"
	"fmt"
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
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, event))
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

func (t *TimelineMemory) addEvent(text string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.events) >= t.Capacity {
		// Shift left to remove oldest
		t.events = t.events[1:]
	}
	t.events = append(t.events, text)
}
