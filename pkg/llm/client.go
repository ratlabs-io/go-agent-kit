package llm

import (
	"context"
)

// CompletionRequest represents a request for LLM completion.
type CompletionRequest struct {
	Model    string                 `json:"model"`
	Prompt   string                 `json:"prompt"`
	Messages []Message              `json:"messages,omitempty"`
	Tools    []ToolDefinition       `json:"tools,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`    // "system", "user", "assistant", "tool"
	Content string `json:"content"`
	Name    string `json:"name,omitempty"` // For tool messages
}

// ToolDefinition represents a tool that can be called by the LLM.
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"` // JSON Schema
}

// CompletionResponse represents the response from an LLM completion.
type CompletionResponse struct {
	Content   string                 `json:"content"`
	ToolCalls []ToolCall             `json:"tool_calls,omitempty"`
	Usage     Usage                  `json:"usage,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ToolCall represents a tool call made by the LLM.
type ToolCall struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Args     map[string]interface{} `json:"args"`
}

// Usage represents token usage information.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Client defines the interface for LLM interactions.
// Users implement this interface with their preferred LLM provider.
type Client interface {
	// Complete performs a completion request and returns the response.
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	
	// Close cleans up any resources used by the client.
	Close() error
}