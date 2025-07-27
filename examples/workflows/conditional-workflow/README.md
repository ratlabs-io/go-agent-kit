# Conditional Workflow Example

This example demonstrates how to use conditional workflows to route requests to different agents based on classification results.

## What it does

1. **Classifies input** - A classifier agent determines if the input is "technical", "creative", or "general"
2. **Routes conditionally** - Based on the classification, routes to the appropriate specialized agent:
   - Technical questions → Technical expert
   - Creative requests → Creative writer  
   - Everything else → General assistant

## Key concepts

- **ConditionalFlow** - Executes different actions based on predicate evaluation
- **Predicates** - Functions that evaluate context and return true/false
- **Sequential chaining** - Classification followed by conditional routing
- **Skipped actions** - Conditional flows skip when predicate returns false

## Running the example

```bash
export OPENAI_API_KEY=your-api-key-here
go run examples/workflows/conditional-workflow/main.go
```

## Sample output

```
=== Conditional Workflow Example ===

--- Test 1 ---
Input: How does machine learning work?
Response: [Technical expert provides detailed ML explanation]

--- Test 2 ---
Input: Write a story about a robot who dreams
Response: [Creative writer provides imaginative story]

--- Test 3 ---
Input: What's the weather like today?
Response: [General assistant provides helpful response]
```