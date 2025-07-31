package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/ratlabs-io/go-agent-kit/pkg/llm"
)

// GeminiClient implements the llm.Client interface using Google's Gemini API.
type GeminiClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewGeminiClient creates a new Gemini client with the provided API key.
func NewGeminiClient(apiKey string) *GeminiClient {
	return &GeminiClient{
		apiKey:  apiKey,
		baseURL: "https://generativelanguage.googleapis.com/v1beta",
		client:  &http.Client{},
	}
}

// Complete implements llm.Client.Complete using Gemini's generateContent API.
func (c *GeminiClient) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	// Convert to Gemini format
	geminiReq := c.convertRequest(req)

	// Make API request
	reqBody, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build URL with model and API key
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, req.Model, c.apiKey)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, body)
	}

	// Read the full response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert back to our format
	return c.convertResponse(geminiResp), nil
}

// Close implements llm.Client.Close (no-op for HTTP client).
func (c *GeminiClient) Close() error {
	return nil
}

// convertRequest converts our request format to Gemini's format.
func (c *GeminiClient) convertRequest(req llm.CompletionRequest) geminiRequest {
	geminiReq := geminiRequest{
		GenerationConfig: geminiGenerationConfig{},
	}

	// Set parameters - use provided values
	if req.MaxTokens > 0 {
		geminiReq.GenerationConfig.MaxOutputTokens = req.MaxTokens
	}
	if req.Temperature > 0 {
		geminiReq.GenerationConfig.Temperature = req.Temperature
	}
	if req.TopP > 0 {
		geminiReq.GenerationConfig.TopP = req.TopP
	}

	// Build contents
	if len(req.Messages) > 0 {
		for _, msg := range req.Messages {
			role := "user"
			if msg.Role == "assistant" {
				role = "model"
			}

			geminiReq.Contents = append(geminiReq.Contents, geminiContent{
				Role: role,
				Parts: []geminiPart{
					{Text: msg.Content},
				},
			})
		}
	} else if req.Prompt != "" {
		geminiReq.Contents = []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{Text: req.Prompt},
				},
			},
		}
	}

	// Convert tools if provided
	if len(req.Tools) > 0 {
		var functionDeclarations []geminiFunctionDeclaration
		for _, tool := range req.Tools {
			functionDeclarations = append(functionDeclarations, geminiFunctionDeclaration{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Parameters,
			})
		}

		geminiReq.Tools = []geminiTool{
			{FunctionDeclarations: functionDeclarations},
		}
	}

	// Handle JSON response formatting using Gemini's native JSON schema support
	if req.ResponseType == llm.ResponseTypeJSONObject || req.ResponseType == llm.ResponseTypeJSONSchema {
		// Set response MIME type for JSON
		geminiReq.GenerationConfig.ResponseMimeType = "application/json"

		// For JSON schema, set the schema in generation config
		if req.ResponseType == llm.ResponseTypeJSONSchema && req.JSONSchema != nil {
			// Clean the schema for Gemini compatibility
			cleanedSchema := c.cleanSchemaForGemini(req.JSONSchema.Schema)
			geminiReq.GenerationConfig.ResponseSchema = cleanedSchema
		} else if req.ResponseType == llm.ResponseTypeJSONObject {
			// For generic JSON object without schema, provide a basic schema
			geminiReq.GenerationConfig.ResponseSchema = map[string]interface{}{
				"type": "object",
			}
		}
	}

	return geminiReq
}

// convertResponse converts Gemini's response to our format.
func (c *GeminiClient) convertResponse(resp geminiResponse) *llm.CompletionResponse {
	result := &llm.CompletionResponse{
		Usage: llm.Usage{
			PromptTokens:     resp.UsageMetadata.PromptTokenCount,
			CompletionTokens: resp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      resp.UsageMetadata.TotalTokenCount,
		},
		Metadata: make(map[string]interface{}),
	}

	if len(resp.Candidates) > 0 {
		candidate := resp.Candidates[0]
		if candidate.Content.Parts != nil {
			for _, part := range candidate.Content.Parts {
				if part.Text != "" {
					result.Content += part.Text
				} else if part.FunctionCall != nil {
					// Convert function calls to tool calls
					result.ToolCalls = append(result.ToolCalls, llm.ToolCall{
						ID:   fmt.Sprintf("call_%d", len(result.ToolCalls)), // Generate ID
						Name: part.FunctionCall.Name,
						Args: part.FunctionCall.Args,
					})
				}
			}
		}
	}

	return result
}

// cleanSchemaForGemini removes schema properties that Gemini doesn't support
func (c *GeminiClient) cleanSchemaForGemini(schema map[string]interface{}) map[string]interface{} {
	cleaned := make(map[string]interface{})

	// List of properties that Gemini doesn't support
	// Based on Gemini documentation, they support a subset of OpenAPI 3.0 Schema
	unsupportedProps := map[string]bool{
		"additionalProperties": true, // Gemini explicitly rejects this
		"$schema":              true, // JSON Schema metadata
		"$id":                  true, // JSON Schema identifier
		"$ref":                 true, // JSON Schema references
	}

	for key, value := range schema {
		if unsupportedProps[key] {
			continue
		}

		// Recursively clean nested objects
		if obj, ok := value.(map[string]interface{}); ok {
			cleaned[key] = c.cleanSchemaForGemini(obj)
		} else if arr, ok := value.([]interface{}); ok {
			// Clean array elements that are objects
			cleanedArr := make([]interface{}, len(arr))
			for i, item := range arr {
				if obj, ok := item.(map[string]interface{}); ok {
					cleanedArr[i] = c.cleanSchemaForGemini(obj)
				} else {
					cleanedArr[i] = item
				}
			}
			cleaned[key] = cleanedArr
		} else {
			cleaned[key] = value
		}
	}

	return cleaned
}

// cleanJSONResponse removes markdown code block formatting
func (c *GeminiClient) cleanJSONResponse(content string) string {
	// Remove leading/trailing whitespace
	content = strings.TrimSpace(content)

	// Pattern to match markdown code blocks with optional language specifier
	// Matches: ```json\n{...}\n``` or ```\n{...}\n```
	// Use (?s) flag to make . match newlines, and use non-greedy matching
	codeBlockRegex := regexp.MustCompile(`(?s)^` + "```" + `(?:json)?\s*\n?(.*?)\n?` + "```" + `\s*$`)

	if matches := codeBlockRegex.FindStringSubmatch(content); len(matches) > 1 {
		// Extract content from inside code blocks
		content = strings.TrimSpace(matches[1])
	} else {
		// If not wrapped entirely, look for embedded code blocks
		// This handles cases where JSON is embedded in a larger response
		embeddedCodeBlockRegex := regexp.MustCompile(`(?s)` + "```" + `(?:json)?\s*\n?(.*?)\n?` + "```")
		if matches := embeddedCodeBlockRegex.FindStringSubmatch(content); len(matches) > 1 {
			content = strings.TrimSpace(matches[1])
		}
	}

	return content
}

// Gemini API types
type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	Tools            []geminiTool           `json:"tools,omitempty"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text         string              `json:"text,omitempty"`
	FunctionCall *geminiFunctionCall `json:"functionCall,omitempty"`
}

type geminiFunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

type geminiTool struct {
	FunctionDeclarations []geminiFunctionDeclaration `json:"functionDeclarations"`
}

type geminiFunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type geminiGenerationConfig struct {
	Temperature      float64                `json:"temperature,omitempty"`
	MaxOutputTokens  int                    `json:"maxOutputTokens,omitempty"`
	TopP             float64                `json:"topP,omitempty"`
	ResponseMimeType string                 `json:"responseMimeType,omitempty"`
	ResponseSchema   map[string]interface{} `json:"responseSchema,omitempty"`
}

type geminiResponse struct {
	Candidates    []geminiCandidate   `json:"candidates"`
	UsageMetadata geminiUsageMetadata `json:"usageMetadata"`
}

type geminiCandidate struct {
	Content geminiContent `json:"content"`
}

type geminiUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}