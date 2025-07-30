package main

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

func main() {
	fmt.Println("=== Error Handling Example ===")

	ctx := context.Background()
	wctx := workflow.NewWorkContext(ctx)

	// Example 1: Basic Try-Catch
	fmt.Println("\n--- Example 1: Basic Try-Catch ---")
	fmt.Println("Simple error handling with catch-all...")

	riskyOperation := workflow.NewActionFunc("risky-operation", func(ctx workflow.WorkContext) workflow.WorkReport {
		fmt.Println("  Attempting risky operation...")
		// Simulate random failure
		if rand.Float64() < 0.6 {
			fmt.Println("  ❌ Operation failed!")
			return workflow.NewFailedWorkReport(fmt.Errorf("operation failed unexpectedly"))
		}
		fmt.Println("  ✅ Operation succeeded!")
		return workflow.NewCompletedWorkReport()
	})

	errorHandler := workflow.NewDefaultErrorHandlerAction("error-handler", func(ctx workflow.WorkContext, err error) workflow.WorkReport {
		fmt.Printf("  🔧 Caught error: %s\n", err.Error())
		fmt.Println("  📝 Logging error and continuing...")
		return workflow.NewCompletedWorkReport()
	})

	basicTryCatch := workflow.NewTryCatch("basic-error-handling").
		WithTryAction(riskyOperation).
		CatchAny(errorHandler)

	basicTryCatch.Run(wctx)

	// Example 2: Specific Error Type Handling
	fmt.Println("\n--- Example 2: Specific Error Type Handling ---")
	fmt.Println("Different handlers for different error types...")

	fileOperation := workflow.NewActionFunc("file-operation", func(ctx workflow.WorkContext) workflow.WorkReport {
		fmt.Println("  Attempting file operation...")

		errorTypes := []string{"timeout", "validation", "network", "success"}
		choice := errorTypes[rand.Intn(len(errorTypes))]

		switch choice {
		case "timeout":
			fmt.Println("  ❌ Connection timeout!")
			return workflow.NewFailedWorkReport(fmt.Errorf("connection timeout"))
		case "validation":
			fmt.Println("  ❌ Validation failed!")
			return workflow.NewFailedWorkReport(fmt.Errorf("validation error: invalid input"))
		case "network":
			fmt.Println("  ❌ Network error!")
			return workflow.NewFailedWorkReport(fmt.Errorf("connection refused"))
		default:
			fmt.Println("  ✅ File operation succeeded!")
			return workflow.NewCompletedWorkReport()
		}
	})

	timeoutHandler := workflow.NewDefaultErrorHandlerAction("timeout-handler", func(ctx workflow.WorkContext, err error) workflow.WorkReport {
		fmt.Printf("  ⏰ Handling timeout: %s\n", err.Error())
		fmt.Println("  🔄 Increasing timeout and will retry later...")
		return workflow.NewCompletedWorkReport()
	})

	validationHandler := workflow.NewDefaultErrorHandlerAction("validation-handler", func(ctx workflow.WorkContext, err error) workflow.WorkReport {
		fmt.Printf("  ✅ Handling validation error: %s\n", err.Error())
		fmt.Println("  📋 Prompting user for correct input...")
		return workflow.NewCompletedWorkReport()
	})

	networkHandler := workflow.NewDefaultErrorHandlerAction("network-handler", func(ctx workflow.WorkContext, err error) workflow.WorkReport {
		fmt.Printf("  🌐 Handling network error: %s\n", err.Error())
		fmt.Println("  🔄 Switching to backup connection...")
		return workflow.NewCompletedWorkReport()
	})

	specificTryCatch := workflow.NewTryCatch("specific-error-handling").
		WithTryAction(fileOperation).
		Catch(workflow.TimeoutError, timeoutHandler).
		Catch(workflow.ValidationError, validationHandler).
		Catch(workflow.NetworkError, networkHandler)

	specificTryCatch.Run(wctx)

	// Example 3: Try-Catch-Finally
	fmt.Println("\n--- Example 3: Try-Catch-Finally ---")
	fmt.Println("Always execute cleanup code...")

	resourceOperation := workflow.NewActionFunc("resource-operation", func(ctx workflow.WorkContext) workflow.WorkReport {
		fmt.Println("  🔓 Acquiring resources...")
		fmt.Println("  📊 Processing data...")

		// Simulate failure 50% of the time
		if rand.Float64() < 0.5 {
			fmt.Println("  ❌ Processing failed!")
			return workflow.NewFailedWorkReport(fmt.Errorf("data processing failed"))
		}

		fmt.Println("  ✅ Processing completed!")
		return workflow.NewCompletedWorkReport()
	})

	resourceHandler := workflow.NewDefaultErrorHandlerAction("resource-error-handler", func(ctx workflow.WorkContext, err error) workflow.WorkReport {
		fmt.Printf("  🛠️ Handling processing error: %s\n", err.Error())
		fmt.Println("  📁 Saving partial results...")
		return workflow.NewCompletedWorkReport()
	})

	cleanupAction := workflow.NewActionFunc("cleanup", func(ctx workflow.WorkContext) workflow.WorkReport {
		fmt.Println("  🧹 Cleaning up resources...")
		fmt.Println("  🔒 Releasing locks...")
		fmt.Println("  📋 Updating status...")
		return workflow.NewCompletedWorkReport()
	})

	finallyTryCatch := workflow.NewTryCatch("resource-handling").
		WithTryAction(resourceOperation).
		CatchAny(resourceHandler).
		Finally(cleanupAction)

	finallyTryCatch.Run(wctx)

	// Example 4: Nested Try-Catch
	fmt.Println("\n--- Example 4: Nested Try-Catch ---")
	fmt.Println("Nested error handling for complex operations...")

	databaseOp := workflow.NewActionFunc("database-operation", func(ctx workflow.WorkContext) workflow.WorkReport {
		fmt.Println("    🗄️ Executing database query...")
		if rand.Float64() < 0.4 {
			return workflow.NewFailedWorkReport(fmt.Errorf("database connection failed"))
		}
		fmt.Println("    ✅ Database query successful!")
		return workflow.NewCompletedWorkReport()
	})

	dbErrorHandler := workflow.NewDefaultErrorHandlerAction("db-error-handler", func(ctx workflow.WorkContext, err error) workflow.WorkReport {
		fmt.Printf("    🔧 Database error caught: %s\n", err.Error())
		return workflow.NewCompletedWorkReport()
	})

	innerTryCatch := workflow.NewTryCatch("database-handling").
		WithTryAction(databaseOp).
		CatchAny(dbErrorHandler)

	apiCall := workflow.NewActionFunc("api-call", func(ctx workflow.WorkContext) workflow.WorkReport {
		fmt.Println("  🌐 Making API call...")
		if rand.Float64() < 0.3 {
			return workflow.NewFailedWorkReport(fmt.Errorf("API rate limit exceeded"))
		}
		fmt.Println("  ✅ API call successful!")
		return workflow.NewCompletedWorkReport()
	})

	outerFlow := workflow.NewSequentialFlow("complex-operation").
		Execute(apiCall).
		Execute(innerTryCatch) // Nested try-catch

	outerErrorHandler := workflow.NewDefaultErrorHandlerAction("outer-error-handler", func(ctx workflow.WorkContext, err error) workflow.WorkReport {
		fmt.Printf("  🚨 Outer error caught: %s\n", err.Error())
		fmt.Println("  📧 Sending alert to administrators...")
		return workflow.NewCompletedWorkReport()
	})

	nestedTryCatch := workflow.NewTryCatch("complex-error-handling").
		WithTryAction(outerFlow).
		CatchAny(outerErrorHandler)

	nestedTryCatch.Run(wctx)

	// Example 5: Error Recovery and Transformation
	fmt.Println("\n--- Example 5: Error Recovery and Transformation ---")
	fmt.Println("Transform errors into successful results...")

	unreliableService := workflow.NewActionFunc("unreliable-service", func(ctx workflow.WorkContext) workflow.WorkReport {
		fmt.Println("  📡 Calling primary service...")
		// Always fail for demonstration
		return workflow.NewFailedWorkReport(fmt.Errorf("primary service unavailable"))
	})

	fallbackHandler := workflow.NewDefaultErrorHandlerAction("fallback-handler", func(ctx workflow.WorkContext, err error) workflow.WorkReport {
		fmt.Printf("  🔄 Primary failed: %s\n", err.Error())
		fmt.Println("  🏥 Switching to backup service...")
		fmt.Println("  ✅ Backup service responded!")

		// Transform error into success with fallback data
		report := workflow.NewCompletedWorkReport()
		report.Data = "fallback data"
		return report
	})

	recoveryTryCatch := workflow.NewTryCatch("service-recovery").
		WithTryAction(unreliableService).
		CatchAny(fallbackHandler)

	result := recoveryTryCatch.Run(wctx)
	if result.Status == workflow.StatusCompleted && result.Data != nil {
		fmt.Printf("  💾 Final result: %s\n", result.Data)
	}

	// Example 6: Complex Error Classification
	fmt.Println("\n--- Example 6: Complex Error Classification ---")
	fmt.Println("Advanced error matching and routing...")

	complexOperation := workflow.NewActionFunc("complex-operation", func(ctx workflow.WorkContext) workflow.WorkReport {
		fmt.Println("  ⚡ Executing complex operation...")

		errors := []string{
			"connection timeout occurred",
			"invalid input format provided",
			"rate limit: too many requests",
			"connection refused by server",
			"success",
		}

		choice := errors[rand.Intn(len(errors))]
		if choice == "success" {
			fmt.Println("  ✅ Complex operation succeeded!")
			return workflow.NewCompletedWorkReport()
		}

		fmt.Printf("  ❌ Error: %s\n", choice)
		return workflow.NewFailedWorkReport(fmt.Errorf("%s", choice))
	})

	// Create combined error matchers
	retryableErrors := workflow.CombineErrorMatchers(
		workflow.TimeoutError,
		workflow.NetworkError,
	)

	retryableHandler := workflow.NewDefaultErrorHandlerAction("retryable-handler", func(ctx workflow.WorkContext, err error) workflow.WorkReport {
		fmt.Printf("  🔄 Retryable error: %s\n", err.Error())
		fmt.Println("  📋 Scheduling for retry...")
		return workflow.NewCompletedWorkReport()
	})

	rateLimitHandler := workflow.NewDefaultErrorHandlerAction("rate-limit-handler", func(ctx workflow.WorkContext, err error) workflow.WorkReport {
		fmt.Printf("  ⏳ Rate limit error: %s\n", err.Error())
		fmt.Println("  😴 Backing off for extended period...")
		return workflow.NewCompletedWorkReport()
	})

	fatalHandler := workflow.NewDefaultErrorHandlerAction("fatal-handler", func(ctx workflow.WorkContext, err error) workflow.WorkReport {
		fmt.Printf("  💀 Fatal error: %s\n", err.Error())
		fmt.Println("  🚨 Alerting support team...")
		return workflow.NewFailedWorkReport(err) // Keep as failure
	})

	classificationTryCatch := workflow.NewTryCatch("error-classification").
		WithTryAction(complexOperation).
		Catch(retryableErrors, retryableHandler).
		Catch(workflow.ErrorMessageContains("rate limit"), rateLimitHandler).
		Catch(workflow.ValidationError, fatalHandler).
		CatchAny(fatalHandler) // Catch anything else as fatal

	classificationTryCatch.Run(wctx)

	fmt.Println("\n=== Error Handling Example Complete ===")
	fmt.Println("\nKey takeaways:")
	fmt.Println("- Try-catch-finally: Structured exception handling")
	fmt.Println("- Error type matching: Route errors to specific handlers")
	fmt.Println("- Error recovery: Transform failures into successes")
	fmt.Println("- Error classification: Complex routing with combined matchers")
	fmt.Println("- Composability: Nest try-catch blocks in workflows")
	fmt.Println("- Finally blocks: Always execute cleanup code")
}
