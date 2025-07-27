# Simple Agent Example

This example demonstrates a basic workflow with a single chat agent - the foundation for understanding go-agent-kit.

## What it demonstrates

- Creating and configuring a chat agent
- Setting up LLM client integration 
- Running a simple agent workflow
- Handling agent responses and errors

## Key concepts

- **ChatAgent creation**: `agent.NewChatAgent("name")`
- **Fluent configuration**: `.WithModel()`, `.WithPrompt()`, `.WithClient()`
- **Execution context**: `workflow.NewWorkContext(ctx)`
- **Running agents**: `agent.Run(workflowCtx)`
- **Response handling**: Checking `report.Status` and extracting results

## Running the example

```bash
export OPENAI_API_KEY=your-actual-api-key-here
go run examples/workflows/simple-agent/main.go
```

## Sample output

```
=== Simple Agent Workflow Example ===
Question: What is the capital of France?
Running agent...

✅ Agent completed successfully!
Answer: The capital of France is Paris.
```