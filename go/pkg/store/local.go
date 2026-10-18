package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghchinoy/pandemonium/pkg/types"
)

// LocalStore implements StateStore using local JSON files.
type LocalStore struct {
	BaseDir string
}

// NewLocalStore initializes a LocalStore rooted at baseDir and ensures required subdirectories exist.
func NewLocalStore(baseDir string) (*LocalStore, error) {
	if err := os.MkdirAll(filepath.Join(baseDir, "scenarios"), 0755); err != nil {
		return nil, fmt.Errorf("creating scenarios directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(baseDir, "sessions"), 0755); err != nil {
		return nil, fmt.Errorf("creating sessions directory: %w", err)
	}
	return &LocalStore{BaseDir: baseDir}, nil
}

// SaveScenario saves a scenario template as a JSON file.
func (s *LocalStore) SaveScenario(_ context.Context, sc types.Scenario) error {
	data, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling scenario %q: %w", sc.Name, err)
	}
	path := filepath.Join(s.BaseDir, "scenarios", sc.Name+".json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing scenario file %q: %w", path, err)
	}
	return nil
}

// GetScenario loads a scenario template from a JSON file.
func (s *LocalStore) GetScenario(_ context.Context, name string) (types.Scenario, error) {
	path := filepath.Join(s.BaseDir, "scenarios", name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return types.Scenario{}, fmt.Errorf("reading scenario file %q: %w", path, err)
	}
	var sc types.Scenario
	if err := json.Unmarshal(data, &sc); err != nil {
		return types.Scenario{}, fmt.Errorf("unmarshaling scenario %q: %w", name, err)
	}
	return sc, nil
}

// ListScenarios returns the names of all scenarios in the local storage directory.
func (s *LocalStore) ListScenarios(_ context.Context) ([]string, error) {
	dir := filepath.Join(s.BaseDir, "scenarios")
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading scenarios directory %q: %w", dir, err)
	}
	var names []string
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".json" {
			names = append(names, strings.TrimSuffix(f.Name(), ".json"))
		}
	}
	return names, nil
}

// SaveEvent appends an event to a local JSONL file.
func (s *LocalStore) SaveEvent(_ context.Context, sessionID string, event types.Event) error {
	path := filepath.Join(s.BaseDir, "sessions", sessionID+".jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening session file %q: %w", path, err)
	}
	defer func() {
		_ = f.Close()
	}()
	if err := json.NewEncoder(f).Encode(event); err != nil {
		return fmt.Errorf("encoding event for session %q: %w", sessionID, err)
	}
	return nil
}

// GetSessionEvents retrieves all events for a session from its local JSONL file.
func (s *LocalStore) GetSessionEvents(_ context.Context, sessionID string) ([]types.Event, error) {
	path := filepath.Join(s.BaseDir, "sessions", sessionID+".jsonl")
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("opening session file %q: %w", path, err)
	}
	defer func() {
		_ = f.Close()
	}()

	var events []types.Event
	decoder := json.NewDecoder(f)
	for {
		var event types.Event
		if err := decoder.Decode(&event); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("decoding event for session %q: %w", sessionID, err)
		}
		events = append(events, event)
	}
	return events, nil
}
