# Chat Agent with Message History Example

This example demonstrates how to use runtime message history with both ChatAgent and ToolAgent to maintain conversation context across multiple interactions.

## Features Demonstrated

1. **Basic Chat with History**: Shows how to load previous conversation at runtime and continue naturally
2. **Tool Agent with History**: Demonstrates tool usage with conversation context
3. **Runtime History Loading**: Shows loading conversation context from external sources
4. **Multi-turn Conversation**: Shows how to build up history dynamically across multiple agent calls

## Running the Example

```bash
# Set your OpenAI API key
export OPENAI_API_KEY=your-api-key-here

# Run the example
go run main.go
```

## Key Concepts

### Loading Message History at Runtime

```go
// Load conversation history (from database, session, etc.)
messageHistory := []llm.Message{
    {Role: "user", Content: "Previous question"},
    {Role: "assistant", Content: "Previous response"},
}

// Create stateless agent
agent := agent.NewChatAgent("assistant").
    WithModel("gpt-3.5-turbo").
    WithClient(llmClient)

// Load history at runtime via context
ctx := workflow.NewWorkContext(context.Background())
ctx.Set("message_history", messageHistory) // Runtime loading
ctx.Set("user_input", "Continue conversation...")
```

### Maintaining Context

The message history allows agents to:
- Reference previous topics naturally
- Maintain conversation flow
- Build on prior information
- Keep track of user preferences or context

### Multi-turn Conversations

For ongoing conversations, you can:
1. Start with empty history
2. After each turn, append the user input and assistant response to your history
3. Load the updated history at runtime for the next turn

## Use Cases

- **Customer Support**: Maintain context across support interactions
- **Educational Tutoring**: Remember what topics have been covered
- **Personal Assistants**: Keep track of user preferences and ongoing tasks
- **Conversational AI**: Build natural, context-aware dialogues

## Best Practices

1. **Stateless Agents**: Agents are stateless - all context comes from WorkContext
2. **History Management**: Keep history to a reasonable length to avoid token limits  
3. **Context Relevance**: Only include relevant previous messages
4. **System Prompts**: Agents automatically deduplicate system prompts in history
5. **Privacy**: Be mindful of storing sensitive information in message history
6. **Runtime Loading**: Always load history at runtime for production applications