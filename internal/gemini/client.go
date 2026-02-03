// Package gemini provides Gemini AI integration for action item extraction.
package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"google.golang.org/genai"
)

// ActionItemResponse is the expected JSON response from Gemini.
type ActionItemResponse struct {
	Assignee string `json:"assignee"`
	Date     string `json:"date"` // YYYY-MM-DD format
}

// Client wraps the Gemini SDK for action item extraction.
type Client struct {
	client    *genai.Client
	modelName string
}

// NewClient creates a new Gemini client for action item extraction.
func NewClient(ctx context.Context, apiKey, modelName string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &Client{
		client:    client,
		modelName: modelName,
	}, nil
}

// ExtractActionItem sends a checkbox text to Gemini and extracts assignee and date.
func (c *Client) ExtractActionItem(ctx context.Context, checkboxText string) (*ActionItemResponse, error) {
	prompt := buildExtractionPrompt(checkboxText)

	result, err := c.client.Models.GenerateContent(ctx, c.modelName, genai.Text(prompt), nil)
	if err != nil {
		return nil, fmt.Errorf("Gemini API error: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}

	// Extract text from the response
	var responseText string
	for _, part := range result.Candidates[0].Content.Parts {
		if part.Text != "" {
			responseText += part.Text
		}
	}

	return parseResponse(responseText)
}

// buildExtractionPrompt creates the prompt for Gemini.
func buildExtractionPrompt(checkboxText string) string {
	return fmt.Sprintf(`You are an action item extraction assistant. Extract the assignee and due date from the following task description.

Task: "%s"

Return your response as a JSON object with exactly two fields:
- "assignee": The name of the person responsible (or null if no clear assignee)
- "date": The due date in YYYY-MM-DD format (or null if no clear date)

Rules:
1. Look for names directly mentioned as responsible
2. Interpret relative dates (e.g., "Friday" = next Friday, "tomorrow", "next week")
3. If today's date is needed, assume it is %s
4. Only return the JSON object, no other text
5. If you cannot determine assignee or date, use null for that field

Example responses:
{"assignee": "John", "date": "2026-02-07"}
{"assignee": "Sarah", "date": null}
{"assignee": null, "date": "2026-02-10"}

Your response:`, checkboxText, time.Now().Format("2006-01-02"))
}

// parseResponse parses the JSON response from Gemini.
func parseResponse(responseText string) (*ActionItemResponse, error) {
	// Clean up the response - remove markdown code blocks if present
	responseText = strings.TrimSpace(responseText)
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	// Try to find JSON object in the response
	jsonRegex := regexp.MustCompile(`\{[^}]+\}`)
	matches := jsonRegex.FindString(responseText)
	if matches != "" {
		responseText = matches
	}

	var result ActionItemResponse
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response as JSON: %w\nResponse was: %s", err, responseText)
	}

	return &result, nil
}

// ParseDate parses a YYYY-MM-DD date string.
func ParseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}
	return time.Parse("2006-01-02", dateStr)
}
