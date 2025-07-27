# Switch Workflow Example

This example demonstrates how to use switch workflows to route requests based on multiple conditions with priority ordering.

## What it does

1. **Analyzes sentiment** - Determines if input is positive, negative, or neutral
2. **Routes with priority** - Uses a switch flow to route based on multiple conditions:
   - Urgent keywords (highest priority) ’ Urgent responder
   - Positive sentiment ’ Enthusiastic responder
   - Negative sentiment ’ Empathetic responder  
   - Neutral sentiment ’ Professional responder
   - Default fallback ’ Professional responder

## Key concepts

- **SwitchFlow** - Evaluates multiple conditions in order and executes first match
- **Builder pattern** - Fluent API for constructing switch cases
- **Priority routing** - Earlier cases take precedence over later ones
- **Default action** - Fallback when no conditions match
- **Multi-factor conditions** - Checking both content and sentiment

## Running the example

```bash
export OPENAI_API_KEY=your-api-key-here
go run examples/workflows/switch-workflow/main.go
```

## Sample output

```
=== Switch Workflow Example ===

--- Test 1 ---
Input: I'm so excited about this new project!
Response: [Enthusiastic response with energy and optimism]

--- Test 2 ---
Input: I'm really struggling with this problem and feeling frustrated.
Response: [Empathetic response with understanding and support]

--- Test 3 ---
Input: Can you explain how databases work?
Response: [Professional response with clear information]

--- Test 4 ---
Input: URGENT: I need help immediately with a server outage!
Response: [Immediate actionable help response]

--- Test 5 ---
Input: This is terrible, nothing is working right.
Response: [Empathetic response addressing frustration]
```