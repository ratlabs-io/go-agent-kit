# Workflow Examples

This directory contains practical examples demonstrating how to use the go-agent-kit library to build agent workflows. Each example is in its own directory with a `main.go` file that can be run independently.

## Examples

### 1. `simple-agent/`
**Basic single agent workflow**
- Shows how to create and run a single chat agent
- Demonstrates basic agent configuration (model, prompt, LLM client)
- **NEW**: Includes JSON response formatting examples
- Good starting point for understanding the library

### 2. `simple-agent-with-callbacks/`
**Agent workflow with event monitoring**
- Single agent with callback registration for events
- Shows how to monitor agent execution with callbacks
- Demonstrates the callback system and event handling

### 3. `structured-json-agent/`
**Structured JSON responses with schemas**
- Comprehensive JSON schema response examples
- Shows both strict schema and generic JSON responses
- Multi-step workflows with structured data flow
- Demonstrates type-safe parsing and validation

### 4. `chat-with-history/`
**Chat agents with message history**
- Shows how to load and maintain conversation context
- Demonstrates both ChatAgent and ToolAgent with history
- Multi-turn conversation example with history building
- Perfect for stateful, context-aware applications

### 5. `sequential-workflow/`
**Sequential multi-agent workflow**
- Multiple agents running one after another
- Shows how to chain agent outputs with `ThenChain()`
- Demonstrates `SequentialFlow` usage

### 6. `chaining-patterns/`
**Different sequential chaining strategies**
- Compares `ThenChain()` vs `ThenAccumulate()` patterns
- Shows how data flows through agent chains
- Pipeline vs collaborative processing approaches

### 7. `parallel-workflow/`
**Parallel multi-agent workflow**
- Multiple agents running simultaneously
- Shows how to combine parallel execution results
- Demonstrates `ParallelFlow` usage

### 8. `conditional-workflow/`
**Conditional workflow execution**
- Shows how to implement branching logic
- Demonstrates `ConditionalFlow` and custom predicates
- Routing based on agent output classification

### 9. `switch-workflow/`
**Switch-based workflow routing**
- Multiple condition evaluation with priority ordering
- Shows sentiment-based routing with urgency override
- Demonstrates `SwitchFlow` builder pattern with default fallback

### 10. `action-func/`
**Simple action creation with ActionFunc**
- Shows how to create actions without boilerplate
- Demonstrates text processing pipeline with validation
- Perfect for quick prototyping and simple transformations

### 11. `tool-agent/`
**Tool-enabled agent workflow**
- Shows how to create agents that can use tools
- Demonstrates tool registration and execution
- Examples of math, echo, and simple tools

### 12. `custom-logging/`
**Flexible logging configuration**
- Default logging behavior (stderr, Info level)
- Disabling all logging with NoOpLogger
- Custom global logger configuration
- Per-context logger settings

## Key Concepts Demonstrated

### Agent Types
- **ChatAgent**: Simple 1-hop LLM completions with text or JSON responses
- **ToolAgent**: Multi-step execution with tool calling capabilities and JSON support

### Workflow Types  
- **SequentialFlow**: Execute actions one after another
- **ParallelFlow**: Execute actions simultaneously
- **ConditionalFlow**: Execute based on predicate conditions

### Response Types
- **Text Responses**: Traditional unstructured text output
- **JSON Objects**: Generic JSON responses for flexible data structures
- **JSON Schema**: Strict schema validation for predictable, type-safe outputs

### Tool Integration
- **Native Tools**: Implement `Tool` interface with full schema control
- **Simple Tools**: Implement `SimpleTool` interface with auto-generated schemas
- **MCP Tools**: Connect to external Model Context Protocol servers

### Context Management
- Data sharing between workflow steps via `Set()`/`Get()`
- Agent result storage and retrieval
- Cross-action communication through context
- Message history for maintaining conversation state

## Running Examples

**Setup Required**: These examples require an OpenAI API key set as an environment variable:

```bash
export OPENAI_API_KEY=your-actual-api-key-here
```

**Running the examples:**

```bash
# Run any example directly
go run examples/workflows/simple-agent/main.go
go run examples/workflows/simple-agent-with-callbacks/main.go
go run examples/workflows/structured-json-agent/main.go
go run examples/workflows/chat-with-history/main.go
go run examples/workflows/sequential-workflow/main.go
go run examples/workflows/chaining-patterns/main.go
go run examples/workflows/parallel-workflow/main.go
go run examples/workflows/conditional-workflow/main.go
go run examples/workflows/switch-workflow/main.go
go run examples/workflows/tool-agent/main.go
go run examples/workflows/action-func/main.go
go run examples/workflows/custom-logging/main.go
```

**Note**: If the `OPENAI_API_KEY` environment variable is not set, the examples will show an error message and exit gracefully.

## Configuration

### LLM Providers
Examples use OpenAI, but you can substitute any LLM client that implements `llm.Client`:

```go
// OpenAI
llmClient := openai.NewClient("your-api-key")

// GoLLM wrapper (requires external dependency)
llmClient, _ := gollm.NewClient("openai", "gpt-3.5-turbo")

// Custom implementation
llmClient := &YourCustomClient{}
```

### Error Handling
All examples include basic error handling patterns:

```go
if report.Status == workflow.StatusCompleted {
    // Success - access results
} else {
    // Handle errors in report.Errors
}
```

## Building Complex Workflows

These examples can be combined and extended:

1. **Nested Workflows**: Use workflows as actions in other workflows
2. **Agent Specialization**: Create domain-specific agents with tailored prompts
3. **Tool Composition**: Combine multiple tools for complex operations
4. **Error Recovery**: Add retry logic and fallback strategies
5. **Event Handling**: Use callbacks for monitoring and debugging

## Best Practices

1. **Agent Naming**: Use descriptive names for debugging and monitoring
2. **Prompt Engineering**: Tailor prompts to specific tasks and contexts
3. **Tool Design**: Keep tools focused and composable
4. **Context Management**: Use meaningful keys for context data
5. **Error Handling**: Always check workflow reports for errors
6. **Resource Cleanup**: Call `Close()` on LLM clients when done