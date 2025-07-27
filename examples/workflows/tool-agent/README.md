# Tool Agent Example

This example demonstrates an agent that can use tools to complete complex tasks requiring external capabilities.

## What it demonstrates

- Creating tool-enabled agents
- Registering multiple tools with agents  
- Tool execution and result handling
- Both simple and complex tool interfaces

## Key concepts

- **ToolAgent creation**: `agent.NewToolAgent("name")`
- **Tool registration**: `.WithTools(tool1, tool2, tool3)`
- **Tool interfaces**: Both `Tool` and `SimpleTool` patterns
- **Tool wrapping**: `tools.WrapSimpleTool()` for simple tools
- **Tool execution**: Automatic tool calling based on LLM decisions

## Running the example

```bash
export OPENAI_API_KEY=your-actual-api-key-here
go run examples/workflows/tool-agent/main.go
```

## Sample output

```
=== Tool Agent Workflow Example ===
Task: Calculate 15 * 23 and echo the result
Running tool agent...

✅ Tool agent completed successfully!

🤖 Agent Response:
==================
I'll help you calculate 15 * 23 and echo the result.

🔧 Tool Results:
================
Math: 15 * 23 = 345
```