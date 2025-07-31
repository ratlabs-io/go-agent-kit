package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ratlabs-io/go-agent-kit/pkg/llm"
)

// AnthropicClient implements the llm.Client interface using Anthropic's API.
type AnthropicClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewAnthropicClient creates a new Anthropic client with the provided API key.
func NewAnthropicClient(apiKey string) *AnthropicClient {
	return &AnthropicClient{
		apiKey:  apiKey,
		baseURL: "https://api.anthropic.com/v1",
		client:  &http.Client{},
	}
}

// Complete implements llm.Client.Complete using Anthropic's messages API.
func (c *AnthropicClient) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	// Convert to Anthropic format
	anthropicReq := c.convertRequest(req)
	
	// Make API request
	reqBody, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, body)
	}
	
	var anthropicResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Convert back to our format
	return c.convertResponse(anthropicResp), nil
}

// Close implements llm.Client.Close (no-op for HTTP client).
func (c *AnthropicClient) Close() error {
	return nil
}

// convertRequest converts our request format to Anthropic's format.
func (c *AnthropicClient) convertRequest(req llm.CompletionRequest) anthropicRequest {
	anthropicReq := anthropicRequest{
		Model: req.Model,
	}

	// Set parameters - use provided values or fall back to Anthropic's API defaults
	if req.MaxTokens > 0 {
		anthropicReq.MaxTokens = req.MaxTokens
	} else {
		anthropicReq.MaxTokens = 1024 // Anthropic requires max_tokens to be set
	}
	if req.Temperature > 0 {
		anthropicReq.Temperature = req.Temperature
	}
	if req.TopP > 0 {
		anthropicReq.TopP = req.TopP
	}
	
	// Build messages - Anthropic separates system from messages
	if len(req.Messages) > 0 {
		for _, msg := range req.Messages {
			if msg.Role == "system" {
				anthropicReq.System = msg.Content
			} else {
				anthropicReq.Messages = append(anthropicReq.Messages, anthropicMessage{
					Role:    msg.Role,
					Content: msg.Content,
				})
			}
		}
	} else if req.Prompt != "" {
		// Convert prompt to user message
		anthropicReq.Messages = []anthropicMessage{
			{Role: "user", Content: req.Prompt},
		}
	}
	
	// Convert tools if provided
	if len(req.Tools) > 0 {
		for _, tool := range req.Tools {
			anthropicReq.Tools = append(anthropicReq.Tools, anthropicTool{
				Name:        tool.Name,
				Description: tool.Description,
				InputSchema: tool.Parameters,
			})
		}
	}
	
	// Handle JSON response formatting - Anthropic doesn't have structured outputs like OpenAI,
	// but we can modify the system prompt to encourage JSON responses
	if req.ResponseType == llm.ResponseTypeJSONObject || req.ResponseType == llm.ResponseTypeJSONSchema {
		if anthropicReq.System == "" {
			anthropicReq.System = "You must respond with valid JSON only."
		} else {
			anthropicReq.System += "\n\nIMPORTANT: You must respond with valid JSON only."
		}
		
		// For JSON schema, add the schema description to the system prompt
		if req.ResponseType == llm.ResponseTypeJSONSchema && req.JSONSchema != nil {
			anthropicReq.System += fmt.Sprintf("\n\nYour response must conform to this JSON schema: %s", req.JSONSchema.Description)
			if schemaStr, err := json.Marshal(req.JSONSchema.Schema); err == nil {
				anthropicReq.System += fmt.Sprintf("\nSchema: %s", string(schemaStr))
			}
		}
	}
	
	return anthropicReq
}

// convertResponse converts Anthropic's response to our format.
func (c *AnthropicClient) convertResponse(resp anthropicResponse) *llm.CompletionResponse {
	result := &llm.CompletionResponse{
		Usage: llm.Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
		Metadata: make(map[string]interface{}),
	}
	
	// Extract content from response
	for _, content := range resp.Content {
		if content.Type == "text" {
			result.Content += content.Text
		} else if content.Type == "tool_use" {
			// Convert tool calls
			result.ToolCalls = append(result.ToolCalls, llm.ToolCall{
				ID:   content.ID,
				Name: content.Name,
				Args: content.Input,
			})
		}
	}
	
	return result
}

// Anthropic API types
type anthropicRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float64            `json:"temperature,omitempty"`
	TopP        float64            `json:"top_p,omitempty"`
	System      string             `json:"system,omitempty"`
	Messages    []anthropicMessage `json:"messages"`
	Tools       []anthropicTool    `json:"tools,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

type anthropicResponse struct {
	Content []anthropicContent `json:"content"`
	Usage   anthropicUsage     `json:"usage"`
}

type anthropicContent struct {
	Type  string                 `json:"type"`
	Text  string                 `json:"text,omitempty"`
	ID    string                 `json:"id,omitempty"`
	Name  string                 `json:"name,omitempty"`
	Input map[string]interface{} `json:"input,omitempty"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}