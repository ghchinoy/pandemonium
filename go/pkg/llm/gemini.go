// Package llm provides language model integration for Pandemonium.
package llm

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"google.golang.org/genai"
)

// Model defines the interface for text generation.
type Model interface {
	Generate(ctx context.Context, prompt string) (string, error)
	GenerateWithSystem(ctx context.Context, system, prompt string) (string, error)
}

// GeminiModel implements the Model interface using the Google GenAI SDK.
type GeminiModel struct {
	client *genai.Client
	model  string
}

// NewGeminiModel initializes a new Gemini client.
func NewGeminiModel(ctx context.Context) (*GeminiModel, error) {
	modelName := os.Getenv("GEMINI_MODEL")
	if modelName == "" {
		modelName = "gemini-3-flash-preview"
	}

	client, err := newGenAIClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("initializing genai client: %w", err)
	}

	return &GeminiModel{
		client: client,
		model:  modelName,
	}, nil
}

// Generate is a convenience method for simple prompts.
func (m *GeminiModel) Generate(ctx context.Context, prompt string) (string, error) {
	return m.GenerateWithSystem(ctx, "", prompt)
}

// GenerateWithSystem sends a prompt with explicit system instructions.
func (m *GeminiModel) GenerateWithSystem(ctx context.Context, system, prompt string) (string, error) {
	// Add delay for visualization
	time.Sleep(800 * time.Millisecond)

	if m.client == nil {
		// Generic deterministic mock for testing without API keys.
		// It avoids character-specific details to ensure the engine logic is decoupled.
		if strings.Contains(prompt, "who should act next") {
			return "Alice", nil
		}
		if strings.Contains(prompt, "is the simulation finished") {
			return "No", nil
		}
		return "The character performs a thoughtful action based on their current surroundings and goals.", nil
	}

	config := &genai.GenerateContentConfig{}
	if system != "" {
		contents := genai.Text(system)
		config.SystemInstruction = contents[0]
	}

	resp, err := m.client.Models.GenerateContent(ctx, m.model, genai.Text(prompt), config)
	if err != nil {
		return "", fmt.Errorf("generating content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content returned from model")
	}

	var result strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Text != "" {
			result.WriteString(part.Text)
		}
	}

	return result.String(), nil
}

// GeminiEmbedder implements the vector.Embedder interface using Google's embedding models.
type GeminiEmbedder struct {
	client *genai.Client
	model  string
}

// NewGeminiEmbedder initializes a new embedder.
func NewGeminiEmbedder(ctx context.Context) (*GeminiEmbedder, error) {
	modelName := os.Getenv("GEMINI_EMBEDDING_MODEL")
	if modelName == "" {
		modelName = "gemini-embedding-001"
	}

	client, err := newGenAIClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("initializing genai client for embedding: %w", err)
	}

	return &GeminiEmbedder{
		client: client,
		model:  modelName,
	}, nil
}

// Embed converts text into a high-dimensional semantic vector.
func (e *GeminiEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
	if e.client == nil {
		// Fallback for mock mode
		return []float64{float64(len(text)), 0, 0}, nil
	}

	res, err := e.client.Models.EmbedContent(ctx, e.model, genai.Text(text), nil)
	if err != nil {
		return nil, fmt.Errorf("embedding content: %w", err)
	}

	if len(res.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	vector := make([]float64, len(res.Embeddings[0].Values))
	for i, v := range res.Embeddings[0].Values {
		vector[i] = float64(v)
	}

	return vector, nil
}

func newGenAIClient(ctx context.Context) (*genai.Client, error) {
	project := os.Getenv("GOOGLE_CLOUD_PROJECT")
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	apiKey := os.Getenv("GOOGLE_GENAI_API_KEY")

	if project != "" {
		if location == "" {
			location = "global"
		}
		return genai.NewClient(ctx, &genai.ClientConfig{
			Project:  project,
			Location: location,
			Backend:  genai.BackendVertexAI,
		})
	}
	if apiKey != "" {
		return genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:  apiKey,
			Backend: genai.BackendGeminiAPI,
		})
	}
	return nil, nil
}
