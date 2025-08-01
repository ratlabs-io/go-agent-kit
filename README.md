# Go Agent Kit

[![Go Reference](https://pkg.go.dev/badge/github.com/ratlabs-io/go-agent-kit.svg)](https://pkg.go.dev/github.com/ratlabs-io/go-agent-kit)
[![Go Report Card](https://goreportcard.com/badge/github.com/ratlabs-io/go-agent-kit)](https://goreportcard.com/report/github.com/ratlabs-io/go-agent-kit)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A lightweight, composable framework for building agent workflows in Go. Go Agent Kit provides a clean architecture where **workflows orchestrate** and **agents participate as nodes**, making it easy to build complex AI-driven applications.

## 🚀 Key Features

- **🔀 Multi-Provider Router**: Built-in router for seamless switching between LLM providers (OpenAI, Anthropic, Grok, Gemini)
- **🔧 Zero Dependencies**: Core library has no external dependencies  
- **🤖 Bring Your Own LLM**: Generic interface supports any LLM provider
- **🛠️ Flexible Tool System**: Simple native tools with full schema control
- **📋 Structured JSON Responses**: Support for JSON schemas and type-safe outputs
- **⚙️ Smart Defaults**: Sensible defaults (4000 tokens, 0.7 temp, 0.95 top-p) with full customization
- **⚡ Composable Workflows**: Sequential, parallel, conditional, switch, loop, retry, and error handling patterns
- **📦 Production Ready**: Clean architecture, comprehensive error handling, and structured logging
- **🔄 Event System**: Callback-based monitoring and metrics collection
- **💬 Message History**: Built-in support for conversation context and chat history

## 📖 Quick Start

### Installation

```bash
go get github.com/ratlabs-io/go-agent-kit
```

### Basic Chat Agent

```go
package main

import (
    "context"
    "fmt"
    "os"
    
    "github.com/ratlabs-io/go-agent-kit/examples/integrations/clients"
    "github.com/ratlabs-io/go-agent-kit/pkg/agent"
    "github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

func main() {
    // Create LLM client (you can use any provider)
    llmClient := clients.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
    
    // Create agent
    chatAgent := agent.NewChatAgent("assistant").
        WithModel("gpt-4o-mini").
        WithPrompt("You are a helpful assistant.").
        WithClient(llmClient)
    
    // Create workflow context
    ctx := workflow.NewWorkContext(context.Background())
    ctx.Set("user_input", "What is the capital of France?")
    
    // Run agent
    report := chatAgent.Run(ctx)
    if report.Status == workflow.StatusCompleted {
        fmt.Printf("Response: %v\n", report.Data)
    }
}
```

### Agent Parameters

All agents come with sensible defaults for generation parameters, but you can customize them:

```go
// Agents use these defaults:
// - MaxTokens: 4000
// - Temperature: 0.7 
// - TopP: 0.95

// Using defaults (recommended for most cases)
chatAgent := agent.NewChatAgent("assistant").
    WithModel("gpt-4o-mini").
    WithPrompt("You are a helpful assistant.").
    WithClient(llmClient)

// Custom parameters for specific needs
chatAgent := agent.NewChatAgent("assistant").
    WithModel("gpt-4o-mini").
    WithPrompt("You are a creative writer.").
    WithMaxTokens(2000).      // Limit output length
    WithTemperature(0.9).     // More creative responses
    WithTopP(0.8).            // Adjust nucleus sampling
    WithClient(llmClient)

// Works with both ChatAgent and ToolAgent
toolAgent := agent.NewToolAgent("assistant").
    WithModel("gpt-4o-mini").
    WithPrompt("You are a helpful assistant.").
    WithMaxTokens(1500).
    WithTemperature(0.3).     // More deterministic for tool use
    WithTopP(0.95).
    WithTools(mathTool).
    WithClient(llmClient)
```

### Chat Agent with Message History

```go
// Create agent (agents are stateless)
chatAgent := agent.NewChatAgent("assistant").
    WithModel("gpt-4o-mini").
    WithPrompt("You are a helpful assistant.").
    WithClient(llmClient)

// Load conversation history at runtime (from database, session, API, etc.)
conversationHistory := loadConversationFromDatabase(sessionID)

ctx := workflow.NewWorkContext(context.Background())
ctx.Set(constants.KeyMessageHistory, conversationHistory) // Load history at runtime
ctx.Set(constants.KeyUserInput, "Continue our conversation...")

// The agent will use the runtime-loaded conversation context
report := chatAgent.Run(ctx)
```

### Sequential Workflow

```go
// Create specialized agents
researchAgent := agent.NewChatAgent("researcher").
    WithModel("gpt-4o").
    WithPrompt("Research the topic and gather key facts.").
    WithClient(llmClient)

summaryAgent := agent.NewChatAgent("summarizer").
    WithModel("gpt-4o-mini").
    WithPrompt("Create a concise summary of the research.").
    WithClient(llmClient)

analyzerAgent := agent.NewChatAgent("analyzer").
    WithModel("gpt-4o").
    WithPrompt("Analyze the research and provide insights.").
    WithClient(llmClient)

// Chain them in sequence with output passing
pipeline := workflow.NewSequentialFlow("research-pipeline").
    Then(researchAgent).
    ThenChain(summaryAgent).    // Gets previous output as input
    ThenChain(analyzerAgent)    // Gets summary as input

// Execute workflow
ctx := workflow.NewWorkContext(context.Background())
ctx.Set(constants.KeyUserInput, "Benefits of renewable energy")
report := pipeline.Run(ctx)
```

### Tool-Enabled Agent

```go
import "github.com/ratlabs-io/go-agent-kit/examples/tools"

// Create tools
mathTool := tools.NewMathTool()
echoTool := tools.NewEchoTool()

// Create tool agent
toolAgent := agent.NewToolAgent("calculator").
    WithModel("gpt-3.5-turbo").
    WithPrompt("You can use tools to help users.").
    WithClient(llmClient).
    WithTools(mathTool, echoTool)

ctx := workflow.NewWorkContext(context.Background())
ctx.Set(constants.KeyUserInput, "Calculate 15 * 23 and echo the result")
report := toolAgent.Run(ctx)
```

### Tool Agent with Message History

```go
// Create tool agent (stateless)
toolAgent := agent.NewToolAgent("assistant").
    WithModel("gpt-3.5-turbo").
    WithPrompt("You are a helpful assistant that can use tools.").
    WithClient(llmClient).
    WithTools(mathTool, echoTool)

// Load conversation history at runtime
conversationHistory := []llm.Message{
    {Role: "user", Content: "Calculate 10 + 20"},
    {Role: "assistant", Content: "I'll calculate that for you. 10 + 20 = 30"},
}

ctx := workflow.NewWorkContext(context.Background())
ctx.Set(constants.KeyMessageHistory, conversationHistory) // Load at runtime
ctx.Set(constants.KeyUserInput, "Now multiply that result by 5")

// Tool agent will use tools with full conversation context
report := toolAgent.Run(ctx)
```

### Structured JSON Responses

Get predictable, parseable responses with JSON schemas:

```go
import "encoding/json"

// Define a JSON schema for structured output
schema := &llm.JSONSchema{
    Name: "task_analysis",
    Description: "Analysis of a user task",
    Schema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "category": map[string]interface{}{
                "type": "string",
                "enum": []string{"question", "request", "task", "other"},
            },
            "complexity": map[string]interface{}{
                "type": "string",
                "enum": []string{"low", "medium", "high"},
            },
            "estimated_time": map[string]interface{}{
                "type": "integer",
                "description": "Estimated time in minutes",
            },
            "requirements": map[string]interface{}{
                "type": "array",
                "items": map[string]interface{}{"type": "string"},
            },
        },
        "required": []string{"category", "complexity", "estimated_time"},
        "additionalProperties": false,
    },
    Strict: true,
}

// Create agent with JSON schema
// Note: Use gpt-4o-mini or gpt-4o for JSON schema support
jsonAgent := agent.NewChatAgent("analyzer").
    WithModel("gpt-4o-mini").
    WithPrompt("Analyze the user's input and provide structured analysis.").
    WithJSONSchema(schema).
    WithClient(llmClient)

ctx := workflow.NewWorkContext(context.Background())
ctx.Set(constants.KeyUserInput, "Help me plan a vacation to Japan")
report := jsonAgent.Run(ctx)

// Response will be valid JSON matching the schema
if report.Status == workflow.StatusCompleted {
    if response, ok := report.Data.(*llm.CompletionResponse); ok {
        // Parse JSON response
        var analysis map[string]interface{}
        json.Unmarshal([]byte(response.Content), &analysis)
        fmt.Printf("Category: %s\n", analysis["category"])
        fmt.Printf("Complexity: %s\n", analysis["complexity"])
    }
}
```

You can also request generic JSON objects without a specific schema:

```go
// Request JSON without specific schema
jsonAgent := agent.NewChatAgent("json-responder").
    WithModel("gpt-4").
    WithPrompt("Respond with a JSON object containing key insights.").
    WithJSONResponse().
    WithClient(llmClient)
```

## 🔀 Multi-Provider Router

For advanced use cases, you can use the built-in router to seamlessly switch between different LLM providers:

```go
package main

import (
    "context"
    "fmt"
    "os"
    
    "github.com/ratlabs-io/go-agent-kit/examples/integrations/clients"
    "github.com/ratlabs-io/go-agent-kit/pkg/agent"
    "github.com/ratlabs-io/go-agent-kit/pkg/llm"
    "github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

func main() {
    // Set up the router once
    router := llm.NewRouterClient()
    
    // Register your preferred LLM providers
    if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
        router.Register("openai", clients.NewOpenAIClient(apiKey))
    }
    if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
        router.Register("anthropic", clients.NewAnthropicClient(apiKey))
    }
    if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
        router.Register("gemini", clients.NewGeminiClient(apiKey))
    }
    if apiKey := os.Getenv("GROK_API_KEY"); apiKey != "" {
        router.Register("grok", clients.NewGrokClient(apiKey))
    }
    
    // Create agents with automatic provider routing
    researchAgent := agent.NewChatAgent("researcher").
        WithModel("openai/gpt-4o").              // Routes to OpenAI
        WithPrompt("Research the topic and gather key facts.").
        WithClient(router)

    summaryAgent := agent.NewChatAgent("summarizer").
        WithModel("anthropic/claude-3-haiku-20240307").  // Routes to Anthropic
        WithPrompt("Create a concise summary of the research.").
        WithClient(router)

    analyzerAgent := agent.NewChatAgent("analyzer").
        WithModel("gemini/gemini-1.5-pro").      // Routes to Gemini
        WithPrompt("Analyze the research and provide insights.").
        WithClient(router)
    
    // Use different providers in the same workflow
    pipeline := workflow.NewSequentialFlow("multi-provider-pipeline").
        Then(researchAgent).
        ThenChain(summaryAgent).
        ThenChain(analyzerAgent)
    
    ctx := workflow.NewWorkContext(context.Background())
    ctx.Set("user_input", "Benefits of renewable energy")
    report := pipeline.Run(ctx)
}
```

### Router Benefits

- **Provider Flexibility**: Switch providers without changing agent code
- **Cost Optimization**: Use different models for different tasks (fast models for simple tasks, powerful models for complex analysis)
- **Redundancy**: Fallback to different providers if one is unavailable
- **A/B Testing**: Easy comparison between different models and providers

## 🔄 Workflow Patterns

### Sequential Execution
Execute agents one after another:
```go
workflow.NewSequentialFlow("pipeline").
    Then(agent1).
    Then(agent2).
    Then(agent3)
```

### Parallel Execution  
Execute agents concurrently for speed:
```go
workflow.NewParallelFlow("concurrent").
    Execute(analyst1).
    Execute(analyst2).
    Execute(analyst3)
```

### Conditional Execution
Branch based on conditions:
```go
isQuestion := func(ctx *workflow.WorkContext) (bool, error) {
    // Your condition logic
    return true, nil
}

workflow.NewConditionalFlow("question-handler", isQuestion, questionAgent, fallbackAgent)
```

### Switch-Based Routing
Multiple conditions with priority:
```go
router := workflow.NewSwitchFlowBuilder("content-router").
    Case(isUrgent, urgentAgent).
    Case(isTechnical, techAgent).
    Case(isCreative, creativeAgent).
    Default(generalAgent).
    Build()
```

### Loop Constructs
Execute repeated operations with different loop patterns:
```go
// Count-based loop
countLoop := workflow.NewLoop("process-batch", 5).
    WithAction(processingAgent)

// While condition loop
whileLoop := workflow.NewLoopWhile("accumulate-data", func(ctx workflow.WorkContext) (bool, error) {
    total, _ := ctx.Get("total")
    return total.(int) < 100, nil
}).WithAction(accumulatorAgent)

// Until condition loop  
untilLoop := workflow.NewLoopUntil("retry-until-success", func(ctx workflow.WorkContext) (bool, error) {
    success, _ := ctx.Get("success")
    return success.(bool), nil
}).WithAction(retryAgent)

// Iterator loop over collections
items := []string{"item1", "item2", "item3"}
iterLoop := workflow.NewLoopOver("process-items", items).
    WithAction(itemProcessor)
```

### Retry Patterns
Handle failures with sophisticated retry strategies:
```go
// Basic retry with exponential backoff
retryAction := workflow.NewRetry("api-call", 3).
    WithAction(apiAgent).
    WithBackoffStrategy(workflow.NewExponentialBackoff(100*time.Millisecond, 5*time.Second, 2.0)).
    WithRetryCondition(workflow.RetryOnTimeoutCondition)

// Retry with custom conditions
retryAction := workflow.NewRetry("service-call", 5).
    WithAction(serviceAgent).
    WithBackoffStrategy(workflow.NewLinearBackoff(500*time.Millisecond, 200*time.Millisecond)).
    WithRetryCondition(workflow.CombineRetryConditions(
        workflow.RetryOnTimeoutCondition,
        workflow.RetryOnNetworkCondition,
    )).
    WithStopCondition(func(ctx workflow.WorkContext, err error, attempt int) bool {
        // Stop retrying on validation errors
        return strings.Contains(err.Error(), "validation")
    })
```

### Error Handling
Structured error handling with try-catch-finally patterns:
```go
// Try-catch with specific error handlers
tryCatchFlow := workflow.NewTryCatch("payment-processing").
    WithTryAction(paymentAgent).
    Catch(workflow.TimeoutError, timeoutHandler).
    Catch(workflow.NetworkError, networkHandler).
    Catch(workflow.ErrorMessageContains("insufficient funds"), fundsHandler).
    CatchAny(generalErrorHandler).
    Finally(cleanupAction)

// Error matchers for specific error types
validationHandler := workflow.NewDefaultErrorHandlerAction("validation-handler", 
    func(ctx workflow.WorkContext, err error) workflow.WorkReport {
        // Handle validation errors
        return workflow.NewCompletedWorkReport()
    })

tryCatch := workflow.NewTryCatch("data-processing").
    WithTryAction(dataProcessor).
    Catch(workflow.ValidationError, validationHandler)
```

### Advanced Reliability Patterns
Circuit breakers, timeouts, and parallel error collection:
```go
// Circuit breaker protection
circuitBreaker := workflow.NewCircuitBreaker("service-breaker", 3, 2*time.Second, 5*time.Second).
    WithAction(unreliableService)

// Timeout wrapper
timeoutWrapper := workflow.NewTimeoutWrapper("api-timeout", 30*time.Second).
    WithAction(slowAPICall)

// Parallel error collector for comprehensive monitoring
errorCollector := workflow.NewParallelErrorCollector("health-check").
    AddActions(databaseCheck, cacheCheck, apiCheck)

// Combine multiple reliability patterns
protectedService := workflow.NewRetry("protected-call", 3).
    WithAction(
        workflow.NewCircuitBreaker("circuit", 2, 1*time.Second, 3*time.Second).
            WithAction(
                workflow.NewTimeoutWrapper("timeout", 5*time.Second).
                    WithAction(externalService),
            ),
    ).
    WithBackoffStrategy(workflow.NewExponentialBackoff(100*time.Millisecond, 2*time.Second, 2.0))
```

### Workflow Composition
Nest workflows within workflows:
```go
subWorkflow := workflow.NewParallelFlow("analysis").
    Execute(techAnalyst).
    Execute(marketAnalyst)

mainWorkflow := workflow.NewSequentialFlow("main").
    Then(classifier).
    Then(subWorkflow).
    Then(synthesizer)
```

## 🛠️ Tool Development

### Simple Tools (Recommended)

```go
type MathTool struct{}

func (t *MathTool) Name() string { 
    return "math" 
}

func (t *MathTool) Description() string { 
    return "Performs basic math operations" 
}

func (t *MathTool) Parameters() tools.ToolParameterSchema {
    return tools.ToolParameterSchema{
        Type: "object",
        Properties: map[string]tools.ToolParameter{
            "operation": {Type: "string", Description: "Math operation (+, -, *, /)"},
            "a": {Type: "number", Description: "First number"},
            "b": {Type: "number", Description: "Second number"},
        },
        Required: []string{"operation", "a", "b"},
    }
}

func (t *MathTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // Your tool logic here
    op := params["operation"].(string)
    a := params["a"].(float64)
    b := params["b"].(float64)
    
    switch op {
    case "+":
        return a + b, nil
    case "-":
        return a - b, nil
    case "*":
        return a * b, nil
    case "/":
        if b == 0 {
            return nil, fmt.Errorf("division by zero")
        }
        return a / b, nil
    default:
        return nil, fmt.Errorf("unknown operation: %s", op)
    }
}
```

### Simple Tool Interface (Less Code)

For simpler tools, use the `SimpleTool` interface:

```go
type EchoTool struct{}

func (t *EchoTool) Name() string { return "echo" }
func (t *EchoTool) Description() string { return "Echoes the input text" }

func (t *EchoTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    text, ok := params["text"].(string)
    if !ok {
        return nil, fmt.Errorf("text parameter required")
    }
    return map[string]interface{}{"echo": text}, nil
}

// Wrap for use with agents
agent.WithTools(tools.WrapSimpleTool(&EchoTool{}))
```

## 💻 LLM Integration

### Built-in Integrations

The library includes example integrations in `examples/integrations/`:

**OpenAI (Zero Dependencies)**
```go
import "github.com/ratlabs-io/go-agent-kit/examples/integrations/openai"

client := openai.NewClient("your-api-key")
// Supports: OpenAI, Azure OpenAI, and compatible APIs
```

**GoLLM Wrapper (External Dependency)**
```go
// First: go get github.com/teilomillet/gollm
import "github.com/ratlabs-io/go-agent-kit/examples/integrations/gollm"

client, err := gollm.NewClient("openai", "gpt-4")
// Supports: Multiple providers through gollm
```

### Custom LLM Integration

Implement the `llm.Client` interface:

```go
import "github.com/ratlabs-io/go-agent-kit/pkg/llm"

type MyLLMClient struct {
    apiKey string
    baseURL string
}

func (c *MyLLMClient) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
    // Your LLM API integration
    // Handle req.Messages, req.Tools, req.MaxTokens, req.Temperature, req.TopP, etc.
    return &llm.CompletionResponse{
        Content: "response text",
        Usage: llm.TokenUsage{
            PromptTokens:     10,
            CompletionTokens: 20,
            TotalTokens:      30,
        },
    }, nil
}

func (c *MyLLMClient) Close() error {
    // Cleanup resources if needed
    return nil
}
```

## 📊 Event Monitoring

Monitor agent execution with callbacks:

```go
// Create callback registry
callbacks := workflow.NewCallbackRegistry()

// Add monitoring callbacks
callbacks.Add(func(ctx context.Context, event workflow.Event) {
    if event.Type == workflow.EventAgentCompleted {
        fmt.Printf("Agent '%s' completed\n", event.Source)
        
        // Extract token usage
        if response, ok := event.Payload.(*llm.CompletionResponse); ok {
            fmt.Printf("Tokens used: %d\n", response.Usage.TotalTokens)
        }
    }
})

// Create work context with callbacks
workCtx := workflow.NewWorkContextWithCallbacks(context.Background(), callbacks)
workCtx.Set(constants.KeyUserInput, "Your question here")

// Run agent - events will be emitted to callbacks
report := chatAgent.Run(workCtx)
```

## 🧪 Examples

Explore comprehensive examples in [`examples/`](./examples/):

### Router & Integrations
- **[llm-router](./examples/llm-router/)**: Multi-provider LLM routing with OpenAI, Anthropic, Grok, Gemini
- **[integrations/clients](./examples/integrations/clients/)**: Individual LLM provider implementations

### Workflow Examples
- **[simple-agent](./examples/workflows/simple-agent/)**: Basic chat completion
- **[simple-agent-with-callbacks](./examples/workflows/simple-agent-with-callbacks/)**: Event monitoring
- **[structured-json-agent](./examples/workflows/structured-json-agent/)**: JSON schema responses
- **[chat-with-history](./examples/workflows/chat-with-history/)**: Conversation context and message history
- **[runtime-message-history](./examples/workflows/runtime-message-history/)**: Runtime message history loading
- **[sequential-workflow](./examples/workflows/sequential-workflow/)**: Multi-step processing  
- **[chaining-patterns](./examples/workflows/chaining-patterns/)**: Different sequential chaining strategies
- **[parallel-workflow](./examples/workflows/parallel-workflow/)**: Concurrent execution
- **[conditional-workflow](./examples/workflows/conditional-workflow/)**: Branching logic
- **[switch-workflow](./examples/workflows/switch-workflow/)**: Priority-based routing
- **[loop-constructs](./examples/workflows/loop-constructs/)**: Loop patterns (count, while, until, iterator)
- **[retry-patterns](./examples/workflows/retry-patterns/)**: Retry strategies with backoff
- **[error-handling](./examples/workflows/error-handling/)**: Try-catch-finally patterns
- **[advanced-constructs](./examples/workflows/advanced-constructs/)**: Circuit breakers, timeouts, parallel error collection
- **[comprehensive-example](./examples/workflows/comprehensive-example/)**: Real-world e-commerce order processing using all constructs
- **[tool-agent](./examples/workflows/tool-agent/)**: Tool-calling scenarios
- **[action-func](./examples/workflows/action-func/)**: Custom actions without boilerplate

### Running Examples

```bash
# Set your API keys (one or more)
export OPENAI_API_KEY=your-openai-key
export ANTHROPIC_API_KEY=your-anthropic-key
export GEMINI_API_KEY=your-gemini-key
export GROK_API_KEY=your-grok-key

# Try the multi-provider router
go run examples/llm-router/main.go

# Run any workflow example
go run examples/workflows/simple-agent/main.go
go run examples/workflows/structured-json-agent/main.go
go run examples/workflows/chat-with-history/main.go
go run examples/workflows/runtime-message-history/main.go
go run examples/workflows/sequential-workflow/main.go
go run examples/workflows/tool-agent/main.go

# Run the new reliability and control flow examples
go run examples/workflows/loop-constructs/main.go
go run examples/workflows/retry-patterns/main.go
go run examples/workflows/error-handling/main.go
go run examples/workflows/advanced-constructs/main.go
go run examples/workflows/comprehensive-example/main.go
```

## 🏗️ Architecture

### Core Philosophy

```
🔄 Workflows (orchestrate) 
   └── 🤖 Agents (participate as nodes)
       └── 🛠️ Tools (extend capabilities)
```

### Project Structure

```
go-agent-kit/
├── pkg/                    # 🏗️ Core library (zero dependencies)
│   ├── workflow/           # Workflow orchestration
│   │   ├── action.go       # Base action interface
│   │   ├── sequential.go   # Sequential execution
│   │   ├── parallel.go     # Parallel execution
│   │   ├── conditional.go  # Conditional execution
│   │   ├── switch.go       # Switch-based routing
│   │   ├── loop.go         # Loop constructs (count, while, until, iterator)
│   │   ├── retry.go        # Retry patterns with backoff strategies
│   │   ├── trycatch.go     # Try-catch-finally error handling
│   │   ├── advanced.go     # Circuit breakers, timeouts, parallel error collection
│   │   ├── context.go      # Shared execution context
│   │   └── callbacks.go    # Event callback system
│   ├── agent/              # Agent implementations
│   │   ├── chat_agent.go   # Simple LLM completion agent
│   │   └── tool_agent.go   # Tool-calling agent
│   ├── tools/              # Tool system
│   │   ├── tool.go         # Tool interfaces
│   │   └── registry.go     # Tool management
│   └── llm/                # LLM abstraction
│       ├── client.go       # Generic LLM interface
│       └── types.go        # Request/response types
├── examples/               # 📚 Examples and integrations
│   ├── workflows/          # Complete workflow examples
│   ├── tools/              # Reference tool implementations  
│   └── integrations/       # LLM provider integrations
│       ├── openai/         # OpenAI client (zero deps)
│       └── gollm/          # GoLLM wrapper (external dep)
└── go.mod                  # Module definition (zero deps)
```

### Design Principles

1. **Zero Dependencies**: Core library is completely self-contained
2. **Composability**: All components can be nested and combined arbitrarily
3. **Bring Your Own**: Generic interfaces for LLMs, tools, and custom logic
4. **Production Ready**: Comprehensive error handling, logging, and monitoring
5. **Clean Architecture**: Clear separation between orchestration and execution

## 🤝 Contributing

We welcome contributions! Areas where you can help:

- **LLM Integrations**: Add support for new LLM providers
- **Tool Implementations**: Create useful reference tools
- **Workflow Patterns**: Add new execution patterns (we just added loops, retry, error handling!)
- **Reliability Patterns**: Enhance circuit breakers, timeouts, and fault tolerance
- **Documentation**: Improve examples and guides
- **Testing**: Expand test coverage

Please see our [contributing guidelines](CONTRIBUTING.md) for detailed information.

## 📄 License

Go Agent Kit is released under the MIT License. See [LICENSE](LICENSE) for details.

## 🚀 Getting Started

1. **Install**: `go get github.com/ratlabs-io/go-agent-kit`
2. **Explore**: Check out [`examples/workflows/`](./examples/workflows/)
3. **Build**: Create your first agent workflow
4. **Extend**: Add custom tools and LLM integrations

---

**Questions or Issues?**

- 📖 **Documentation**: Check the [examples directory](./examples/)
- 🐛 **Bug Reports**: [Create an issue](https://github.com/ratlabs-io/go-agent-kit/issues)
- 💡 **Feature Requests**: [Start a discussion](https://github.com/ratlabs-io/go-agent-kit/discussions)
- 💬 **Community**: [Join our Discord](https://discord.gg/your-server)

Built with ❤️ for the Go community.