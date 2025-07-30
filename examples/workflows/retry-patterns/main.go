package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

func main() {
	fmt.Println("=== Retry Patterns Example ===")

	ctx := context.Background()
	wctx := workflow.NewWorkContext(ctx)

	// Example 1: Basic Retry with Fixed Backoff
	fmt.Println("\n--- Example 1: Basic Retry with Fixed Backoff ---")
	fmt.Println("Attempting API call with fixed 100ms delays...")

	attempt1 := 0
	apiCall := workflow.NewActionFunc("api-call", func(ctx workflow.WorkContext) workflow.WorkReport {
		attempt1++
		fmt.Printf("  Attempt %d: Making API call...", attempt1)

		// Simulate 70% failure rate
		if rand.Float64() < 0.7 {
			fmt.Println(" ❌ Failed (network error)")
			return workflow.NewFailedWorkReport(fmt.Errorf("network error"))
		}

		fmt.Println(" ✅ Success!")
		return workflow.NewCompletedWorkReport()
	})

	basicRetry := workflow.NewRetry("api-retry", 5).
		WithAction(apiCall).
		WithBackoffStrategy(workflow.NewFixedBackoff(100 * time.Millisecond))

	basicRetry.Run(wctx)

	// Example 2: Linear Backoff Strategy
	fmt.Println("\n--- Example 2: Linear Backoff Strategy ---")
	fmt.Println("Database connection with increasing delays (100ms, 200ms, 300ms...)...")

	attempt2 := 0
	dbConnect := workflow.NewActionFunc("db-connect", func(ctx workflow.WorkContext) workflow.WorkReport {
		attempt2++
		fmt.Printf("  Attempt %d: Connecting to database...", attempt2)

		// Succeed after 3 attempts
		if attempt2 < 3 {
			fmt.Println(" ❌ Failed (connection refused)")
			return workflow.NewFailedWorkReport(fmt.Errorf("connection refused"))
		}

		fmt.Println(" ✅ Connected!")
		return workflow.NewCompletedWorkReport()
	})

	linearRetry := workflow.NewRetry("db-retry", 5).
		WithAction(dbConnect).
		WithBackoffStrategy(workflow.NewLinearBackoff(100*time.Millisecond, 100*time.Millisecond))

	linearRetry.Run(wctx)

	// Example 3: Exponential Backoff Strategy
	fmt.Println("\n--- Example 3: Exponential Backoff Strategy ---")
	fmt.Println("External service with exponential backoff (100ms, 200ms, 400ms...)...")

	attempt3 := 0
	serviceCall := workflow.NewActionFunc("service-call", func(ctx workflow.WorkContext) workflow.WorkReport {
		attempt3++
		fmt.Printf("  Attempt %d: Calling external service...", attempt3)

		// Succeed after 3 attempts
		if attempt3 < 3 {
			fmt.Println(" ❌ Failed (service unavailable)")
			return workflow.NewFailedWorkReport(fmt.Errorf("service unavailable"))
		}

		fmt.Println(" ✅ Service responded!")
		return workflow.NewCompletedWorkReport()
	})

	exponentialRetry := workflow.NewRetry("service-retry", 5).
		WithAction(serviceCall).
		WithBackoffStrategy(workflow.NewExponentialBackoff(100*time.Millisecond, 2*time.Second, 2.0))

	exponentialRetry.Run(wctx)

	// Example 4: Custom Retry Conditions
	fmt.Println("\n--- Example 4: Custom Retry Conditions ---")
	fmt.Println("Only retry on timeout errors, not validation errors...")

	attempt4 := 0
	errors := []error{
		fmt.Errorf("validation error: invalid input"),
		fmt.Errorf("connection timeout"),
		fmt.Errorf("connection timeout"),
		nil, // success
	}

	conditionalCall := workflow.NewActionFunc("conditional-call", func(ctx workflow.WorkContext) workflow.WorkReport {
		if attempt4 < len(errors) {
			err := errors[attempt4]
			attempt4++

			if err != nil {
				fmt.Printf("  Attempt %d: ❌ %s\n", attempt4, err.Error())
				return workflow.NewFailedWorkReport(err)
			}
		}

		fmt.Printf("  Attempt %d: ✅ Success!\n", attempt4)
		return workflow.NewCompletedWorkReport()
	})

	conditionalRetry := workflow.NewRetry("conditional-retry", 5).
		WithAction(conditionalCall).
		WithRetryCondition(workflow.RetryOnTimeoutCondition). // Only retry timeouts
		WithBackoffStrategy(workflow.NewFixedBackoff(50 * time.Millisecond))

	conditionalRetry.Run(wctx)

	// Example 5: Combined Retry Conditions
	fmt.Println("\n--- Example 5: Combined Retry Conditions ---")
	fmt.Println("Retry on timeouts OR network errors...")

	attempt5 := 0
	networkErrors := []error{
		fmt.Errorf("connection refused"),
		fmt.Errorf("rate limit exceeded"),
		fmt.Errorf("connection timeout"),
		nil, // success
	}

	networkCall := workflow.NewActionFunc("network-call", func(ctx workflow.WorkContext) workflow.WorkReport {
		if attempt5 < len(networkErrors) {
			err := networkErrors[attempt5]
			attempt5++

			if err != nil {
				fmt.Printf("  Attempt %d: ❌ %s\n", attempt5, err.Error())
				return workflow.NewFailedWorkReport(err)
			}
		}

		fmt.Printf("  Attempt %d: ✅ Success!\n", attempt5)
		return workflow.NewCompletedWorkReport()
	})

	combinedCondition := workflow.CombineRetryConditions(
		workflow.RetryOnTimeoutCondition,
		workflow.RetryOnNetworkCondition,
		workflow.RetryOnRateLimitCondition,
	)

	combinedRetry := workflow.NewRetry("combined-retry", 6).
		WithAction(networkCall).
		WithRetryCondition(combinedCondition).
		WithBackoffStrategy(workflow.NewLinearBackoff(25*time.Millisecond, 25*time.Millisecond))

	combinedRetry.Run(wctx)

	// Example 6: Stop Condition
	fmt.Println("\n--- Example 6: Stop Condition Based on Context ---")
	fmt.Println("Stop retrying if total time exceeds threshold...")

	startTime := time.Now()
	attempt6 := 0
	timedCall := workflow.NewActionFunc("timed-call", func(ctx workflow.WorkContext) workflow.WorkReport {
		attempt6++
		elapsed := time.Since(startTime)
		fmt.Printf("  Attempt %d (after %v): Processing...", attempt6, elapsed.Round(time.Millisecond))

		// Always fail for demonstration
		fmt.Println(" ❌ Failed")
		return workflow.NewFailedWorkReport(fmt.Errorf("processing failed"))
	})

	stopCondition := func(ctx workflow.WorkContext) bool {
		elapsed := time.Since(startTime)
		return elapsed > 300*time.Millisecond // Stop after 300ms
	}

	timedRetry := workflow.NewRetry("timed-retry", 10).
		WithAction(timedCall).
		WithStopCondition(stopCondition).
		WithBackoffStrategy(workflow.NewFixedBackoff(100 * time.Millisecond))

	timedRetry.Run(wctx)

	// Example 7: Complex Retry Pattern - File Processing
	fmt.Println("\n--- Example 7: Complex Retry Pattern ---")
	fmt.Println("File processing with sophisticated retry logic...")

	files := []string{"config.yaml", "data.json", "large-file.zip"}
	processedFiles := make(map[string]bool)

	fileProcessor := workflow.NewActionFunc("file-processor", func(ctx workflow.WorkContext) workflow.WorkReport {
		// Simulate processing all files
		for _, file := range files {
			if !processedFiles[file] {
				fmt.Printf("  Processing %s...", file)

				// Simulate occasional failures
				if file == "large-file.zip" && len(processedFiles) < 2 {
					fmt.Println(" ❌ Failed (disk full)")
					return workflow.NewFailedWorkReport(fmt.Errorf("disk full"))
				}

				processedFiles[file] = true
				fmt.Println(" ✅ Processed")
			}
		}

		if len(processedFiles) == len(files) {
			fmt.Println("  ✅ All files processed successfully!")
			return workflow.NewCompletedWorkReport()
		}

		return workflow.NewFailedWorkReport(fmt.Errorf("processing incomplete"))
	})

	fileRetry := workflow.NewRetry("file-processing-retry", 4).
		WithAction(fileProcessor).
		WithBackoffStrategy(workflow.NewExponentialBackoff(50*time.Millisecond, 500*time.Millisecond, 1.5)).
		WithRetryCondition(workflow.DefaultRetryCondition)

	fileRetry.Run(wctx)

	fmt.Println("\n=== Retry Patterns Example Complete ===")
	fmt.Println("\nKey takeaways:")
	fmt.Println("- Fixed backoff: Consistent delays between attempts")
	fmt.Println("- Linear backoff: Steadily increasing delays")
	fmt.Println("- Exponential backoff: Rapidly increasing delays with cap")
	fmt.Println("- Retry conditions: Control when to retry based on error type")
	fmt.Println("- Stop conditions: Halt retries based on context state")
	fmt.Println("- Composability: Combine conditions for sophisticated logic")
}
