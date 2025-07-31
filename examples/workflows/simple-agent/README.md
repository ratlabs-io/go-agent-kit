# Simple Agent Example

This example demonstrates a basic workflow with a single chat agent - the foundation for understanding go-agent-kit.

## What it demonstrates

- Creating and configuring a chat agent
- Setting up LLM client integration 
- Running a simple agent workflow
- Handling agent responses and errors
- **NEW**: JSON response formatting for structured output

## Key concepts

- **ChatAgent creation**: `agent.NewChatAgent("name")` with smart defaults (4000 tokens, 0.7 temperature, 0.95 top-p)
- **Fluent configuration**: `.WithModel()`, `.WithPrompt()`, `.WithClient()`
- **Generation parameters**: `.WithMaxTokens()`, `.WithTemperature()`, `.WithTopP()` for fine-tuning
- **JSON responses**: `.WithJSONResponse()` for structured output
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

--- Example 1: Basic Text Response ---
Question: What is the capital of France?
Running chat agent...
✅ Agent completed successfully!
Answer: The capital of France is Paris.

--- Example 2: JSON Response ---
Question: What is the capital of France?
Running JSON chat agent...
✅ Agent completed successfully!
JSON Response:
{
  "answer": "Paris",
  "confidence": "high", 
  "category": "geography",
  "additional_info": "Paris is also the largest city in France and a major cultural center."
}
```

The example now shows both traditional text responses and structured JSON responses, demonstrating how the same agent can provide data in different formats depending on your application needs.