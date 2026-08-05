package vector

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
)

// Generate random vector of given dimension
func randomVector(dim int) []float64 {
	vec := make([]float64, dim)
	for i := 0; i < dim; i++ {
		vec[i] = rand.Float64()
	}
	return vec
}

func BenchmarkMemoryStore_Search(b *testing.B) {
	ctx := context.Background()
	dim := 768 // Standard Gemini embedding dimension
	
	sizes := []int{10, 100, 1000}
	
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			store := NewMemoryStore()
			for i := 0; i < size; i++ {
				store.Add(ctx, Item{
					Text:      fmt.Sprintf("Memory item %d", i),
					Embedding: randomVector(dim),
				})
			}
			
			query := randomVector(dim)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				store.Search(ctx, query, 5)
			}
		})
	}
}

func BenchmarkMarkdownStore_Add(b *testing.B) {
	ctx := context.Background()
	tmpFile := "bench_test.md"
	defer os.Remove(tmpFile)

	store, _ := NewMarkdownStore(tmpFile)
	item := Item{Text: "A short observation about the weather in the pub."}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Add(ctx, item)
	}
}
