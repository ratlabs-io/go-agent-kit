# Simple Agent with Callbacks Example

This example demonstrates how to use the callback system in go-agent-kit to monitor agent execution and gather metrics.

## Key Features

- **Event-driven monitoring**: Register callbacks to receive events when agents complete
- **Multiple callback types**: Different callbacks for different monitoring purposes
- **Token usage tracking**: Extract and display LLM token usage statistics
- **Execution timing**: Monitor agent execution time and performance
- **Event metadata**: Access rich metadata from agent execution events
- **Reliable completion**: Callbacks are guaranteed to complete before workflow exits

## Code Highlights

### Callback Registration
```go
callbacks := workflow.NewCallbackRegistry()

// Agent completion monitoring
callbacks.Add(func(ctx context.Context, event workflow.Event) {
    if event.Type == workflow.EventAgentCompleted {
        fmt.Printf("📊 Event: Agent '%s' completed\n", event.Source)
    }
})

// Token usage tracking
callbacks.Add(func(ctx context.Context, event workflow.Event) {
    if event.Type == workflow.EventAgentCompleted {
        if response, ok := event.Payload.(*llm.CompletionResponse); ok {
            fmt.Printf("🎯 Token usage - Prompt: %d, Completion: %d, Total: %d\n", 
                response.Usage.PromptTokens, 
                response.Usage.CompletionTokens, 
                response.Usage.TotalTokens)
        }
    }
})
```

### WorkContext with Callbacks Usage
```go
// Create WorkContext with callbacks (instead of plain WorkContext)
workCtx := workflow.NewWorkContextWithCallbacks(ctx, callbacks)
workCtx.Set("user_input", "What are the three largest cities in Japan?")

// Run agent - events will be automatically emitted to callbacks
report := chatAgent.Run(workCtx)
```

## Event Types

The callback system supports various event types:

- `workflow.EventAgentCompleted` - Agent finished successfully
- `workflow.EventAgentFailed` - Agent execution failed
- `workflow.EventAgentStarted` - Agent execution started
- `workflow.EventToolCalled` - Tool was invoked
- Additional workflow and tool events

## Use Cases

This pattern is useful for:

- **Performance monitoring** - Track execution times and identify bottlenecks
- **Cost tracking** - Monitor token usage for billing and optimization
- **Debugging** - Log agent execution flow and identify issues
- **Analytics** - Gather statistics on agent usage patterns
- **Alerting** - Trigger notifications on failures or performance issues

## Running the Example

```bash
# Set your OpenAI API key
export OPENAI_API_KEY=your-api-key-here

# Run the example
go run examples/workflows/simple-agent-with-callbacks/main.go
```

## Expected Output

```
=== Simple Agent with Callbacks Example ===
🚀 Running chat agent with callback monitoring...
📝 Question: What are the three largest cities in Japan?

📊 Event: Agent 'assistant' completed
🔔 [19:59:18] Event: agent.completed from assistant
⏱️  Execution time: 1.130746s
🤖 Agent type: chat
🎯 Token usage - Prompt: 34, Completion: 16, Total: 50

⏰ Total execution time: 1.13098575s
✅ Agent completed successfully!
💬 Response: The three largest cities in Japan are Tokyo, Yokohama, and Osaka.
```

This demonstrates the full callback system in action, showing how callbacks can provide real-time monitoring and metrics collection for agent workflows.

## Implementation Details

### Callback Execution
- Callbacks are executed **concurrently** in separate goroutines for performance
- Each callback includes panic recovery to prevent crashes
- The framework uses a `sync.WaitGroup` to track running callbacks
- Agents automatically wait for all callbacks to complete before returning

### Synchronous vs Asynchronous
- `EmitEvent()` - Runs callbacks concurrently (default, non-blocking)
- `EmitEventSync()` - Runs callbacks sequentially (blocking)

This ensures that critical operations like logging, metrics collection, and notifications complete reliably without requiring manual synchronization in user code.