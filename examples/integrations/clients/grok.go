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

// GrokClient implements the llm.Client interface using xAI's Grok API.
type GrokClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewGrokClient creates a new Grok client with the provided API key.
func NewGrokClient(apiKey string) *GrokClient {
	return &GrokClient{
		apiKey:  apiKey,
		baseURL: "https://api.x.ai/v1",
		client:  &http.Client{},
	}
}

// Complete implements llm.Client.Complete using Grok's chat completions API.
func (c *GrokClient) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	// Convert to Grok format (OpenAI-compatible)
	grokReq := c.convertRequest(req)
	
	// Make API request
	reqBody, err := json.Marshal(grokReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, body)
	}
	
	var grokResp grokResponse
	if err := json.NewDecoder(resp.Body).Decode(&grokResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Convert back to our format
	return c.convertResponse(grokResp), nil
}

// Close implements llm.Client.Close (no-op for HTTP client).
func (c *GrokClient) Close() error {
	return nil
}

// convertRequest converts our request format to Grok's format.
func (c *GrokClient) convertRequest(req llm.CompletionRequest) grokRequest {
	grokReq := grokRequest{
		Model: req.Model,
	}

	// Set parameters - use provided values
	if req.MaxTokens > 0 {
		grokReq.MaxTokens = req.MaxTokens
	}
	if req.Temperature > 0 {
		grokReq.Temperature = req.Temperature
	}
	if req.TopP > 0 {
		grokReq.TopP = req.TopP
	}
	
	// Build messages
	if len(req.Messages) > 0 {
		// Use provided messages
		for _, msg := range req.Messages {
			grokReq.Messages = append(grokReq.Messages, grokMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	} else if req.Prompt != "" {
		// Convert prompt to user message
		grokReq.Messages = []grokMessage{
			{Role: "user", Content: req.Prompt},
		}
	}
	
	// Convert tools if provided (Grok supports OpenAI-compatible tools)
	if len(req.Tools) > 0 {
		for _, tool := range req.Tools {
			grokReq.Tools = append(grokReq.Tools, grokTool{
				Type: "function",
				Function: grokFunction{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.Parameters,
				},
			})
		}
		grokReq.ToolChoice = "auto"
	}
	
	// Handle JSON response formatting (if Grok supports OpenAI-compatible response_format)
	if req.ResponseType == llm.ResponseTypeJSONObject {
		grokReq.ResponseFormat = &grokResponseFormat{
			Type: "json_object",
		}
	} else if req.ResponseType == llm.ResponseTypeJSONSchema && req.JSONSchema != nil {
		grokReq.ResponseFormat = &grokResponseFormat{
			Type: "json_schema",
			JSONSchema: &grokJSONSchema{
				Name:        req.JSONSchema.Name,
				Description: req.JSONSchema.Description,
				Schema:      req.JSONSchema.Schema,
				Strict:      req.JSONSchema.Strict,
			},
		}
	}
	
	return grokReq
}

// convertResponse converts Grok's response to our format.
func (c *GrokClient) convertResponse(resp grokResponse) *llm.CompletionResponse {
	result := &llm.CompletionResponse{
		Usage: llm.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		Metadata: make(map[string]interface{}),
	}
	
	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		result.Content = choice.Message.Content
		
		// Convert tool calls if present
		for _, toolCall := range choice.Message.ToolCalls {
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				// If parsing fails, store as string
				args = map[string]interface{}{"raw": toolCall.Function.Arguments}
			}
			
			result.ToolCalls = append(result.ToolCalls, llm.ToolCall{
				ID:   toolCall.ID,
				Name: toolCall.Function.Name,
				Args: args,
			})
		}
	}
	
	return result
}

// Grok API types (OpenAI-compatible)
type grokRequest struct {
	Model          string              `json:"model"`
	Messages       []grokMessage       `json:"messages"`
	Tools          []grokTool          `json:"tools,omitempty"`
	ToolChoice     string              `json:"tool_choice,omitempty"`
	MaxTokens      int                 `json:"max_tokens,omitempty"`
	Temperature    float64             `json:"temperature,omitempty"`
	TopP           float64             `json:"top_p,omitempty"`
	ResponseFormat *grokResponseFormat `json:"response_format,omitempty"`
}

type grokMessage struct {
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	ToolCalls []grokToolCall `json:"tool_calls,omitempty"`
}

type grokTool struct {
	Type     string       `json:"type"`
	Function grokFunction `json:"function"`
}

type grokFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type grokToolCall struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Function grokFunctionCall `json:"function"`
}

type grokFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type grokResponse struct {
	Choices []grokChoice `json:"choices"`
	Usage   grokUsage    `json:"usage"`
}

type grokChoice struct {
	Message grokMessage `json:"message"`
}

type grokUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type grokResponseFormat struct {
	Type       string           `json:"type"`
	JSONSchema *grokJSONSchema `json:"json_schema,omitempty"`
}

type grokJSONSchema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"schema"`
	Strict      bool                   `json:"strict"`
}