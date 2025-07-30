# Chat Agent with Message History Example

This example demonstrates how to use message history with both ChatAgent and ToolAgent to maintain conversation context across multiple interactions.

## Features Demonstrated

1. **Basic Chat with History**: Shows how to load previous conversation and continue naturally
2. **Tool Agent with History**: Demonstrates tool usage with conversation context
3. **Multi-turn Conversation**: Shows how to build up history across multiple agent calls

## Running the Example

```bash
# Set your OpenAI API key
export OPENAI_API_KEY=your-api-key-here

# Run the example
go run main.go
```

## Key Concepts

### Loading Message History

```go
// Create message history
messageHistory := []llm.Message{
    {Role: "user", Content: "Previous question"},
    {Role: "assistant", Content: "Previous response"},
}

// Create agent with history
agent := agent.NewChatAgent("assistant").
    WithMessageHistory(messageHistory)
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
2. After each turn, append the user input and assistant response
3. Update the agent with the new history for the next turn

## Use Cases

- **Customer Support**: Maintain context across support interactions
- **Educational Tutoring**: Remember what topics have been covered
- **Personal Assistants**: Keep track of user preferences and ongoing tasks
- **Conversational AI**: Build natural, context-aware dialogues

## Best Practices

1. **History Management**: Keep history to a reasonable length to avoid token limits
2. **Context Relevance**: Only include relevant previous messages
3. **System Prompts**: Be careful not to duplicate system prompts in history
4. **Privacy**: Be mindful of storing sensitive information in message history