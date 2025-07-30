# Error Handling Example

This example demonstrates structured error handling patterns using try-catch-finally constructs in go-agent-kit.

## What it does

1. **Basic Try-Catch** - Simple error handling with catch-all handler
2. **Specific Error Types** - Different handlers for different error types
3. **Try-Catch-Finally** - Always execute cleanup code regardless of outcome
4. **Nested Try-Catch** - Complex operations with nested error handling
5. **Error Recovery** - Transform errors into successful results with fallback
6. **Complex Classification** - Advanced error matching and routing

## Key concepts

- **NewTryCatch(name)** - Create try-catch-finally construct
- **WithTryAction(action)** - Set the action to execute in try block
- **Catch(matcher, handler)** - Handle specific error types/conditions
- **CatchAny(handler)** - Handle any unmatched errors (catch-all)
- **Finally(action)** - Always execute cleanup code
- **Error Matchers**:
  - `TimeoutError` - Match timeout-related errors
  - `NetworkError` - Match network-related errors
  - `ValidationError` - Match validation-related errors
  - `ErrorMessageContains(text)` - Match errors containing text
  - `ErrorMessageEquals(text)` - Match exact error messages
  - `CombineErrorMatchers()` - Combine multiple matchers with OR logic
- **Error Recovery** - Transform failures into successes
- **Composability** - Nest try-catch blocks within workflows

## Running the example

```bash
go run examples/workflows/error-handling/main.go
```

## Sample output

```
=== Error Handling Example ===

--- Example 1: Basic Try-Catch ---
Simple error handling with catch-all...
  Attempting risky operation...
  ❌ Operation failed!
  🔧 Caught error: operation failed unexpectedly
  📝 Logging error and continuing...

--- Example 2: Specific Error Type Handling ---
Different handlers for different error types...
  Attempting file operation...
  ❌ Connection timeout!
  ⏰ Handling timeout: connection timeout
  🔄 Increasing timeout and will retry later...

--- Example 3: Try-Catch-Finally ---
Always execute cleanup code...
  🔓 Acquiring resources...
  📊 Processing data...
  ❌ Processing failed!
  🛠️ Handling processing error: data processing failed
  📁 Saving partial results...
  🧹 Cleaning up resources...
  🔒 Releasing locks...
  📋 Updating status...

--- Example 4: Nested Try-Catch ---
Nested error handling for complex operations...
  🌐 Making API call...
  ✅ API call successful!
    🗄️ Executing database query...
    ✅ Database query successful!

--- Example 5: Error Recovery and Transformation ---
Transform errors into successful results...
  📡 Calling primary service...
  🔄 Primary failed: primary service unavailable
  🏥 Switching to backup service...
  ✅ Backup service responded!
  💾 Final result: fallback data

--- Example 6: Complex Error Classification ---
Advanced error matching and routing...
  ⚡ Executing complex operation...
  ❌ Error: rate limit: too many requests
  ⏳ Rate limit error: rate limit: too many requests
  😴 Backing off for extended period...
```

## Use cases

- **API Error Handling** - Handle different HTTP error codes appropriately
- **Database Operations** - Manage connection failures, timeouts, constraint violations
- **File Operations** - Handle permission errors, disk full, file not found
- **Network Operations** - Manage timeouts, connection failures, rate limits
- **Resource Management** - Ensure proper cleanup with finally blocks
- **Service Integration** - Graceful degradation with fallback services
- **Batch Processing** - Continue processing despite individual item failures
- **User Input Validation** - Different responses for different validation errors

## Error handling patterns

- **Fail Fast** - Let errors bubble up immediately
- **Graceful Degradation** - Fall back to alternative approaches
- **Error Recovery** - Transform errors into successful outcomes
- **Logging and Alerting** - Capture errors for monitoring
- **Resource Cleanup** - Always release resources in finally blocks
- **Error Classification** - Route different errors to appropriate handlers
- **Retry Integration** - Combine with retry patterns for robustness