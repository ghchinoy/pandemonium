package engine

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ghchinoy/pandemonium/pkg/component/memory"
	"github.com/ghchinoy/pandemonium/pkg/entity"
	"github.com/ghchinoy/pandemonium/pkg/types"
	"github.com/ghchinoy/pandemonium/pkg/vector"
)

// Checkpoint captures the serializable state of a simulation at a step boundary.
type Checkpoint struct {
	CompletedSteps    int                    `json:"completed_steps"`
	LastActor         string                 `json:"last_actor,omitempty"`
	LastRotationIndex int                    `json:"last_rotation_index,omitempty"`
	ContextID         string                 `json:"context_id,omitempty"`
	Entities          map[string]EntityState `json:"entities"`
}

// EntityState holds the serializable memory and timeline state of an entity.
type EntityState struct {
	Timeline []string      `json:"timeline,omitempty"`
	Memories []vector.Item `json:"memories,omitempty"`
}

// SaveCheckpoint writes a simulation checkpoint as indented JSON to path.
func SaveCheckpoint(path string, cp Checkpoint) error {
	data, err := json.MarshalIndent(cp, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling checkpoint: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing checkpoint %q: %w", path, err)
	}
	return nil
}

// LoadCheckpoint reads and unmarshals a simulation checkpoint from path.
func LoadCheckpoint(path string) (Checkpoint, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Checkpoint{}, fmt.Errorf("reading checkpoint %q: %w", path, err)
	}
	var cp Checkpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return Checkpoint{}, fmt.Errorf("unmarshaling checkpoint %q: %w", path, err)
	}
	return cp, nil
}

// CaptureEntityState extracts the timeline and in-memory vector items from an entity.
func CaptureEntityState(ent types.Entity) EntityState {
	var state EntityState
	base, ok := ent.(*entity.BaseEntity)
	if !ok {
		return state
	}
	for _, comp := range base.Components {
		switch c := comp.(type) {
		case *memory.TimelineMemory:
			state.Timeline = c.Events()
		case *memory.AssociativeMemory:
			if ms := extractMemoryStore(c.Store); ms != nil {
				state.Memories = ms.Items()
			}
		}
	}
	return state
}

// RestoreEntityState populates an entity's timeline and in-memory vector items from state.
func RestoreEntityState(ent types.Entity, state EntityState) {
	base, ok := ent.(*entity.BaseEntity)
	if !ok {
		return
	}
	for _, comp := range base.Components {
		switch c := comp.(type) {
		case *memory.TimelineMemory:
			c.SetEvents(state.Timeline)
		case *memory.AssociativeMemory:
			if ms := extractMemoryStore(c.Store); ms != nil {
				ms.SetItems(state.Memories)
			}
		}
	}
}

func extractMemoryStore(s vector.Store) *vector.MemoryStore {
	switch st := s.(type) {
	case *vector.MemoryStore:
		return st
	case *vector.FileStore:
		if ms, ok := st.Store.(*vector.MemoryStore); ok {
			return ms
		}
	}
	return nil
}
