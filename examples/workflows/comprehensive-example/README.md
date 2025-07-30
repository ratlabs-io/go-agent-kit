# Comprehensive Workflow Example

This example demonstrates a real-world e-commerce order processing system that uses **all** the new workflow constructs working together seamlessly.

## Scenario: E-commerce Order Processing

A complete order processing pipeline that handles multiple orders through a sophisticated workflow with:
- Validation
- Inventory checking
- Payment processing  
- Fulfillment coordination
- Cleanup and finalization

## Architecture Overview

```
📦 Order Processing Pipeline
│
├── 🔄 Loop Over Orders (NewLoopOver)
│   └── For each order:
│       │
│       ├── 📋 Validation Step
│       │   └── Try-Catch error handling
│       │
│       ├── 📦 Inventory Step  
│       │   ├── Timeout Wrapper (100ms)
│       │   ├── Circuit Breaker (2 failures → open)
│       │   └── Retry with Exponential Backoff (3 attempts)
│       │
│       ├── 💳 Payment Step
│       │   ├── Try-Catch with specific error handlers:
│       │   │   ├── Insufficient funds → Notify customer
│       │   │   ├── Timeout → Retry
│       │   │   └── Network error → Switch to backup
│       │   └── Retry with Linear Backoff (timeout errors only)
│       │
│       ├── 🎯 Fulfillment Step
│       │   └── Parallel Error Collector:
│       │       ├── Warehouse Service
│       │       ├── Shipping Service  
│       │       └── Notification Service
│       │
│       └── 🧹 Cleanup Step (Finally block)
```

## Constructs Demonstrated

### 1. **Loop Constructs**
- `NewLoopOver()` - Process each order in the batch

### 2. **Try-Catch Error Handling**
- `NewTryCatch()` - Structured exception handling
- `Catch()` - Specific error type handlers
- `CatchAny()` - Catch-all error handlers  
- `Finally()` - Always-execute cleanup

### 3. **Retry Patterns**
- `NewRetry()` - Retry failed operations
- `WithBackoffStrategy()` - Exponential and linear backoff
- `WithRetryCondition()` - Conditional retry logic

### 4. **Advanced Constructs**
- `NewCircuitBreaker()` - Prevent cascading failures
- `NewTimeoutWrapper()` - Timeout protection
- `NewParallelErrorCollector()` - Comprehensive parallel execution

### 5. **Existing Workflow Patterns**
- `NewSequentialFlow()` - Step-by-step processing
- `NewActionFunc()` - Custom business logic

## Error Handling Strategy

The example demonstrates multiple layers of error handling:

1. **Validation Errors** → Fail fast with clear messages
2. **Infrastructure Errors** → Circuit breaker + retry with backoff
3. **Payment Errors** → Different strategies per error type:
   - Insufficient funds → Notify customer (permanent failure)
   - Timeout → Retry with backoff
   - Network → Switch to backup gateway (recovery)
4. **Fulfillment Errors** → Continue if critical services succeed
5. **System Errors** → Comprehensive logging and cleanup

## Running the example

```bash
go run examples/workflows/comprehensive-example/main.go
```

## Sample output

```
=== Comprehensive Workflow Example ===
Demonstrating all new constructs working together in a real-world scenario...

📦 Processing 3 orders through comprehensive workflow...

🔄 Processing Order ORD-001...
  📋 Validating order ORD-001...
  ✅ Order ORD-001 validated
  📦 Checking inventory for 2 items...
  ✅ Inventory confirmed for order ORD-001
  💳 Processing payment of $99.99...
  ✅ Payment processed for order ORD-001
  🎯 Coordinating fulfillment services...
  🏭 Warehouse: Preparing shipment...
  🚚 Shipping: Creating label...
  📧 Notifications: Sending updates...
  ✅ Warehouse: Items packaged
  ✅ Shipping: Label created
  ✅ Notifications: Customer notified
  📊 Fulfillment result: 3/3 services succeeded
  ✅ Order ORD-001 fulfillment initiated
  🧹 Cleanup: Finalizing order ORD-001...
  ✅ Cleanup: Order ORD-001 finalized
  ✅ Order ORD-001 processed successfully!

🔄 Processing Order ORD-002...
  📋 Validating order ORD-002...
  ✅ Order ORD-002 validated
  📦 Checking inventory for 1 items...
  ✅ Inventory confirmed for order ORD-002
  💳 Processing payment of $49.99...
  🌐 Network issue - switching to backup gateway
  🎯 Coordinating fulfillment services...
  🏭 Warehouse: Preparing shipment...
  🚚 Shipping: Creating label...
  📧 Notifications: Sending updates...
  ✅ Warehouse: Items packaged
  ⚠️ Some fulfillment services had issues:
    - shipping label generation failed
  📊 Fulfillment result: 2/3 services succeeded
  ✅ Order ORD-002 fulfillment initiated
  🧹 Cleanup: Finalizing order ORD-002...
  ✅ Cleanup: Order ORD-002 finalized
  ✅ Order ORD-002 processed successfully!

============================================================
🎯 Order Processing Complete! (Total time: 543ms)
✅ All orders processed successfully!

🏆 This example demonstrated:
  • Loop constructs for batch processing
  • Try-catch for structured error handling
  • Retry with backoff for resilience
  • Circuit breakers for fault tolerance
  • Timeout wrappers for resource protection
  • Parallel error collection for comprehensive monitoring
  • Sequential flows for step-by-step processing
  • All constructs working together seamlessly!
```

## Key Benefits Demonstrated

### 1. **Resilience**
- Circuit breakers prevent cascading failures
- Retries with backoff handle transient errors
- Timeouts prevent resource exhaustion

### 2. **Reliability**
- Structured error handling with recovery strategies
- Parallel execution with comprehensive error collection
- Always-execute cleanup ensures consistent state

### 3. **Observability**
- Detailed logging at each step
- Error classification and routing
- Comprehensive metrics and reporting

### 4. **Maintainability**
- Clear separation of concerns
- Composable and reusable constructs
- Easy to add new error handling strategies

### 5. **Performance**
- Parallel execution where possible
- Fail-fast validation
- Efficient resource utilization with timeouts

## Real-world Applications

This pattern is ideal for:
- **E-commerce order processing**
- **Financial transaction processing**
- **Data pipeline orchestration**
- **Microservice coordination**
- **Batch job processing**
- **ETL workflows**
- **API gateway routing**
- **System health monitoring**

## Customization

The example can be easily adapted for different scenarios by:
- Changing error handling strategies per business requirements
- Adjusting timeout and retry parameters
- Adding new validation or processing steps
- Modifying parallel execution patterns
- Integrating with external monitoring systems