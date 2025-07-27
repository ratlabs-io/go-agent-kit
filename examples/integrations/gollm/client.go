package gollm

import (
	"context"
	"fmt"

	"github.com/ratlabs-io/go-agent-kit/pkg/llm"
	"github.com/teilomillet/gollm"
)

// Client implements the llm.Client interface using the gollm library.
// This is an example implementation - users can use this or create their own.
// 
// To use this, add to your go.mod:
// require github.com/teilomillet/gollm v0.1.9
type Client struct {
	llm gollm.LLM
}

// NewClient creates a new gollm client with the specified provider and model.
func NewClient(provider, model string) (*Client, error) {
	llmInstance, err := gollm.NewLLM(
		gollm.SetProvider(provider),
		gollm.SetModel(model),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gollm instance: %w", err)
	}
	
	return &Client{
		llm: llmInstance,
	}, nil
}

// NewClientWithOptions creates a new gollm client with custom options.
func NewClientWithOptions(opts ...gollm.ConfigOption) (*Client, error) {
	llmInstance, err := gollm.NewLLM(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gollm instance: %w", err)
	}
	
	return &Client{
		llm: llmInstance,
	}, nil
}

// Complete implements llm.Client.Complete using gollm.
func (c *Client) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	// For now, we'll do a simple generation
	// TODO: Enhance with proper tool calling integration when gollm supports it
	
	var prompt string
	if len(req.Messages) > 0 {
		// Convert messages to a simple prompt
		// This is a basic implementation - gollm may have better message support
		for _, msg := range req.Messages {
			prompt += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
		}
	} else {
		prompt = req.Prompt
	}
	
	// Create gollm prompt
	gollmPrompt := c.llm.NewPrompt(prompt)
	
	response, err := c.llm.Generate(ctx, gollmPrompt)
	if err != nil {
		return nil, fmt.Errorf("gollm generation failed: %w", err)
	}
	
	return &llm.CompletionResponse{
		Content: response,
		Usage: llm.Usage{
			// TODO: Extract actual usage from gollm when available
			// For now, use a rough approximation
			PromptTokens:     estimateTokens(prompt),
			CompletionTokens: estimateTokens(response),
			TotalTokens:      estimateTokens(prompt) + estimateTokens(response),
		},
		Metadata: req.Metadata,
		// Note: Tool calls not yet implemented - depends on gollm's tool calling capabilities
	}, nil
}

// Close implements llm.Client.Close.
func (c *Client) Close() error {
	// gollm doesn't currently have a Close method
	// This is here for future compatibility
	return nil
}

// estimateTokens provides a rough token count estimation.
// This is a placeholder - real implementations should use proper tokenizers.
func estimateTokens(text string) int {
	// Very rough approximation: ~4 characters per token
	return len(text) / 4
}