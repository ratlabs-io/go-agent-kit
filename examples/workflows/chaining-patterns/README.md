# Chaining Patterns Example

This example demonstrates different strategies for chaining agents in sequential workflows, showing how output passes between agents.

## What it demonstrates

- **ThenChain pattern**: Each agent receives only the previous agent's output as input
- **ThenAccumulate pattern**: Each agent receives original input plus all previous outputs (snowball effect)
- Comparing different chaining strategies for agent workflows

## Key concepts

- **Output chaining**: `ThenChain()` rotates context so previous output becomes current input
- **Output accumulation**: `ThenAccumulate()` builds up context with all previous outputs
- **Context management**: How data flows through sequential workflows

## Chaining Strategies

### ThenChain (Output → Input)
```go
chainFlow := workflow.NewSequentialFlow("chain-pattern").
    Then(researcher).        // Gets original user input
    ThenChain(summarizer)    // Gets only researcher's output as input
```

### ThenAccumulate (Snowball Effect)  
```go
accumFlow := workflow.NewSequentialFlow("accumulate-pattern").
    Then(researcher).           // Gets original user input
    ThenAccumulate(summarizer). // Gets original input + researcher output
    ThenAccumulate(reviewer)    // Gets original + researcher + summarizer
```

## When to use each pattern

- **ThenChain**: When each step should transform the previous result (pipeline processing)
- **ThenAccumulate**: When agents need full context history (collaborative refinement)

## Running the example

```bash
export OPENAI_API_KEY=your-actual-api-key-here
go run examples/workflows/chaining-patterns/main.go
```

## Sample output

```
=== Chaining Patterns Example ===

🔗 Pattern 1: ThenChain (output → input)
Final: Machine learning is a subset of AI focusing on algorithms that improve through experience...

📚 Pattern 2: ThenAccumulate (snowball effect)  
Final: Based on the research and summary, this is a comprehensive overview that covers key aspects of machine learning...
```

This example helps you choose the right chaining strategy for your specific workflow needs.