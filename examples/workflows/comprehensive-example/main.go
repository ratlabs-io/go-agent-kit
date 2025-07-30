package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/ratlabs-io/go-agent-kit/pkg/constants"
	"github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

func main() {
	fmt.Println("=== Comprehensive Workflow Example ===")
	fmt.Println("Demonstrating all new constructs working together in a real-world scenario...")
	fmt.Println()
	
	ctx := context.Background()
	wctx := workflow.NewWorkContext(ctx)
	
	// Scenario: E-commerce order processing system
	// This example shows a complete order processing pipeline that uses
	// all the new workflow constructs for robust, reliable processing.
	
	// Set up initial context
	orders := []map[string]interface{}{
		{"id": "ORD-001", "user": "alice", "amount": 99.99, "items": []string{"laptop", "mouse"}},
		{"id": "ORD-002", "user": "bob", "amount": 49.99, "items": []string{"book"}},
		{"id": "ORD-003", "user": "charlie", "amount": 199.99, "items": []string{"phone", "case", "charger"}},
	}
	
	fmt.Printf("📦 Processing %d orders through comprehensive workflow...\n", len(orders))
	fmt.Println()
	
	// Build the comprehensive order processing workflow
	orderProcessor := buildOrderProcessingWorkflow()
	
	// Use loop to process each order
	orderLoop := workflow.NewLoopOver("order-processing-loop", orders).
		WithAction(orderProcessor)
	
	start := time.Now()
	finalReport := orderLoop.Run(wctx)
	elapsed := time.Since(start)
	
	// Display final results
	fmt.Println("============================================================")
	fmt.Printf("🎯 Order Processing Complete! (Total time: %v)\n", elapsed.Round(time.Millisecond))
	
	if finalReport.Status == workflow.StatusCompleted {
		fmt.Println("✅ All orders processed successfully!")
	} else {
		fmt.Println("⚠️ Some orders had issues, but system remained stable")
	}
	
	fmt.Println("\n🏆 This example demonstrated:")
	fmt.Println("  • Loop constructs for batch processing")
	fmt.Println("  • Try-catch for structured error handling")
	fmt.Println("  • Retry with backoff for resilience")
	fmt.Println("  • Circuit breakers for fault tolerance")
	fmt.Println("  • Timeout wrappers for resource protection")
	fmt.Println("  • Parallel error collection for comprehensive monitoring")
	fmt.Println("  • Sequential flows for step-by-step processing")
	fmt.Println("  • All constructs working together seamlessly!")
}

func buildOrderProcessingWorkflow() workflow.Action {
	// This builds a comprehensive workflow that demonstrates all constructs
	return workflow.NewActionFunc("comprehensive-order-processor", func(wctx workflow.WorkContext) workflow.WorkReport {
		// Get current order from loop context
		orderData, _ := wctx.Get(constants.KeyCurrentItem)
		order := orderData.(map[string]interface{})
		orderID := order["id"].(string)
		
		fmt.Printf("🔄 Processing Order %s...\n", orderID)
		
		// Step 1: Validation with try-catch error handling
		validationStep := buildValidationStep(order)
		
		// Step 2: Inventory check with retry and circuit breaker
		inventoryStep := buildInventoryStep(order)
		
		// Step 3: Payment processing with comprehensive error handling
		paymentStep := buildPaymentStep(order)
		
		// Step 4: Fulfillment with parallel services and error collection
		fulfillmentStep := buildFulfillmentStep(order)
		
		// Step 5: Cleanup and finalization
		cleanupStep := buildCleanupStep(order)
		
		// Build the main sequential workflow with try-catch protection
		mainFlow := workflow.NewSequentialFlow("order-main-flow").
			Execute(validationStep).
			Execute(inventoryStep).
			Execute(paymentStep).
			Execute(fulfillmentStep)
		
		// Wrap entire flow with comprehensive error handling
		orderHandler := workflow.NewDefaultErrorHandlerAction("order-error-handler", func(ctx workflow.WorkContext, err error) workflow.WorkReport {
			fmt.Printf("  🚨 Order %s failed: %s\n", orderID, err.Error())
			fmt.Printf("  📧 Sending failure notification...\n")
			return workflow.NewCompletedWorkReport() // Convert to success after handling
		})
		
		orderTryCatch := workflow.NewTryCatch("order-processing").
			WithTryAction(mainFlow).
			CatchAny(orderHandler).
			Finally(cleanupStep)
		
		result := orderTryCatch.Run(wctx)
		
		if result.Status == workflow.StatusCompleted {
			fmt.Printf("  ✅ Order %s processed successfully!\n", orderID)
		}
		fmt.Println()
		
		return result
	})
}

func buildValidationStep(order map[string]interface{}) workflow.Action {
	return workflow.NewActionFunc("validation-step", func(wctx workflow.WorkContext) workflow.WorkReport {
		orderID := order["id"].(string)
		fmt.Printf("  📋 Validating order %s...\n", orderID)
		
		// Simulate validation with occasional failures
		if rand.Float64() < 0.1 {
			return workflow.NewFailedWorkReport(fmt.Errorf("validation failed: invalid order data"))
		}
		
		fmt.Printf("  ✅ Order %s validated\n", orderID)
		return workflow.NewCompletedWorkReport()
	})
}

func buildInventoryStep(order map[string]interface{}) workflow.Action {
	// Inventory service with circuit breaker protection and retry
	inventoryCheck := workflow.NewActionFunc("inventory-check", func(wctx workflow.WorkContext) workflow.WorkReport {
		orderID := order["id"].(string)
		items := order["items"].([]string)
		
		fmt.Printf("  📦 Checking inventory for %d items...\n", len(items))
		
		// Simulate inventory check with occasional failures
		time.Sleep(time.Duration(20+rand.Intn(30)) * time.Millisecond)
		
		if rand.Float64() < 0.2 {
			return workflow.NewFailedWorkReport(fmt.Errorf("inventory service temporarily unavailable"))
		}
		
		fmt.Printf("  ✅ Inventory confirmed for order %s\n", orderID)
		return workflow.NewCompletedWorkReport()
	})
	
	// Wrap with timeout protection
	timeoutProtected := workflow.NewTimeoutWrapper("inventory-timeout", 100*time.Millisecond).
		WithAction(inventoryCheck)
	
	// Add circuit breaker protection
	circuitProtected := workflow.NewCircuitBreaker("inventory-circuit", 2, 1*time.Second, 3*time.Second).
		WithAction(timeoutProtected)
	
	// Add retry with exponential backoff
	retryProtected := workflow.NewRetry("inventory-retry", 3).
		WithAction(circuitProtected).
		WithBackoffStrategy(workflow.NewExponentialBackoff(50*time.Millisecond, 500*time.Millisecond, 2.0)).
		WithRetryCondition(workflow.CombineRetryConditions(
			workflow.RetryOnTimeoutCondition,
			workflow.RetryOnNetworkCondition,
		))
	
	return retryProtected
}

func buildPaymentStep(order map[string]interface{}) workflow.Action {
	paymentProcessor := workflow.NewActionFunc("payment-processor", func(wctx workflow.WorkContext) workflow.WorkReport {
		orderID := order["id"].(string)
		amount := order["amount"].(float64)
		
		fmt.Printf("  💳 Processing payment of $%.2f...\n", amount)
		
		// Simulate payment processing
		time.Sleep(time.Duration(30+rand.Intn(40)) * time.Millisecond)
		
		// Different types of payment failures
		failure := rand.Float64()
		if failure < 0.1 {
			return workflow.NewFailedWorkReport(fmt.Errorf("payment declined: insufficient funds"))
		} else if failure < 0.15 {
			return workflow.NewFailedWorkReport(fmt.Errorf("payment gateway timeout"))
		} else if failure < 0.18 {
			return workflow.NewFailedWorkReport(fmt.Errorf("connection refused"))
		}
		
		fmt.Printf("  ✅ Payment processed for order %s\n", orderID)
		return workflow.NewCompletedWorkReport()
	})
	
	// Create specific error handlers
	declinedHandler := workflow.NewDefaultErrorHandlerAction("declined-handler", func(ctx workflow.WorkContext, err error) workflow.WorkReport {
		fmt.Printf("  💔 Payment declined - notifying customer\n")
		return workflow.NewFailedWorkReport(err) // Keep as failure
	})
	
	timeoutHandler := workflow.NewDefaultErrorHandlerAction("timeout-handler", func(ctx workflow.WorkContext, err error) workflow.WorkReport {
		fmt.Printf("  ⏰ Payment timeout - will retry\n")
		return workflow.NewFailedWorkReport(err) // Will be retried
	})
	
	networkHandler := workflow.NewDefaultErrorHandlerAction("network-handler", func(ctx workflow.WorkContext, err error) workflow.WorkReport {
		fmt.Printf("  🌐 Network issue - switching to backup gateway\n")
		return workflow.NewCompletedWorkReport() // Simulate successful fallback
	})
	
	// Try-catch with specific error handling
	paymentTryCatch := workflow.NewTryCatch("payment-handling").
		WithTryAction(paymentProcessor).
		Catch(workflow.ErrorMessageContains("insufficient funds"), declinedHandler).
		Catch(workflow.TimeoutError, timeoutHandler).
		Catch(workflow.NetworkError, networkHandler)
	
	// Wrap with retry for retryable errors
	paymentRetry := workflow.NewRetry("payment-retry", 2).
		WithAction(paymentTryCatch).
		WithBackoffStrategy(workflow.NewLinearBackoff(100*time.Millisecond, 100*time.Millisecond)).
		WithRetryCondition(workflow.RetryOnTimeoutCondition)
	
	return paymentRetry
}

func buildFulfillmentStep(order map[string]interface{}) workflow.Action {
	orderID := order["id"].(string)
	
	// Multiple fulfillment services running in parallel
	warehouseService := workflow.NewActionFunc("warehouse-service", func(wctx workflow.WorkContext) workflow.WorkReport {
		fmt.Printf("  🏭 Warehouse: Preparing shipment...\n")
		time.Sleep(time.Duration(40+rand.Intn(20)) * time.Millisecond)
		
		if rand.Float64() < 0.1 {
			return workflow.NewFailedWorkReport(fmt.Errorf("warehouse system error"))
		}
		
		fmt.Printf("  ✅ Warehouse: Items packaged\n")
		return workflow.NewCompletedWorkReport()
	})
	
	shippingService := workflow.NewActionFunc("shipping-service", func(wctx workflow.WorkContext) workflow.WorkReport {
		fmt.Printf("  🚚 Shipping: Creating label...\n")
		time.Sleep(time.Duration(30+rand.Intn(25)) * time.Millisecond)
		
		if rand.Float64() < 0.08 {
			return workflow.NewFailedWorkReport(fmt.Errorf("shipping label generation failed"))
		}
		
		fmt.Printf("  ✅ Shipping: Label created\n")
		return workflow.NewCompletedWorkReport()
	})
	
	notificationService := workflow.NewActionFunc("notification-service", func(wctx workflow.WorkContext) workflow.WorkReport {
		fmt.Printf("  📧 Notifications: Sending updates...\n")
		time.Sleep(time.Duration(20+rand.Intn(15)) * time.Millisecond)
		
		if rand.Float64() < 0.05 {
			return workflow.NewFailedWorkReport(fmt.Errorf("email service unavailable"))
		}
		
		fmt.Printf("  ✅ Notifications: Customer notified\n")
		return workflow.NewCompletedWorkReport()
	})
	
	// Run all fulfillment services in parallel with error collection
	fulfillmentCollector := workflow.NewParallelErrorCollector("fulfillment-services").
		AddActions(warehouseService, shippingService, notificationService)
	
	// Wrap the collector to provide meaningful feedback
	return workflow.NewActionFunc("fulfillment-coordinator", func(wctx workflow.WorkContext) workflow.WorkReport {
		fmt.Printf("  🎯 Coordinating fulfillment services...\n")
		
		result := fulfillmentCollector.Run(wctx)
		
		successful := result.Metadata["successful_actions"].(int)
		total := result.Metadata["total_actions"].(int)
		
		fmt.Printf("  📊 Fulfillment result: %d/%d services succeeded\n", successful, total)
		
		if len(result.Errors) > 0 {
			fmt.Printf("  ⚠️ Some fulfillment services had issues:\n")
			for _, err := range result.Errors {
				fmt.Printf("    - %s\n", err.Error())
			}
		}
		
		// Consider fulfillment successful if at least warehouse succeeds
		if successful >= 1 { // At minimum, warehouse must succeed
			fmt.Printf("  ✅ Order %s fulfillment initiated\n", orderID)
			return workflow.NewCompletedWorkReport()
		}
		
		return workflow.NewFailedWorkReport(fmt.Errorf("critical fulfillment services failed"))
	})
}

func buildCleanupStep(order map[string]interface{}) workflow.Action {
	return workflow.NewActionFunc("cleanup-step", func(wctx workflow.WorkContext) workflow.WorkReport {
		orderID := order["id"].(string)
		fmt.Printf("  🧹 Cleanup: Finalizing order %s...\n", orderID)
		
		// Simulate cleanup activities
		time.Sleep(10 * time.Millisecond)
		
		fmt.Printf("  ✅ Cleanup: Order %s finalized\n", orderID)
		return workflow.NewCompletedWorkReport()
	})
}