// Package vector implements vector storage and semantic search.
package vector

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// MarkdownStore provides human-readable persistence by appending memories to an .md file.
type MarkdownStore struct {
	FilePath string
}

// NewMarkdownStore initializes a new markdown-backed storage.
func NewMarkdownStore(path string) (*MarkdownStore, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating markdown storage dir: %w", err)
	}

	return &MarkdownStore{FilePath: path}, nil
}

// Add appends memories as markdown list items.
func (s *MarkdownStore) Add(_ context.Context, items ...Item) error {
	f, err := os.OpenFile(s.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening markdown memory file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	for _, item := range items {
		line := fmt.Sprintf("- %s\n", item.Text)
		if _, err := f.WriteString(line); err != nil {
			return err
		}
	}
	return nil
}

// Search for MarkdownStore is currently a no-op (Full Context strategy).
func (s *MarkdownStore) Search(_ context.Context, _ []float64, _ int) ([]Item, error) {
	// For now, return empty. The 'Timeline' or 'Loader' usually handles 
	// providing the full text from the file if this strategy is chosen.
	return nil, nil
}
