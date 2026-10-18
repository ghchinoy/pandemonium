package vector_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ghchinoy/pandemonium/pkg/vector"
)

func TestMemoryStore_Search(t *testing.T) {
	ctx := t.Context()
	store := vector.NewMemoryStore()

	items := []vector.Item{
		{Text: "apple", Embedding: []float64{1, 0, 0}},
		{Text: "apricot", Embedding: []float64{0.9, 0.1, 0}},
		{Text: "banana", Embedding: []float64{0, 1, 0}},
		{Text: "wrong_dim", Embedding: []float64{1, 0}},
	}
	if err := store.Add(ctx, items...); err != nil {
		t.Fatalf("store.Add(ctx, %+v) returned unexpected error: %v", items, err)
	}

	tests := []struct {
		name  string
		query []float64
		k     int
		want  []vector.Item
	}{
		{
			name:  "top_2_similar",
			query: []float64{1, 0, 0},
			k:     2,
			want: []vector.Item{
				{Text: "apple", Embedding: []float64{1, 0, 0}},
				{Text: "apricot", Embedding: []float64{0.9, 0.1, 0}},
			},
		},
		{
			name:  "k_larger_than_matching_items",
			query: []float64{0, 1, 0},
			k:     10,
			want: []vector.Item{
				{Text: "banana", Embedding: []float64{0, 1, 0}},
				{Text: "apricot", Embedding: []float64{0.9, 0.1, 0}},
				{Text: "apple", Embedding: []float64{1, 0, 0}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := store.Search(ctx, tc.query, tc.k)
			if err != nil {
				t.Fatalf("store.Search(ctx, %v, %d) returned unexpected error: %v", tc.query, tc.k, err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("store.Search(ctx, %v, %d) returned unexpected diff (-want +got):\n%s", tc.query, tc.k, diff)
			}
		})
	}
}

func TestFileStore_AddAndReload(t *testing.T) {
	ctx := t.Context()
	path := filepath.Join(t.TempDir(), "memories.jsonl")

	fs1, err := vector.NewFileStore(ctx, path, vector.NewMemoryStore())
	if err != nil {
		t.Fatalf("NewFileStore(ctx, %q, inner) returned unexpected error: %v", path, err)
	}

	firstBatch := []vector.Item{
		{Text: "first memory", Embedding: []float64{1, 0, 0}},
		{Text: "second memory", Embedding: []float64{0, 1, 0}},
	}
	if err := fs1.Add(ctx, firstBatch...); err != nil {
		t.Fatalf("fs1.Add(ctx, %+v) returned unexpected error: %v", firstBatch, err)
	}

	thirdItem := vector.Item{Text: "third memory", Embedding: []float64{0, 0, 1}}
	if err := fs1.Add(ctx, thirdItem); err != nil {
		t.Fatalf("fs1.Add(ctx, %+v) returned unexpected error: %v", thirdItem, err)
	}

	// Reload into a fresh MemoryStore to verify JSONL persistence round-trip.
	fs2, err := vector.NewFileStore(ctx, path, vector.NewMemoryStore())
	if err != nil {
		t.Fatalf("NewFileStore(ctx, %q, freshInner) on reload returned unexpected error: %v", path, err)
	}

	query := []float64{0, 0, 1}
	got, err := fs2.Search(ctx, query, 3)
	if err != nil {
		t.Fatalf("fs2.Search(ctx, %v, 3) returned unexpected error: %v", query, err)
	}
	if len(got) != 3 {
		t.Fatalf("len(fs2.Search(ctx, %v, 3)) = %d, want 3 (got=%+v)", query, len(got), got)
	}
	if got[0].Text != "third memory" {
		t.Errorf("fs2.Search(ctx, %v, 3)[0].Text = %q, want %q", query, got[0].Text, "third memory")
	}
}

func TestMarkdownStore_Add(t *testing.T) {
	ctx := t.Context()
	path := filepath.Join(t.TempDir(), "nested", "memories.md")

	ms, err := vector.NewMarkdownStore(path)
	if err != nil {
		t.Fatalf("NewMarkdownStore(%q) returned unexpected error: %v", path, err)
	}

	items := []vector.Item{
		{Text: "Entered the Rusty Anchor."},
		{Text: "Saw one cookie on the plate."},
	}
	if err := ms.Add(ctx, items...); err != nil {
		t.Fatalf("ms.Add(ctx, %+v) returned unexpected error: %v", items, err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) returned unexpected error: %v", path, err)
	}
	got := string(data)
	want := "- Entered the Rusty Anchor.\n- Saw one cookie on the plate.\n"
	if got != want {
		t.Errorf("MarkdownStore file content = %q, want %q", got, want)
	}
}
