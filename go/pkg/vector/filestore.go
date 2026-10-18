package vector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fs, nil
		}
		return nil, fmt.Errorf("opening memory file %q: %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()

	var items []Item
	decoder := json.NewDecoder(file)
	for {
		var item Item
		if err := decoder.Decode(&item); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("decoding memory file %q: %w", path, err)
		}
		items = append(items, item)
	}

	if len(items) > 0 {
		if err := fs.Store.Add(ctx, items...); err != nil {
			return nil, fmt.Errorf("loading memory items from %q: %w", path, err)
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
		return fmt.Errorf("opening memory file %q: %w", f.FilePath, err)
	}
	defer func() {
		_ = file.Close()
	}()

	encoder := json.NewEncoder(file)
	for _, item := range items {
		if err := encoder.Encode(item); err != nil {
			return fmt.Errorf("encoding memory item to %q: %w", f.FilePath, err)
		}
	}

	return nil
}

// Search performs a similarity search using the inner store.
func (f *FileStore) Search(ctx context.Context, query []float64, k int) ([]Item, error) {
	return f.Store.Search(ctx, query, k)
}
