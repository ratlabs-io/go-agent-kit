# Advanced Constructs Example

This example demonstrates advanced reliability patterns including circuit breakers, timeout wrappers, and parallel error collection in go-agent-kit.

## What it does

1. **Circuit Breaker** - Protect against cascading failures with automatic state management
2. **Timeout Wrapper** - Prevent operations from running too long with context cancellation
3. **Parallel Error Collector** - Run multiple actions concurrently and collect comprehensive results
4. **Combined Patterns** - Demonstrate composing multiple reliability patterns together
5. **Health Check System** - Build a robust system health monitoring solution

## Key concepts

- **NewCircuitBreaker(name, threshold, recovery, reset)** - Failure protection with state tracking
  - `threshold` - Number of failures before opening circuit
  - `recovery` - Time to wait before trying half-open state
  - `reset` - Time to reset failure count in closed state
- **NewTimeoutWrapper(name, timeout)** - Context-based timeout protection
- **NewParallelErrorCollector(name)** - Comprehensive parallel execution with error collection
- **Circuit Breaker States**:
  - `Closed` - Normal operation, requests allowed
  - `Open` - Failing state, requests rejected
  - `Half-Open` - Testing recovery, limited requests allowed
- **Composability** - Combine patterns for layered protection
- **Metrics** - Get detailed metrics from circuit breakers

## Running the example

```bash
go run examples/workflows/advanced-constructs/main.go
```

## Sample output

```
=== Advanced Constructs Example ===

--- Example 1: Circuit Breaker ---
Protecting against cascading failures...
Call 1:   ❌ Service call failed
    State: 0, Failures: 1/3
Call 2:   ❌ Service call failed
    State: 0, Failures: 2/3
Call 3:   ❌ Service call failed
    State: 1, Failures: 3/3
Call 4: 🚫 Circuit breaker is OPEN - request rejected
    State: 1, Failures: 3/3
Call 5: 🚫 Circuit breaker is OPEN - request rejected
    State: 1, Failures: 3/3

--- Example 2: Timeout Wrapper ---
Preventing long-running operations...
Attempt 1:   🐌 Starting slow operation...
  ⏳ Working for 145ms...
  ✅ Slow operation completed
✅ Succeeded after 145ms
Attempt 2:   🐌 Starting slow operation...
  ⏳ Working for 267ms...
  ⏰ Operation was cancelled due to timeout
❌ Failed after 200ms

--- Example 3: Parallel Error Collector ---
Running multiple services and collecting all results...
  🔐 Auth service: ✅ User authenticated
  💳 Payment service: ❌ Payment declined
  📦 Inventory service: ✅ Item reserved
  📧 Notification service: ✅ Email sent

📊 Results after 80ms:
  Total services: 4
  Successful: 3
  Failed: 1
  Total errors: 1
  🚨 Errors encountered:
    1. payment declined
  ⚠️ Overall status: Some services have issues

--- Example 4: Combined Advanced Patterns ---
Circuit breaker + timeout + error collection...
Running protected services...

Round 1:
  Completed in 150ms
  Success rate: 1/2
    ⚠️ service timeout

Round 2:
  Completed in 20ms
  Success rate: 1/2
    ⚠️ circuit breaker service-circuit is open

--- Example 5: Comprehensive Health Check System ---
Building a robust health check system...
Running comprehensive health checks...

🏥 System Health Report:
  Total checks: 3
  Passed: 2
  Failed: 1
  Health score: 66.7%
  Status: 🟡 Some systems degraded
```

## Use cases

### Circuit Breaker
- **Microservice protection** - Prevent cascading failures between services
- **Database connection pooling** - Protect against database overload
- **External API calls** - Handle third-party service failures gracefully
- **Resource protection** - Prevent resource exhaustion

### Timeout Wrapper
- **API calls** - Prevent hanging requests to external services
- **Database queries** - Avoid long-running queries blocking resources
- **File operations** - Handle slow disk I/O operations
- **Network operations** - Manage network latency and failures

### Parallel Error Collector
- **Health checks** - Monitor multiple system components
- **Batch processing** - Process multiple items with comprehensive error reporting
- **Service orchestration** - Coordinate multiple microservices
- **Validation pipelines** - Run multiple validation checks

## Reliability patterns

- **Fail Fast** - Circuit breakers reject requests when service is known to be down
- **Timeout Protection** - Prevent resource exhaustion from slow operations
- **Comprehensive Monitoring** - Collect all errors for full system visibility
- **Graceful Degradation** - Continue operation even with some component failures
- **State Management** - Circuit breakers track failure patterns over time
- **Composable Protection** - Layer multiple patterns for robust systems

## Best practices

- **Choose appropriate thresholds** - Balance protection with availability
- **Monitor circuit breaker metrics** - Track state changes and failure patterns
- **Set reasonable timeouts** - Based on expected operation duration
- **Use parallel collection for independence** - When operations don't depend on each other
- **Combine patterns strategically** - Layer protection without over-engineering
- **Test failure scenarios** - Verify patterns work under actual failure conditions