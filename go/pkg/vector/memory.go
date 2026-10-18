package vector

import (
	"cmp"
	"context"
	"slices"
	"sync"

	"gonum.org/v1/gonum/mat"
)

// Item represents a single piece of embedded text.
type Item struct {
	Text      string    `json:"text" yaml:"text"`
	Embedding []float64 `json:"embedding,omitempty" yaml:"embedding,omitempty"`
}

// MemoryStore implements an in-memory vector storage with optimized Gonum matrix search.
type MemoryStore struct {
	mu    sync.RWMutex
	items []Item
}

// NewMemoryStore initializes a new in-memory vector store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

// Add appends new items to the store.
func (s *MemoryStore) Add(_ context.Context, items ...Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, items...)
	return nil
}

// Items returns a copy of all items stored in memory.
func (s *MemoryStore) Items() []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return slices.Clone(s.items)
}

// SetItems replaces the stored items in memory.
func (s *MemoryStore) SetItems(items []Item) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = slices.Clone(items)
}

// Search performs a high-performance cosine similarity search against the stored items using Gonum.
func (s *MemoryStore) Search(_ context.Context, query []float64, k int) ([]Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.items) == 0 {
		return nil, nil
	}

	// 1. Convert query to Gonum vector
	queryVec := mat.NewVecDense(len(query), query)
	queryNorm := mat.Norm(queryVec, 2)

	type scoredItem struct {
		item  Item
		score float64
	}
	scores := make([]scoredItem, 0, len(s.items))

	// 2. Batch calculate similarities
	// In a larger implementation, we would stack items into a Matrix
	// and use mat.MulVec for O(1) bulk calculation.
	for _, item := range s.items {
		if len(item.Embedding) != len(query) {
			continue
		}
		itemVec := mat.NewVecDense(len(item.Embedding), item.Embedding)

		// Cosine Similarity = (A . B) / (||A|| * ||B||)
		dot := mat.Dot(queryVec, itemVec)
		itemNorm := mat.Norm(itemVec, 2)

		score := 0.0
		if queryNorm > 0 && itemNorm > 0 {
			score = dot / (queryNorm * itemNorm)
		}

		scores = append(scores, scoredItem{item: item, score: score})
	}

	// 3. Sort by score descending
	slices.SortFunc(scores, func(a, b scoredItem) int {
		return cmp.Compare(b.score, a.score)
	})

	// 4. Return top K
	limit := min(k, len(scores))

	result := make([]Item, limit)
	for i := 0; i < limit; i++ {
		result[i] = scores[i].item
	}

	return result, nil
}
