# LLM Client Implementations

This directory contains concrete implementations of the `llm.Client` interface for various LLM providers.

## Available Clients

### OpenAI Client (`openai.go`)
- Supports GPT models via OpenAI's chat completions API
- Features: Tool calling, JSON schema responses, streaming (future)
- Models: `gpt-4o`, `gpt-4o-mini`, `o1-preview`, etc.

### Anthropic Client (`anthropic.go`)  
- Supports Claude models via Anthropic's messages API
- Features: Tool calling, system prompts, JSON responses via prompt engineering
- Models: `claude-3-opus-20240229`, `claude-3-haiku-20240307`, etc.

### Grok Client (`grok.go`)
- Supports xAI's Grok models via OpenAI-compatible API
- Features: Tool calling, JSON schema responses
- Models: `grok-beta`, `grok-2`, etc.

### Gemini Client (`gemini.go`)
- Supports Google's Gemini models via Generative AI API  
- Features: Tool calling, native JSON schema support, function calling
- Models: `gemini-1.5-pro`, `gemini-1.5-flash`, etc.

### Gollm Client (`gollm.go`)
- Supports multiple providers via the gollm library
- Features: Provider abstraction, multi-model support
- Note: Requires `github.com/teilomillet/gollm v0.1.9` dependency

## Usage with Router

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/ratlabs-io/go-agent-kit/pkg/llm"
    "github.com/ratlabs-io/go-agent-kit/examples/integrations/clients"
)

func main() {
    // Create router
    router := llm.NewRouterClient()

    // Register clients with their API keys
    if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
        openaiClient := clients.NewOpenAIClient(apiKey)
        router.Register("openai", openaiClient)
    }

    if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
        anthropicClient := clients.NewAnthropicClient(apiKey)
        router.Register("anthropic", anthropicClient)
    }

    if apiKey := os.Getenv("GROK_API_KEY"); apiKey != "" {
        grokClient := clients.NewGrokClient(apiKey)
        router.Register("grok", grokClient)
    }

    if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
        geminiClient := clients.NewGeminiClient(apiKey)
        router.Register("gemini", geminiClient)
    }

    // Or use gollm for multi-provider support
    gollmClient, err := clients.NewGollmClient("openai", "gpt-4o-mini")
    if err == nil {
        router.Register("gollm-openai", gollmClient)
    }

    // Use with provider/model format
    response, err := router.Complete(context.Background(), llm.CompletionRequest{
        Model: "openai/gpt-4o-mini", // Routes to OpenAI client
        Messages: []llm.Message{
            {Role: "user", Content: "Hello, world!"},
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Response: %s", response.Content)
}
```

## Environment Variables

Set the appropriate API keys:

```bash
export OPENAI_API_KEY="your-openai-key"
export ANTHROPIC_API_KEY="your-anthropic-key"  
export GROK_API_KEY="your-grok-key"
export GEMINI_API_KEY="your-gemini-key"
```

## Model Format

All requests must use the `provider/model` format:

- `openai/gpt-4o-mini`
- `anthropic/claude-3-haiku-20240307`
- `grok/grok-beta`
- `gemini/gemini-1.5-flash`

The router will strip the provider prefix and pass the bare model name to the appropriate client.