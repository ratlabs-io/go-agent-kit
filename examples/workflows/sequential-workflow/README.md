# Sequential Workflow Example

This example demonstrates a multi-agent workflow where agents run one after another, building on each other's output.

## What it demonstrates

- Creating multiple specialized agents
- Chaining agents with `ThenChain()` for output passing
- Data flow between workflow steps
- Using `SequentialFlow` for ordered execution

## Key concepts

- **Sequential execution**: `workflow.NewSequentialFlow("name")`
- **Agent chaining**: `.ThenChain()` passes previous output as next input
- **Agent specialization**: Different agents for research, summarization, analysis
- **Context management**: Shared data flows automatically between steps

## Running the example

```bash
export OPENAI_API_KEY=your-actual-api-key-here
go run examples/workflows/sequential-workflow/main.go
```

## Sample output

```
=== Sequential Workflow Example ===
Research Topic: Benefits of renewable energy

✅ Workflow completed successfully!

Final Analysis: [Combined research, summary, and analysis results showing the progression through each agent step]
```