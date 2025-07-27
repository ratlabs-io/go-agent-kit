# Custom Logging Example

This example demonstrates how to configure logging in go-agent-kit. The library provides flexible logging options without forcing any particular solution on users.

## What it demonstrates

- **Default logging**: Library uses sensible defaults (stderr, Info level)
- **Disable logging**: Complete logging suppression with `NoOpLogger`
- **Custom global logger**: Set your own logger for the entire application
- **Per-context logger**: Use different loggers for different workflows

## Key concepts

- **Default logger**: Works out of the box, logs to stderr at Info level
- **NoOpLogger**: Disables all logging (zero overhead)
- **Custom loggers**: Use any `slog.Logger` or implement `Logger` interface
- **Context-specific**: Different workflows can use different loggers

## Logging options

### 1. Default (no configuration needed)
```go
ctx := workflow.NewWorkContext(context.Background())
// Uses default logger - logs to stderr at Info level
```

### 2. Disable all logging
```go
workflow.SetDefaultLogger(workflow.NewNoOpLogger())
// All logging operations become no-ops
```

### 3. Custom global logger
```go
customLogger := workflow.NewSlogLogger(
    slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    })),
)
workflow.SetDefaultLogger(customLogger)
```

### 4. Per-context logger
```go
contextLogger := workflow.NewSlogLogger(yourLogger)
ctx := workflow.NewWorkContextWithLogger(context.Background(), contextLogger)
```

## Running the example

```bash
export OPENAI_API_KEY=your-actual-api-key-here
go run examples/workflows/custom-logging/main.go
```

## Sample output

```
=== Custom Logging Example ===

--- Test 1: Default Logging (to stderr) ---
INFO agent completed agent=ChatAgent name=assistant elapsed=1.2s
✅ Response: Hello! How can I help you today?

--- Test 2: No Logging ---
✅ Response: Hello! How can I help you? (no logs should appear)

--- Test 3: Custom Logger ---
{"level":"INFO","msg":"agent completed","app":"go-agent-kit-example","agent":"ChatAgent","name":"assistant","elapsed":"1.1s"}
✅ Response: Hello! How can I assist you today?

--- Test 4: Per-Context Logger ---
INFO agent completed context=special-context agent=ChatAgent name=assistant elapsed=1.0s
✅ Response: Hello! I'm here to help!
```

## Implementation Notes

- **Zero dependencies**: Custom loggers use standard `log/slog`
- **No forced logging**: Users can completely disable if desired
- **Backward compatible**: Default behavior works without configuration
- **Flexible**: Support for global, per-context, and custom loggers
- **Production ready**: Proper log levels, structured logging, performance-conscious