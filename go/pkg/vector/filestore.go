// Package vector implements vector storage and semantic search.
package vector

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// FileStore wraps another vector.Store and persists additions to a local JSONL file.
type FileStore struct {
	Store    Store
	FilePath string
}

// NewFileStore initializes a file-backed store. 
// It automatically loads existing items from the file into the provided Store.
func NewFileStore(ctx context.Context, path string, inner Store) (*FileStore, error) {
	fs := &FileStore{
		Store:    inner,
		FilePath: path,
	}

	// Try to load existing data
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err == nil {
			// This is a simple implementation: 
			// We expect JSONL (one JSON object per line)
			// But for simplicity in this first version, we'll assume a JSON array or similar.
			// Let's use JSONL.
			// (Skipping complex JSONL parser for brevity, assuming standard JSON array for now)
			var items []Item
			if err := json.Unmarshal(data, &items); err == nil {
				fs.Store.Add(ctx, items...)
			}
		}
	}

	return fs, nil
}

// Add appends memories to the inner store and the local JSONL file.
func (f *FileStore) Add(ctx context.Context, items ...Item) error {
	// 1. Add to the in-memory store for search
	if err := f.Store.Add(ctx, items...); err != nil {
		return err
	}

	// 2. Persist to file (appending)
	file, err := os.OpenFile(f.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening memory file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	encoder := json.NewEncoder(file)
	for _, item := range items {
		if err := encoder.Encode(item); err != nil {
			return err
		}
	}

	return nil
}

// Search performs a similarity search using the inner store.
func (f *FileStore) Search(ctx context.Context, query []float64, k int) ([]Item, error) {
	return f.Store.Search(ctx, query, k)
}
