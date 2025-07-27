# Parallel Workflow Example

This example demonstrates concurrent execution of multiple agents followed by result synthesis.

## What it demonstrates

- Running multiple agents simultaneously for speed
- Collecting and combining results from parallel execution
- Combining parallel and sequential workflows
- Multi-perspective analysis pattern

## Key concepts

- **Parallel execution**: `workflow.NewParallelFlow("name")`
- **Agent concurrency**: `.Execute()` adds agents to run simultaneously
- **Result combination**: Parallel outputs automatically combined
- **Workflow composition**: Parallel flow used within sequential flow
- **Performance optimization**: Multiple agents run concurrently

## Running the example

```bash
export OPENAI_API_KEY=your-actual-api-key-here
go run examples/workflows/parallel-workflow/main.go
```

## Sample output

```
=== Parallel Workflow Example ===
Company: Tesla

✅ Workflow completed successfully!

Synthesized Analysis: [Combined insights from technical, market, and risk analysts running in parallel, then synthesized by a final agent]
```