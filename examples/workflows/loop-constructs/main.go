package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ratlabs-io/go-agent-kit/pkg/constants"
	"github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

func main() {
	fmt.Println("=== Loop Constructs Example ===")
	
	ctx := context.Background()
	wctx := workflow.NewWorkContext(ctx)

	// Example 1: Count Loop - Process a batch of items
	fmt.Println("\n--- Example 1: Count Loop ---")
	fmt.Println("Processing 3 items in a batch...")
	
	batchProcessor := workflow.NewActionFunc("batch-processor", func(ctx workflow.WorkContext) workflow.WorkReport {
		iteration, _ := ctx.Get(constants.KeyLoopIteration)
		fmt.Printf("  Processing item %d...\n", iteration)
		return workflow.NewCompletedWorkReport()
	})
	
	countLoop := workflow.NewLoop("batch-process", 3).WithAction(batchProcessor)
	countLoop.Run(wctx)

	// Example 2: While Loop - Process until a condition is met
	fmt.Println("\n--- Example 2: While Loop ---")
	fmt.Println("Processing until we reach 100 points...")
	
	points := 0
	pointsAccumulator := workflow.NewActionFunc("points-accumulator", func(ctx workflow.WorkContext) workflow.WorkReport {
		iteration, _ := ctx.Get(constants.KeyLoopIteration)
		points += 25 // Earn 25 points per iteration
		fmt.Printf("  Iteration %d: Earned 25 points (total: %d)\n", iteration, points)
		return workflow.NewCompletedWorkReport()
	})
	
	whileCondition := func(ctx workflow.WorkContext) (bool, error) {
		return points < 100, nil
	}
	
	whileLoop := workflow.NewLoopWhile("accumulate-points", whileCondition).WithAction(pointsAccumulator)
	whileLoop.Run(wctx)

	// Example 3: Until Loop - Retry until success
	fmt.Println("\n--- Example 3: Until Loop ---")
	fmt.Println("Attempting to connect until successful...")
	
	attempts := 0
	connectionAttempt := workflow.NewActionFunc("connection-attempt", func(ctx workflow.WorkContext) workflow.WorkReport {
		iteration, _ := ctx.Get(constants.KeyLoopIteration)
		attempts++
		
		// Simulate success on 3rd attempt
		if attempts < 3 {
			fmt.Printf("  Attempt %d: Connection failed\n", iteration)
		} else {
			fmt.Printf("  Attempt %d: Connection successful!\n", iteration)
		}
		return workflow.NewCompletedWorkReport()
	})
	
	untilCondition := func(ctx workflow.WorkContext) (bool, error) {
		return attempts >= 3, nil // Stop when we've made 3 attempts
	}
	
	untilLoop := workflow.NewLoopUntil("connect-until-success", untilCondition).WithAction(connectionAttempt)
	untilLoop.Run(wctx)

	// Example 4: Loop Over Slice - Process each item in a collection
	fmt.Println("\n--- Example 4: Loop Over Slice ---")
	fmt.Println("Processing each file in the list...")
	
	files := []string{"config.yaml", "data.json", "report.pdf", "image.png"}
	
	fileProcessor := workflow.NewActionFunc("file-processor", func(ctx workflow.WorkContext) workflow.WorkReport {
		iteration, _ := ctx.Get(constants.KeyLoopIteration)
		index, _ := ctx.Get(constants.KeyCurrentIndex)
		item, _ := ctx.Get(constants.KeyCurrentItem)
		
		fmt.Printf("  Iteration %d: Processing file[%d] = %s\n", iteration, index, item)
		return workflow.NewCompletedWorkReport()
	})
	
	iterLoop := workflow.NewLoopOver("process-files", files).WithAction(fileProcessor)
	iterLoop.Run(wctx)

	// Example 5: Loop Over Map - Process key-value pairs
	fmt.Println("\n--- Example 5: Loop Over Map ---")
	fmt.Println("Processing configuration settings...")
	
	config := map[string]interface{}{
		"port":     8080,
		"debug":    true,
		"database": "postgresql://localhost:5432/mydb",
		"timeout":  30,
	}
	
	configProcessor := workflow.NewActionFunc("config-processor", func(ctx workflow.WorkContext) workflow.WorkReport {
		iteration, _ := ctx.Get(constants.KeyLoopIteration)
		key, _ := ctx.Get(constants.KeyCurrentIndex)     // For maps, index is the key
		value, _ := ctx.Get(constants.KeyCurrentItem)    // For maps, item is the value
		
		fmt.Printf("  Iteration %d: Setting %s = %v\n", iteration, key, value)
		return workflow.NewCompletedWorkReport()
	})
	
	mapLoop := workflow.NewLoopOver("process-config", config).WithAction(configProcessor)
	mapLoop.Run(wctx)

	// Example 6: Complex Loop Example - Nested processing
	fmt.Println("\n--- Example 6: Complex Loop with Sequential Flow ---")
	fmt.Println("Processing multiple batches of data...")
	
	batches := [][]string{
		{"user1", "user2"},
		{"user3", "user4", "user5"},
		{"user6"},
	}
	
	batchHeader := workflow.NewActionFunc("batch-header", func(ctx workflow.WorkContext) workflow.WorkReport {
		batchIndex, _ := ctx.Get(constants.KeyCurrentIndex)
		batch, _ := ctx.Get(constants.KeyCurrentItem)
		batchSlice := batch.([]string)
		fmt.Printf("  Processing batch %d with %d users\n", batchIndex, len(batchSlice))
		return workflow.NewCompletedWorkReport()
	})
	
	userProcessor := workflow.NewActionFunc("user-processor", func(ctx workflow.WorkContext) workflow.WorkReport {
		batchItem, _ := ctx.Get(constants.KeyCurrentItem)
		batch := batchItem.([]string)
		
		for i, user := range batch {
			fmt.Printf("    - User %d: %s\n", i+1, user)
		}
		return workflow.NewCompletedWorkReport()
	})
	
	// Create a sequential flow for each batch
	batchFlow := workflow.NewSequentialFlow("process-batch").
		Execute(batchHeader).
		Execute(userProcessor)
	
	complexLoop := workflow.NewLoopOver("process-all-batches", batches).WithAction(batchFlow)
	complexLoop.Run(wctx)

	fmt.Println("\n=== Loop Constructs Example Complete ===")
	fmt.Println("\nKey takeaways:")
	fmt.Println("- Count loops: Fixed number of iterations")
	fmt.Println("- While loops: Continue while condition is true") 
	fmt.Println("- Until loops: Continue until condition becomes true")
	fmt.Println("- Iterator loops: Process each item in collections")
	fmt.Println("- Context keys: current_item, current_index, loop_iteration")
	fmt.Println("- Composable: Loops can contain other workflows")
}