// Package memory implements associative and chronological memory components for Pandemonium entities.
package memory

import (
	"context"
	"fmt"
	"strings"

	"github.com/ghchinoy/pandemonium/pkg/types"
	"github.com/ghchinoy/pandemonium/pkg/vector"
)

// AssociativeMemory implements associative retrieval from a vector store.
type AssociativeMemory struct {
	Store    vector.Store
	Embedder vector.Embedder
	K        int
}

// Name returns the name of the component.
func (m *AssociativeMemory) Name() string {
	return "AssociativeMemory"
}

// PreAct retrieves relevant memories based on the action specification.
func (m *AssociativeMemory) PreAct(ctx context.Context, spec types.ActionSpec) (string, error) {
	// 1. Embed the call to action
	query, err := m.Embedder.Embed(ctx, spec.CallToAction)
	if err != nil {
		return "", fmt.Errorf("embedding query: %w", err)
	}

	// 2. Search
	items, err := m.Store.Search(ctx, query, m.K)
	if err != nil {
		return "", fmt.Errorf("searching memory: %w", err)
	}

	// 3. Format results
	var sb strings.Builder
	sb.WriteString("Relevant memories:\n")
	for _, item := range items {
		fmt.Fprintf(&sb, "- %s\n", item.Text)
	}

	return sb.String(), nil
}

// PostAct adds the agent's action to the memory bank.
func (m *AssociativeMemory) PostAct(ctx context.Context, action string) (string, error) {
	// Add the action to memory
	return "", m.addMemory(ctx, fmt.Sprintf("I decided to: %s", action))
}

// PreObserve adds an incoming observation to the memory bank.
func (m *AssociativeMemory) PreObserve(ctx context.Context, observation string) (string, error) {
	// Add observation to memory
	return "", m.addMemory(ctx, observation)
}

// PostObserve is a no-op for AssociativeMemory.
func (m *AssociativeMemory) PostObserve(_ context.Context) (string, error) {
	return "", nil
}

// Update is a no-op for AssociativeMemory.
func (m *AssociativeMemory) Update(_ context.Context) error {
	return nil
}

func (m *AssociativeMemory) addMemory(ctx context.Context, text string) error {
	emb, err := m.Embedder.Embed(ctx, text)
	if err != nil {
		return err
	}
	return m.Store.Add(ctx, vector.Item{
		Text:      text,
		Embedding: emb,
	})
}
