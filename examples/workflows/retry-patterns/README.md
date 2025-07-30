# Retry Patterns Example

This example demonstrates all the retry patterns and backoff strategies available in go-agent-kit for robust error recovery.

## What it does

1. **Basic Retry** - Simple retry with fixed backoff delays (API calls)
2. **Linear Backoff** - Steadily increasing delays (database connections)
3. **Exponential Backoff** - Rapidly increasing delays with cap (external services)
4. **Custom Retry Conditions** - Only retry specific error types (timeout vs validation)
5. **Combined Conditions** - Retry on multiple error types (network issues)
6. **Stop Conditions** - Halt retries based on context state (time limits)
7. **Complex Pattern** - Sophisticated file processing with retry logic

## Key concepts

- **NewRetry(name, maxAttempts)** - Basic retry with attempt count
- **Backoff Strategies**:
  - `NewFixedBackoff(delay)` - Consistent delays
  - `NewLinearBackoff(base, increment)` - Linear increase
  - `NewExponentialBackoff(base, max, factor)` - Exponential increase with cap
- **Retry Conditions**:
  - `DefaultRetryCondition` - Retry on any error
  - `RetryOnTimeoutCondition` - Only timeout errors
  - `RetryOnNetworkCondition` - Only network errors
  - `RetryOnRateLimitCondition` - Only rate limit errors
  - `CombineRetryConditions()` - Combine multiple conditions
- **Stop Conditions** - Custom logic to halt retries early
- **Metadata** - Retry information in final report

## Running the example

```bash
go run examples/workflows/retry-patterns/main.go
```

## Sample output

```
=== Retry Patterns Example ===

--- Example 1: Basic Retry with Fixed Backoff ---
Attempting API call with fixed 100ms delays...
  Attempt 1: Making API call... ❌ Failed (network error)
  Attempt 2: Making API call... ❌ Failed (network error)  
  Attempt 3: Making API call... ✅ Success!

--- Example 2: Linear Backoff Strategy ---
Database connection with increasing delays (100ms, 200ms, 300ms...)...
  Attempt 1: Connecting to database... ❌ Failed (connection refused)
  Attempt 2: Connecting to database... ❌ Failed (connection refused)
  Attempt 3: Connecting to database... ✅ Connected!

--- Example 3: Exponential Backoff Strategy ---
External service with exponential backoff (100ms, 200ms, 400ms...)...
  Attempt 1: Calling external service... ❌ Failed (service unavailable)
  Attempt 2: Calling external service... ❌ Failed (service unavailable)
  Attempt 3: Calling external service... ✅ Service responded!

--- Example 4: Custom Retry Conditions ---
Only retry on timeout errors, not validation errors...
  Attempt 1: ❌ validation error: invalid input
  (No retry - validation errors are not retryable)

--- Example 5: Combined Retry Conditions ---
Retry on timeouts OR network errors...
  Attempt 1: ❌ connection refused
  Attempt 2: ❌ rate limit exceeded  
  Attempt 3: ❌ connection timeout
  Attempt 4: ✅ Success!

--- Example 6: Stop Condition Based on Context ---
Stop retrying if total time exceeds threshold...
  Attempt 1 (after 0ms): Processing... ❌ Failed
  Attempt 2 (after 100ms): Processing... ❌ Failed
  Attempt 3 (after 200ms): Processing... ❌ Failed
  (Stopped - time limit exceeded)
```

## Use cases

- **API Resilience** - Handle transient network failures
- **Database Connections** - Retry connection establishment
- **External Services** - Robust integration with third-party APIs
- **File Operations** - Handle temporary file system issues
- **Rate Limiting** - Respect API rate limits with backoff
- **Circuit Breaking** - Combine with circuit breakers for advanced patterns
- **Batch Processing** - Retry failed items in batch operations

## Best practices

- **Choose appropriate backoff**: Fixed for quick operations, exponential for external services
- **Set reasonable limits**: Don't retry forever, respect time/attempt bounds
- **Match conditions to errors**: Only retry recoverable errors
- **Monitor retry patterns**: Use metadata to track retry behavior
- **Combine with other patterns**: Circuit breakers, timeouts, etc.