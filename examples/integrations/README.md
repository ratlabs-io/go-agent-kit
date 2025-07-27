# LLM Client Integrations

This directory contains example implementations of the `llm.Client` interface for popular LLM providers. These are **optional** implementations that users can import if they want, or use as reference for creating their own.

## Available Integrations

### OpenAI (`./openai/`)
- Pure Go implementation using OpenAI's HTTP API
- Supports chat completions and tool calling
- Zero dependencies beyond Go stdlib
- Works with OpenAI, Azure OpenAI, and compatible APIs

**Usage:**
```go
import "github.com/ratlabs-io/go-agent-kit/examples/integrations/openai"

client := openai.NewClient("your-api-key")
// or for Azure/custom endpoints:
client := openai.NewClientWithURL("your-api-key", "https://your-endpoint.com/v1")
```

### GoLLM (`./gollm/`)
- Wrapper around the `github.com/teilomillet/gollm` library
- Supports multiple providers (OpenAI, Anthropic, etc.)
- Requires adding gollm dependency to your project

**Usage:**
```go
// Add to your go.mod:
// require github.com/teilomillet/gollm v0.0.0-20241223144942-f730d7d49f95

import "github.com/ratlabs-io/go-agent-kit/examples/integrations/gollm"

client, err := gollm.NewClient("openai", "gpt-4")
// or with custom options:
client, err := gollm.NewClientWithOptions(
    gollm.SetProvider("openai"),
    gollm.SetModel("gpt-4"),
    gollm.SetAPIKey("your-key"),
)
```

## Creating Your Own Integration

Implement the `llm.Client` interface from `pkg/llm/client.go`:

```go
type Client interface {
    Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
    Close() error
}
```

### Example: Custom Client
```go
package myclient

import (
    "context"
    "github.com/ratlabs-io/go-agent-kit/pkg/llm"
)

type MyClient struct {
    // your client fields
}

func (c *MyClient) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
    // your implementation
}

func (c *MyClient) Close() error {
    // cleanup resources
}
```

## Integration Guidelines

1. **Keep it simple**: Focus on implementing the interface cleanly
2. **Handle errors gracefully**: Return descriptive error messages
3. **Support context cancellation**: Respect the provided context
4. **Document dependencies**: Clearly state what external libraries are needed
5. **Provide examples**: Show how to use your integration

## Note

These integrations are **examples only** - the core go-agent-kit library has zero external dependencies. Users are free to:
- Use these example integrations
- Modify them for their needs  
- Create entirely custom implementations
- Use any LLM library or API they prefer

The goal is maximum flexibility while providing helpful starting points.