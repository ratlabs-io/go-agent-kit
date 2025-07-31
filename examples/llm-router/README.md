# LLM Router Example

This example demonstrates how to use the new LLM router pattern in go-agent-kit.

## Overview

The router pattern allows you to:
1. Register multiple LLM providers with a single router
2. Use a unified interface to switch between providers
3. Pass the router to agents and workflows for automatic provider routing

## Usage Pattern

```go
// Set up the router once
router := llm.NewRouterClient()
router.Register("openai", openaiClient)
router.Register("anthropic", anthropicClient)

// Pass it to your workflow/agent
chatAgent := agent.NewChatAgent("assistant").
    WithLLMClient(router).
    WithModel("openai/gpt-4o-mini") // Provider is automatically routed
```

## Running the Example

1. Set your API keys:
```bash
export OPENAI_API_KEY="your-openai-key"
export ANTHROPIC_API_KEY="your-anthropic-key"
export GROK_API_KEY="your-grok-key"
export GEMINI_API_KEY="your-gemini-key"
```

2. Run the example:
```bash
go run examples/llm-router/main.go
```

## What it demonstrates

1. **Router Registration**: How to register multiple LLM clients with the router
2. **Direct Usage**: Using the router directly for completions
3. **Agent Integration**: Using the router with go-agent-kit agents and workflows

## Model Format

All models must use the `provider/model` format:
- `openai/gpt-4o-mini`
- `anthropic/claude-3-haiku-20240307`
- `grok/grok-beta`
- `gemini/gemini-1.5-flash`

The router automatically strips the provider prefix and routes to the appropriate client.