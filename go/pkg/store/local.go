package store

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/ghchinoy/pandemonium/pkg/types"
)

// LocalStore implements StateStore using local JSON files.
type LocalStore struct {
	BaseDir string
}

func NewLocalStore(baseDir string) (*LocalStore, error) {
	if err := os.MkdirAll(filepath.Join(baseDir, "scenarios"), 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(baseDir, "sessions"), 0755); err != nil {
		return nil, err
	}
	return &LocalStore{BaseDir: baseDir}, nil
}

// SaveScenario saves a scenario template as a JSON file.
func (s *LocalStore) SaveScenario(_ context.Context, sc types.Scenario) error {
	data, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.BaseDir, "scenarios", sc.Name+".json"), data, 0644)
}

// GetScenario loads a scenario template from a JSON file.
func (s *LocalStore) GetScenario(_ context.Context, name string) (types.Scenario, error) {
	data, err := os.ReadFile(filepath.Join(s.BaseDir, "scenarios", name+".json"))
	if err != nil {
		return types.Scenario{}, err
	}
	var sc types.Scenario
	err = json.Unmarshal(data, &sc)
	return sc, err
}

// ListScenarios returns the names of all scenarios in the local storage directory.
func (s *LocalStore) ListScenarios(_ context.Context) ([]string, error) {
	files, err := os.ReadDir(filepath.Join(s.BaseDir, "scenarios"))
	if err != nil {
		return nil, err
	}
	var names []string
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".json" {
			names = append(names, f.Name()[:len(f.Name())-5])
		}
	}
	return names, nil
}

// SaveEvent appends an event to a local JSONL file.
func (s *LocalStore) SaveEvent(_ context.Context, sessionID string, event types.Event) error {
	f, err := os.OpenFile(filepath.Join(s.BaseDir, "sessions", sessionID+".jsonl"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	return json.NewEncoder(f).Encode(event)
}

// GetSessionEvents is a placeholder for retrieving events from local storage.
func (s *LocalStore) GetSessionEvents(_ context.Context, _ string) ([]types.Event, error) {
	// Implementation for reading JSONL... (skipping for now for brevity)
	return nil, nil
}
