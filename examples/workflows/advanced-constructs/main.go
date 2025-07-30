package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

func main() {
	fmt.Println("=== Advanced Constructs Example ===")

	ctx := context.Background()
	wctx := workflow.NewWorkContext(ctx)

	// Example 1: Circuit Breaker - Protect against cascading failures
	fmt.Println("\n--- Example 1: Circuit Breaker ---")
	fmt.Println("Protecting against cascading failures...")

	unreliableService := workflow.NewActionFunc("unreliable-service", func(ctx workflow.WorkContext) workflow.WorkReport {
		// Simulate 80% failure rate
		if rand.Float64() < 0.8 {
			fmt.Println("  ❌ Service call failed")
			return workflow.NewFailedWorkReport(fmt.Errorf("service unavailable"))
		}
		fmt.Println("  ✅ Service call succeeded")
		return workflow.NewCompletedWorkReport()
	})

	// Circuit breaker: 3 failures triggers open state, 2s recovery timeout
	circuitBreaker := workflow.NewCircuitBreaker("service-breaker", 3, 2*time.Second, 5*time.Second).
		WithAction(unreliableService)

	// Make several calls to trigger the circuit breaker
	for i := 1; i <= 8; i++ {
		fmt.Printf("Call %d: ", i)
		report := circuitBreaker.Run(wctx)

		if report.Status == workflow.StatusFailure {
			if len(report.Errors) > 0 && report.Errors[0].Error() == "circuit breaker service-breaker is open" {
				fmt.Println("🚫 Circuit breaker is OPEN - request rejected")
			}
		}

		// Show circuit breaker metrics
		metrics := circuitBreaker.GetMetrics()
		fmt.Printf("    State: %v, Failures: %d/%d\n",
			metrics["state"], metrics["failures"], metrics["failure_threshold"])

		time.Sleep(500 * time.Millisecond)
	}

	// Example 2: Timeout Wrapper - Prevent operations from running too long
	fmt.Println("\n--- Example 2: Timeout Wrapper ---")
	fmt.Println("Preventing long-running operations...")

	slowOperation := workflow.NewActionFunc("slow-operation", func(ctx workflow.WorkContext) workflow.WorkReport {
		fmt.Println("  🐌 Starting slow operation...")

		// Simulate work that might take varying amounts of time
		duration := time.Duration(rand.Intn(300)) * time.Millisecond
		fmt.Printf("  ⏳ Working for %v...\n", duration)

		select {
		case <-ctx.Context().Done():
			fmt.Println("  ⏰ Operation was cancelled due to timeout")
			return workflow.NewFailedWorkReport(fmt.Errorf("operation cancelled"))
		case <-time.After(duration):
			fmt.Println("  ✅ Slow operation completed")
			return workflow.NewCompletedWorkReport()
		}
	})

	// Wrap with 200ms timeout
	timeoutWrapper := workflow.NewTimeoutWrapper("operation-timeout", 200*time.Millisecond).
		WithAction(slowOperation)

	for i := 1; i <= 4; i++ {
		fmt.Printf("Attempt %d: ", i)
		start := time.Now()
		report := timeoutWrapper.Run(wctx)
		elapsed := time.Since(start)

		if report.Status == workflow.StatusFailure {
			fmt.Printf("❌ Failed after %v\n", elapsed.Round(time.Millisecond))
		} else {
			fmt.Printf("✅ Succeeded after %v\n", elapsed.Round(time.Millisecond))
		}
	}

	// Example 3: Parallel Error Collector - Run multiple actions and collect all errors
	fmt.Println("\n--- Example 3: Parallel Error Collector ---")
	fmt.Println("Running multiple services and collecting all results...")

	// Create multiple services with different failure patterns
	services := []workflow.Action{
		workflow.NewActionFunc("auth-service", func(ctx workflow.WorkContext) workflow.WorkReport {
			time.Sleep(50 * time.Millisecond)
			if rand.Float64() < 0.3 {
				fmt.Println("  🔐 Auth service: ❌ Authentication failed")
				return workflow.NewFailedWorkReport(fmt.Errorf("authentication failed"))
			}
			fmt.Println("  🔐 Auth service: ✅ User authenticated")
			return workflow.NewCompletedWorkReport()
		}),

		workflow.NewActionFunc("payment-service", func(ctx workflow.WorkContext) workflow.WorkReport {
			time.Sleep(80 * time.Millisecond)
			if rand.Float64() < 0.4 {
				fmt.Println("  💳 Payment service: ❌ Payment declined")
				return workflow.NewFailedWorkReport(fmt.Errorf("payment declined"))
			}
			fmt.Println("  💳 Payment service: ✅ Payment processed")
			return workflow.NewCompletedWorkReport()
		}),

		workflow.NewActionFunc("inventory-service", func(ctx workflow.WorkContext) workflow.WorkReport {
			time.Sleep(30 * time.Millisecond)
			if rand.Float64() < 0.2 {
				fmt.Println("  📦 Inventory service: ❌ Out of stock")
				return workflow.NewFailedWorkReport(fmt.Errorf("item out of stock"))
			}
			fmt.Println("  📦 Inventory service: ✅ Item reserved")
			return workflow.NewCompletedWorkReport()
		}),

		workflow.NewActionFunc("notification-service", func(ctx workflow.WorkContext) workflow.WorkReport {
			time.Sleep(40 * time.Millisecond)
			if rand.Float64() < 0.1 {
				fmt.Println("  📧 Notification service: ❌ Email failed")
				return workflow.NewFailedWorkReport(fmt.Errorf("email delivery failed"))
			}
			fmt.Println("  📧 Notification service: ✅ Email sent")
			return workflow.NewCompletedWorkReport()
		}),
	}

	parallelCollector := workflow.NewParallelErrorCollector("service-health-check").
		AddActions(services...)

	start := time.Now()
	report := parallelCollector.Run(wctx)
	elapsed := time.Since(start)

	fmt.Printf("\n📊 Results after %v:\n", elapsed.Round(time.Millisecond))
	fmt.Printf("  Total services: %d\n", report.Metadata["total_actions"])
	fmt.Printf("  Successful: %d\n", report.Metadata["successful_actions"])
	fmt.Printf("  Failed: %d\n", report.Metadata["failed_actions"])
	fmt.Printf("  Total errors: %d\n", report.Metadata["total_errors"])

	if len(report.Errors) > 0 {
		fmt.Println("  🚨 Errors encountered:")
		for i, err := range report.Errors {
			fmt.Printf("    %d. %s\n", i+1, err.Error())
		}
	}

	if report.Status == workflow.StatusCompleted {
		fmt.Println("  🎉 Overall status: All services healthy!")
	} else {
		fmt.Println("  ⚠️ Overall status: Some services have issues")
	}

	// Example 4: Combined Advanced Patterns
	fmt.Println("\n--- Example 4: Combined Advanced Patterns ---")
	fmt.Println("Circuit breaker + timeout + error collection...")

	// Create a protected service (circuit breaker + timeout)
	flakeyService := workflow.NewActionFunc("flakey-service", func(ctx workflow.WorkContext) workflow.WorkReport {
		// Random delay between 50-250ms
		delay := time.Duration(50+rand.Intn(200)) * time.Millisecond

		select {
		case <-ctx.Context().Done():
			return workflow.NewFailedWorkReport(fmt.Errorf("service timeout"))
		case <-time.After(delay):
			// 60% failure rate
			if rand.Float64() < 0.6 {
				return workflow.NewFailedWorkReport(fmt.Errorf("service error"))
			}
			return workflow.NewCompletedWorkReport()
		}
	})

	// Wrap with timeout first, then circuit breaker
	timeoutProtected := workflow.NewTimeoutWrapper("service-timeout", 150*time.Millisecond).
		WithAction(flakeyService)

	circuitProtected := workflow.NewCircuitBreaker("service-circuit", 2, 1*time.Second, 3*time.Second).
		WithAction(timeoutProtected)

	// Create multiple protected services
	protectedServices := []workflow.Action{
		circuitProtected,
		workflow.NewActionFunc("backup-service", func(ctx workflow.WorkContext) workflow.WorkReport {
			time.Sleep(20 * time.Millisecond)
			// Backup is more reliable
			if rand.Float64() < 0.1 {
				return workflow.NewFailedWorkReport(fmt.Errorf("backup service error"))
			}
			return workflow.NewCompletedWorkReport()
		}),
	}

	// Run them in parallel and collect all results
	advancedCollector := workflow.NewParallelErrorCollector("advanced-service-check").
		AddActions(protectedServices...)

	fmt.Println("Running protected services...")
	for i := 1; i <= 3; i++ {
		fmt.Printf("\nRound %d:\n", i)
		start := time.Now()
		report := advancedCollector.Run(wctx)
		elapsed := time.Since(start)

		fmt.Printf("  Completed in %v\n", elapsed.Round(time.Millisecond))
		fmt.Printf("  Success rate: %d/%d\n",
			report.Metadata["successful_actions"], report.Metadata["total_actions"])

		if len(report.Errors) > 0 {
			for _, err := range report.Errors {
				fmt.Printf("    ⚠️ %s\n", err.Error())
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	// Example 5: Health Check System
	fmt.Println("\n--- Example 5: Comprehensive Health Check System ---")
	fmt.Println("Building a robust health check system...")

	// Database health check with circuit breaker
	dbCheck := workflow.NewActionFunc("database-check", func(ctx workflow.WorkContext) workflow.WorkReport {
		time.Sleep(time.Duration(20+rand.Intn(30)) * time.Millisecond)
		if rand.Float64() < 0.15 {
			return workflow.NewFailedWorkReport(fmt.Errorf("database connection failed"))
		}
		return workflow.NewCompletedWorkReport()
	})

	dbCircuit := workflow.NewCircuitBreaker("db-circuit", 3, 2*time.Second, 5*time.Second).
		WithAction(dbCheck)

	dbTimeout := workflow.NewTimeoutWrapper("db-timeout", 100*time.Millisecond).
		WithAction(dbCircuit)

	// Cache health check with timeout
	cacheCheck := workflow.NewActionFunc("cache-check", func(ctx workflow.WorkContext) workflow.WorkReport {
		time.Sleep(time.Duration(10+rand.Intn(20)) * time.Millisecond)
		if rand.Float64() < 0.1 {
			return workflow.NewFailedWorkReport(fmt.Errorf("cache connection failed"))
		}
		return workflow.NewCompletedWorkReport()
	})

	cacheTimeout := workflow.NewTimeoutWrapper("cache-timeout", 50*time.Millisecond).
		WithAction(cacheCheck)

	// External API health check
	apiCheck := workflow.NewActionFunc("api-check", func(ctx workflow.WorkContext) workflow.WorkReport {
		time.Sleep(time.Duration(30+rand.Intn(50)) * time.Millisecond)
		if rand.Float64() < 0.2 {
			return workflow.NewFailedWorkReport(fmt.Errorf("external API unreachable"))
		}
		return workflow.NewCompletedWorkReport()
	})

	apiTimeout := workflow.NewTimeoutWrapper("api-timeout", 150*time.Millisecond).
		WithAction(apiCheck)

	// Combine all health checks
	healthChecks := []workflow.Action{dbTimeout, cacheTimeout, apiTimeout}

	healthSystem := workflow.NewParallelErrorCollector("system-health").
		AddActions(healthChecks...)

	fmt.Println("Running comprehensive health checks...")
	healthReport := healthSystem.Run(wctx)

	fmt.Println("\n🏥 System Health Report:")
	fmt.Printf("  Total checks: %d\n", healthReport.Metadata["total_actions"])
	fmt.Printf("  Passed: %d\n", healthReport.Metadata["successful_actions"])
	fmt.Printf("  Failed: %d\n", healthReport.Metadata["failed_actions"])

	healthScore := float64(healthReport.Metadata["successful_actions"].(int)) /
		float64(healthReport.Metadata["total_actions"].(int)) * 100

	fmt.Printf("  Health score: %.1f%%\n", healthScore)

	if healthScore >= 100 {
		fmt.Println("  Status: 🟢 All systems operational")
	} else if healthScore >= 66 {
		fmt.Println("  Status: 🟡 Some systems degraded")
	} else {
		fmt.Println("  Status: 🔴 Critical systems down")
	}

	fmt.Println("\n=== Advanced Constructs Example Complete ===")
	fmt.Println("\nKey takeaways:")
	fmt.Println("- Circuit breakers: Prevent cascading failures with state tracking")
	fmt.Println("- Timeout wrappers: Protect against long-running operations")
	fmt.Println("- Parallel error collectors: Comprehensive error reporting")
	fmt.Println("- Composability: Combine patterns for robust systems")
	fmt.Println("- Health monitoring: Build comprehensive system health checks")
}
