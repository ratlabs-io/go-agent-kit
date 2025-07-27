# ActionFunc Example

This example demonstrates the `ActionFunc` helper that allows creating workflow actions on the fly without implementing the full `Action` interface.

## Key Features

- **Simple action creation**: Just provide a name and a function
- **No boilerplate**: No need to create structs or implement interfaces
- **Full workflow integration**: Works with sequential, parallel, and conditional flows
- **Context access**: Full access to WorkContext for reading and writing data

## Code Highlights

### Creating Simple Actions
```go
action := workflow.NewActionFunc("process", func(ctx *workflow.WorkContext) {
    // Your logic here - no return needed!
    if data, ok := ctx.Get("input"); ok {
        // Process data
        ctx.Set("output", processedData)
    }
    // Use panic("error message") for failures
})
```

### Building Workflows
```go
workflow := workflow.NewSequentialFlow("pipeline").
    Then(workflow.NewActionFunc("step1", step1Func)).
    Then(workflow.NewActionFunc("step2", step2Func)).
    Then(workflow.NewActionFunc("step3", step3Func))
```

## Use Cases

`ActionFunc` is perfect for:

1. **Quick prototyping** - Test workflow ideas without creating full agents
2. **Simple transformations** - Data preprocessing, validation, formatting
3. **Glue logic** - Connect agents with custom logic
4. **Debugging** - Add logging or inspection steps
5. **One-off actions** - Actions that don't need reuse

## Running the Example

```bash
go run examples/workflows/action-func/main.go
```

## Example Output

```
=== ActionFunc Example ===
This example shows how to create simple actions without boilerplate

Test 1: Valid input
-------------------
📝 Preprocessed: '  Hello World  ' → 'hello world'
✅ Validation passed: text length = 11
🔄 Transformed: 'hello world' → 'dlrow olleh'

📊 Context State:
----------------
  input_text:   Hello World  
  processed_text: hello world
  validated: true
  result: dlrow olleh

✅ Pipeline completed successfully!
```

## Benefits

1. **Reduced Boilerplate**: No need to create structs for simple actions
2. **Inline Logic**: Define action behavior right where you use it
3. **Full Power**: Access to all WorkContext features
4. **Type Safe**: Go's type system ensures correctness
5. **Composable**: Mix with regular Actions and Agents seamlessly

This makes go-agent-kit more accessible for simple use cases while maintaining the full power of the workflow system.