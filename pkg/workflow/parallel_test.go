package workflow

import (
	"context"
	"testing"
	"time"
)

func TestParallelFlow_Success(t *testing.T) {
	// Create mock actions with delays to ensure parallel execution
	action1 := &MockAction{name: "action1", result: "result1", delay: 10 * time.Millisecond}
	action2 := &MockAction{name: "action2", result: "result2", delay: 20 * time.Millisecond}
	action3 := &MockAction{name: "action3", result: "result3", delay: 15 * time.Millisecond}
	
	// Create parallel flow
	flow := NewParallelFlow("test-parallel-flow").
		Execute(action1).
		Execute(action2).
		Execute(action3)
	
	// Create context
	ctx := NewWorkContext(context.Background())
	
	// Measure execution time
	start := time.Now()
	report := flow.Run(ctx)
	elapsed := time.Since(start)
	
	// Verify success
	if report.Status != StatusCompleted {
		t.Errorf("Expected StatusCompleted, got %v", report.Status)
	}
	
	if len(report.Errors) != 0 {
		t.Errorf("Expected no errors, got %v", report.Errors)
	}
	
	// Verify all actions ran and stored results
	if result, ok := ctx.Get("action1_result"); !ok || result != "result1" {
		t.Errorf("Action1 result not found or incorrect")
	}
	if result, ok := ctx.Get("action2_result"); !ok || result != "result2" {
		t.Errorf("Action2 result not found or incorrect")
	}
	if result, ok := ctx.Get("action3_result"); !ok || result != "result3" {
		t.Errorf("Action3 result not found or incorrect")
	}
	
	// Verify parallel execution (should be faster than sequential)
	// Total sequential time would be ~45ms, parallel should be ~20ms (longest action)
	if elapsed > 35*time.Millisecond {
		t.Errorf("Execution took too long (%v), may not be parallel", elapsed)
	}
}

func TestParallelFlow_OneFailure(t *testing.T) {
	// Create mock actions - second one fails
	action1 := &MockAction{name: "action1", result: "result1", delay: 10 * time.Millisecond}
	action2 := &MockAction{name: "action2", shouldErr: true, delay: 5 * time.Millisecond}
	action3 := &MockAction{name: "action3", result: "result3", delay: 15 * time.Millisecond}
	
	// Create parallel flow
	flow := NewParallelFlow("test-parallel-flow").
		Execute(action1).
		Execute(action2).
		Execute(action3)
	
	// Create context
	ctx := NewWorkContext(context.Background())
	
	// Run flow
	report := flow.Run(ctx)
	
	// Verify partial failure (parallel flows typically continue even if one fails)
	if report.Status != StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", report.Status)
	}
	
	if len(report.Errors) == 0 {
		t.Errorf("Expected errors, got none")
	}
	
	// Verify successful actions still completed
	if result, ok := ctx.Get("action1_result"); !ok || result != "result1" {
		t.Errorf("Action1 result not found or incorrect")
	}
	if result, ok := ctx.Get("action3_result"); !ok || result != "result3" {
		t.Errorf("Action3 result not found or incorrect")
	}
	// Action2 shouldn't store result on failure
	if _, ok := ctx.Get("action2_result"); ok {
		t.Errorf("Action2 should not have stored result on failure")
	}
}

func TestParallelFlow_EmptyFlow(t *testing.T) {
	// Create empty flow
	flow := NewParallelFlow("empty-parallel-flow")
	
	// Create context
	ctx := NewWorkContext(context.Background())
	
	// Run flow
	report := flow.Run(ctx)
	
	// Verify skipped (empty flow should be marked as skipped since no actions completed)
	if report.Status != StatusSkipped {
		t.Errorf("Expected StatusSkipped for empty flow, got %v", report.Status)
	}
	
	if len(report.Errors) != 0 {
		t.Errorf("Expected no errors for empty flow, got %v", report.Errors)
	}
}

// Concurrent access test without external dependencies
func TestParallelFlow_SimpleConcurrent(t *testing.T) {
	// Create 3 actions for concurrent access test
	action1 := &MockAction{name: "concurrent1", result: "result1", delay: 5 * time.Millisecond}
	action2 := &MockAction{name: "concurrent2", result: "result2", delay: 5 * time.Millisecond}
	action3 := &MockAction{name: "concurrent3", result: "result3", delay: 5 * time.Millisecond}
	
	// Create parallel flow
	flow := NewParallelFlow("concurrent-test").
		Execute(action1).
		Execute(action2).
		Execute(action3)
	
	// Run flow multiple times to test for race conditions
	for i := 0; i < 5; i++ {
		// Create fresh context for each run
		testCtx := NewWorkContext(context.Background())
		
		report := flow.Run(testCtx)
		
		if report.Status != StatusCompleted {
			t.Errorf("Run %d: Expected StatusCompleted, got %v", i, report.Status)
		}
		
		// Verify all results
		if result, ok := testCtx.Get("concurrent1_result"); !ok || result != "result1" {
			t.Errorf("Run %d: concurrent1 result incorrect", i)
		}
		if result, ok := testCtx.Get("concurrent2_result"); !ok || result != "result2" {
			t.Errorf("Run %d: concurrent2 result incorrect", i)
		}
		if result, ok := testCtx.Get("concurrent3_result"); !ok || result != "result3" {
			t.Errorf("Run %d: concurrent3 result incorrect", i)
		}
	}
}