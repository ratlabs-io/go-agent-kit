package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

// ActionFuncExample demonstrates creating simple actions on the fly
func main() {
	fmt.Println("=== ActionFunc Example ===")
	fmt.Println("This example shows how to create simple actions without boilerplate")
	fmt.Println()

	// Create simple actions using ActionFunc - much cleaner!
	preprocessAction := workflow.NewActionFunc("preprocess", func(ctx workflow.WorkContext) workflow.WorkReport {
		if text, ok := ctx.Get("input_text"); ok {
			if textStr, ok := text.(string); ok {
				processed := strings.TrimSpace(strings.ToLower(textStr))
				ctx.Set("processed_text", processed)
				fmt.Printf("📝 Preprocessed: '%s' → '%s'\n", textStr, processed)
				return workflow.NewCompletedWorkReport()
			}
		}
		return workflow.NewFailedWorkReport(fmt.Errorf("no input_text provided"))
	})

	validateAction := workflow.NewActionFunc("validate", func(ctx workflow.WorkContext) workflow.WorkReport {
		if text, ok := ctx.Get("processed_text"); ok {
			if textStr, ok := text.(string); ok {
				if len(textStr) < 5 {
					fmt.Println("❌ Validation failed: text too short")
					return workflow.NewFailedWorkReport(fmt.Errorf("text must be at least 5 characters"))
				}
				fmt.Printf("✅ Validation passed: text length = %d\n", len(textStr))
				ctx.Set("validated", true)
				return workflow.NewCompletedWorkReport()
			}
		}
		return workflow.NewFailedWorkReport(fmt.Errorf("no processed_text provided"))
	})

	transformAction := workflow.NewActionFunc("transform", func(ctx workflow.WorkContext) workflow.WorkReport {
		if validated, ok := ctx.Get("validated"); ok && validated.(bool) {
			if text, ok := ctx.Get("processed_text"); ok {
				if textStr, ok := text.(string); ok {
					// Simple transformation: reverse the string
					runes := []rune(textStr)
					for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
						runes[i], runes[j] = runes[j], runes[i]
					}
					result := string(runes)
					ctx.Set("result", result)
					fmt.Printf("🔄 Transformed: '%s' → '%s'\n", textStr, result)
					return workflow.NewCompletedWorkReport()
				}
			}
		}
		return workflow.NewFailedWorkReport(fmt.Errorf("validation not passed or no processed_text"))
	})

	logAction := workflow.NewActionFunc("log", func(ctx workflow.WorkContext) workflow.WorkReport {
		fmt.Println("\n📊 Context State:")
		fmt.Println("----------------")
		for _, key := range []string{"input_text", "processed_text", "validated", "result"} {
			if val, ok := ctx.Get(key); ok {
				fmt.Printf("  %s: %v\n", key, val)
			}
		}
		return workflow.NewCompletedWorkReport()
	})

	// Create sequential workflow
	pipeline := workflow.NewSequentialFlow("text-processing-pipeline").
		Then(preprocessAction).
		Then(validateAction).
		Then(transformAction).
		Then(logAction)

	// Test with valid input
	fmt.Println("Test 1: Valid input")
	fmt.Println("-------------------")
	ctx := context.Background()
	wctx1 := workflow.NewWorkContext(ctx)
	wctx1.Set("input_text", "  Hello World  ")

	report1 := pipeline.Run(wctx1)
	if report1.Status == workflow.StatusCompleted {
		fmt.Println("\n✅ Pipeline completed successfully!")
	} else {
		fmt.Printf("\n❌ Pipeline failed: %v\n", report1.Errors)
	}

	// Test with invalid input
	fmt.Println("\n\nTest 2: Invalid input (too short)")
	fmt.Println("----------------------------------")
	wctx2 := workflow.NewWorkContext(ctx)
	wctx2.Set("input_text", "Hi")

	report2 := pipeline.Run(wctx2)
	if report2.Status == workflow.StatusCompleted {
		fmt.Println("\n✅ Pipeline completed successfully!")
	} else {
		fmt.Printf("\n❌ Pipeline failed: %v\n", report2.Errors)
	}

	// Example of parallel actions
	fmt.Println("\n\nTest 3: Parallel actions")
	fmt.Println("------------------------")

	// Create parallel actions that simulate async operations
	asyncAction1 := workflow.NewActionFunc("async1", func(ctx workflow.WorkContext) workflow.WorkReport {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("⚡ Async action 1 completed")
		ctx.Set("async1_result", "data from async1")
		return workflow.NewCompletedWorkReport()
	})

	asyncAction2 := workflow.NewActionFunc("async2", func(ctx workflow.WorkContext) workflow.WorkReport {
		time.Sleep(50 * time.Millisecond)
		fmt.Println("⚡ Async action 2 completed")
		ctx.Set("async2_result", "data from async2")
		return workflow.NewCompletedWorkReport()
	})

	asyncAction3 := workflow.NewActionFunc("async3", func(ctx workflow.WorkContext) workflow.WorkReport {
		time.Sleep(75 * time.Millisecond)
		fmt.Println("⚡ Async action 3 completed")
		ctx.Set("async3_result", "data from async3")
		return workflow.NewCompletedWorkReport()
	})

	combineAction := workflow.NewActionFunc("combine", func(ctx workflow.WorkContext) workflow.WorkReport {
		var results []string
		for i := 1; i <= 3; i++ {
			if result, ok := ctx.Get(fmt.Sprintf("async%d_result", i)); ok {
				results = append(results, result.(string))
			}
		}
		combined := strings.Join(results, " + ")
		fmt.Printf("🔗 Combined results: %s\n", combined)
		return workflow.NewCompletedWorkReport()
	})

	// Create workflow with parallel execution followed by combination
	parallelFlow := workflow.NewSequentialFlow("parallel-then-combine").
		Then(workflow.NewParallelFlow("async-operations").
			Execute(asyncAction1).
			Execute(asyncAction2).
			Execute(asyncAction3)).
		Then(combineAction)

	wctx3 := workflow.NewWorkContext(ctx)
	report3 := parallelFlow.Run(wctx3)

	if report3.Status == workflow.StatusCompleted {
		fmt.Println("✅ Parallel workflow completed successfully!")
	} else {
		fmt.Printf("❌ Parallel workflow failed: %v\n", report3.Errors)
	}
}
