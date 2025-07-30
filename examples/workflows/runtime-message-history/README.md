# Runtime Message History Example

This example demonstrates how to load and manage message history at runtime using the WorkContext, rather than setting it at compile time when creating the agent.

## Key Features

- **Runtime History Loading**: Load conversation history from external sources (database, session storage, etc.)
- **Dynamic Updates**: Update message history between agent calls
- **Session Management**: Simulate real conversation sessions with evolving context
- **Flexible Integration**: Works with any external storage or session management system

## Usage Patterns

### 1. Runtime History Loading
```go
// Create agent without compile-time history
agent := agent.NewChatAgent("assistant").WithClient(client)

// Load history at runtime (from database, API, etc.)
history := loadConversationFromDatabase(sessionID)

// Set history in context
ctx := workflow.NewWorkContext(context.Background())
ctx.Set("message_history", history)
ctx.Set("user_input", "Continue our conversation...")

// Agent will use the runtime-loaded history
report := agent.Run(ctx)
```

### 2. Session Management
```go
// Evolving conversation session
sessionHistory := []llm.Message{}

for userInput := range conversationInputs {
    ctx := workflow.NewWorkContext(context.Background())
    ctx.Set("message_history", sessionHistory)
    ctx.Set("user_input", userInput)
    
    report := agent.Run(ctx)
    
    // Update session history
    sessionHistory = append(sessionHistory,
        llm.Message{Role: "user", Content: userInput},
        llm.Message{Role: "assistant", Content: response.Content},
    )
}
```

## Running the Example

```bash
# Set your OpenAI API key
export OPENAI_API_KEY=your-api-key-here

# Run the example
go run main.go
```

## Examples Demonstrated

1. **Runtime History Loading**: Shows how to load conversation context at runtime
2. **Conversation Session**: Simulates a real session with evolving history
3. **Dynamic Updates**: Demonstrates building history across multiple interactions

## Priority Order

The agents check for message history in this order:
1. **Runtime history** from `ctx.Get("message_history")` (highest priority)
2. **Compile-time history** from `WithMessageHistory()` (fallback)

This allows maximum flexibility - you can set a default history at creation time and override it at runtime when needed.

## Use Cases

- **Web Applications**: Load user conversation history from session storage
- **Chat Bots**: Retrieve conversation context from database
- **API Services**: Pass conversation history via request parameters
- **Multi-turn Workflows**: Build up conversation context across workflow steps
- **Session Management**: Maintain conversation state across multiple API calls

## Best Practices

1. **Memory Management**: Limit history size to avoid token limits
2. **Persistence**: Save conversation history to external storage
3. **Error Handling**: Handle cases where history loading fails
4. **Performance**: Cache frequently accessed conversation histories
5. **Security**: Validate and sanitize message history from external sources