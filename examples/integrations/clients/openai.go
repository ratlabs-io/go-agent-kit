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

// OpenAIClient implements the llm.Client interface using OpenAI's API.
type OpenAIClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewOpenAIClient creates a new OpenAI client with the provided API key.
func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		client:  &http.Client{},
	}
}

// Complete implements llm.Client.Complete using OpenAI's chat completions API.
func (c *OpenAIClient) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	// Convert to OpenAI format
	openAIReq := c.convertRequest(req)
	
	// Make API request
	reqBody, err := json.Marshal(openAIReq)
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
	
	var openAIResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Convert back to our format
	return c.convertResponse(openAIResp), nil
}

// Close implements llm.Client.Close (no-op for HTTP client).
func (c *OpenAIClient) Close() error {
	return nil
}

// convertRequest converts our request format to OpenAI's format.
func (c *OpenAIClient) convertRequest(req llm.CompletionRequest) openAIRequest {
	openAIReq := openAIRequest{
		Model: req.Model,
	}

	// Set parameters - use provided values or sensible defaults
	if req.MaxTokens > 0 {
		openAIReq.MaxTokens = req.MaxTokens
	}
	if req.Temperature > 0 {
		openAIReq.Temperature = req.Temperature
	}
	if req.TopP > 0 {
		openAIReq.TopP = req.TopP
	}
	
	// Build messages
	if len(req.Messages) > 0 {
		// Use provided messages
		for _, msg := range req.Messages {
			openAIReq.Messages = append(openAIReq.Messages, openAIMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	} else if req.Prompt != "" {
		// Convert prompt to user message
		openAIReq.Messages = []openAIMessage{
			{Role: "user", Content: req.Prompt},
		}
	}
	
	// Convert tools if provided
	if len(req.Tools) > 0 {
		for _, tool := range req.Tools {
			openAIReq.Tools = append(openAIReq.Tools, openAITool{
				Type: "function",
				Function: openAIFunction{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.Parameters,
				},
			})
		}
		openAIReq.ToolChoice = "auto"
	}
	
	// Handle JSON response formatting
	if req.ResponseType == llm.ResponseTypeJSONObject {
		openAIReq.ResponseFormat = &openAIResponseFormat{
			Type: "json_object",
		}
	} else if req.ResponseType == llm.ResponseTypeJSONSchema && req.JSONSchema != nil {
		openAIReq.ResponseFormat = &openAIResponseFormat{
			Type: "json_schema",
			JSONSchema: &openAIJSONSchema{
				Name:        req.JSONSchema.Name,
				Description: req.JSONSchema.Description,
				Schema:      req.JSONSchema.Schema,
				Strict:      req.JSONSchema.Strict,
			},
		}
	}
	
	return openAIReq
}

// convertResponse converts OpenAI's response to our format.
func (c *OpenAIClient) convertResponse(resp openAIResponse) *llm.CompletionResponse {
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

// OpenAI API types
type openAIRequest struct {
	Model          string                `json:"model"`
	Messages       []openAIMessage       `json:"messages"`
	Tools          []openAITool          `json:"tools,omitempty"`
	ToolChoice     string                `json:"tool_choice,omitempty"`
	MaxTokens      int                   `json:"max_tokens,omitempty"`
	Temperature    float64               `json:"temperature,omitempty"`
	TopP           float64               `json:"top_p,omitempty"`
	ResponseFormat *openAIResponseFormat `json:"response_format,omitempty"`
}

type openAIMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []openAIToolCall `json:"tool_calls,omitempty"`
}

type openAITool struct {
	Type     string         `json:"type"`
	Function openAIFunction `json:"function"`
}

type openAIFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type openAIToolCall struct {
	ID       string              `json:"id"`
	Type     string              `json:"type"`
	Function openAIFunctionCall `json:"function"`
}

type openAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
	Usage   openAIUsage    `json:"usage"`
}

type openAIChoice struct {
	Message openAIMessage `json:"message"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type openAIResponseFormat struct {
	Type       string             `json:"type"`
	JSONSchema *openAIJSONSchema `json:"json_schema,omitempty"`
}

type openAIJSONSchema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"schema"`
	Strict      bool                   `json:"strict"`
}