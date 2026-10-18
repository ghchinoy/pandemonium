package vector

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"
)

func BenchmarkMemoryStore_Search(b *testing.B) {
	ctx := b.Context()
	dim := 768 // Standard Gemini embedding dimension

	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			store := NewMemoryStore()
			for i := 0; i < size; i++ {
				if err := store.Add(ctx, Item{
					Text:      fmt.Sprintf("Memory item %d", i),
					Embedding: randomVector(dim),
				}); err != nil {
					b.Fatalf("store.Add() returned unexpected error: %v", err)
				}
			}

			query := randomVector(dim)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := store.Search(ctx, query, 5); err != nil {
					b.Fatalf("store.Search(ctx, query, 5) returned unexpected error: %v", err)
				}
			}
		})
	}
}

func BenchmarkMarkdownStore_Add(b *testing.B) {
	ctx := b.Context()
	tmpFile := filepath.Join(b.TempDir(), "bench_test.md")

	store, err := NewMarkdownStore(tmpFile)
	if err != nil {
		b.Fatalf("NewMarkdownStore(%q) returned unexpected error: %v", tmpFile, err)
	}
	item := Item{Text: "A short observation about the weather in the pub."}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := store.Add(ctx, item); err != nil {
			b.Fatalf("store.Add(ctx, %+v) returned unexpected error: %v", item, err)
		}
	}
}

func randomVector(dim int) []float64 {
	vec := make([]float64, dim)
	for i := 0; i < dim; i++ {
		vec[i] = rand.Float64()
	}
	return vec
}
