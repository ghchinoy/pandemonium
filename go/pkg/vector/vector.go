// Package vector implements vector storage and semantic search.
package vector

import (
	"context"
)

// Store defines the interface for semantic and chronological memory retrieval.
type Store interface {
	Add(ctx context.Context, items ...Item) error
	Search(ctx context.Context, query []float64, k int) ([]Item, error)
}

// Embedder defines the interface for generating embeddings.
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float64, error)
}
