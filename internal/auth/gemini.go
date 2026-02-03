// Package auth provides authentication for Google Workspace APIs and Gemini.
package auth

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/genai"
)

// GeminiClient wraps the Gemini SDK client with API key authentication.
type GeminiClient struct {
	client    *genai.Client
	modelName string
}

// NewGeminiClient creates a new Gemini client using the provided API key.
func NewGeminiClient(ctx context.Context, apiKey, modelName string) (*GeminiClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiClient{
		client:    client,
		modelName: modelName,
	}, nil
}

// NewGeminiClientFromEnv creates a new Gemini client using GEMINI_API_KEY environment variable.
func NewGeminiClientFromEnv(ctx context.Context, modelName string) (*GeminiClient, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	return NewGeminiClient(ctx, apiKey, modelName)
}

// Client returns the underlying genai.Client.
func (g *GeminiClient) Client() *genai.Client {
	return g.client
}

// ModelName returns the configured model name.
func (g *GeminiClient) ModelName() string {
	return g.modelName
}

// Close closes the Gemini client.
func (g *GeminiClient) Close() error {
	// The genai client doesn't have a Close method in current SDK
	// This is a placeholder for future cleanup if needed
	return nil
}
