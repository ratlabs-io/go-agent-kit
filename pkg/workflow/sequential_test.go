package workflow

import (
	"context"
	"testing"
	"time"
)

// MockAction implements the Action interface for testing
type MockAction struct {
	name      string
	shouldErr bool
	delay     time.Duration
	result    interface{}
}

func (ma *MockAction) Name() string {
	return ma.name
}

func (ma *MockAction) Run(wctx *WorkContext) WorkReport {
	if ma.delay > 0 {
		time.Sleep(ma.delay)
	}
	
	if ma.shouldErr {
		return NewFailedWorkReport(&MockError{Message: "mock error"})
	}
	
	report := NewCompletedWorkReport()
	report.Data = ma.result
	
	// Store result in context for chaining
	wctx.Set(ma.name+"_result", ma.result)
	
	return report
}

func TestSequentialFlow_Success(t *testing.T) {
	// Create mock actions
	action1 := &MockAction{name: "action1", result: "result1"}
	action2 := &MockAction{name: "action2", result: "result2"}
	action3 := &MockAction{name: "action3", result: "result3"}
	
	// Create sequential flow
	flow := NewSequentialFlow("test-flow").
		Then(action1).
		Then(action2).
		Then(action3)
	
	// Create context
	ctx := NewWorkContext(context.Background())
	
	// Run flow
	report := flow.Run(ctx)
	
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
}

func TestSequentialFlow_EarlyFailure(t *testing.T) {
	// Create mock actions - second one fails
	action1 := &MockAction{name: "action1", result: "result1"}
	action2 := &MockAction{name: "action2", shouldErr: true}
	action3 := &MockAction{name: "action3", result: "result3"}
	
	// Create sequential flow
	flow := NewSequentialFlow("test-flow").
		Then(action1).
		Then(action2).
		Then(action3)
	
	// Create context
	ctx := NewWorkContext(context.Background())
	
	// Run flow
	report := flow.Run(ctx)
	
	// Verify failure
	if report.Status != StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", report.Status)
	}
	
	if len(report.Errors) == 0 {
		t.Errorf("Expected errors, got none")
	}
	
	// Verify only first action ran
	if result, ok := ctx.Get("action1_result"); !ok || result != "result1" {
		t.Errorf("Action1 result not found or incorrect")
	}
	if _, ok := ctx.Get("action2_result"); ok {
		t.Errorf("Action2 should not have stored result on failure")
	}
	if _, ok := ctx.Get("action3_result"); ok {
		t.Errorf("Action3 should not have run after action2 failed")
	}
}

func TestSequentialFlow_EmptyFlow(t *testing.T) {
	// Create empty flow
	flow := NewSequentialFlow("empty-flow")
	
	// Create context
	ctx := NewWorkContext(context.Background())
	
	// Run flow
	report := flow.Run(ctx)
	
	// Verify success (empty flow should complete successfully)
	if report.Status != StatusCompleted {
		t.Errorf("Expected StatusCompleted for empty flow, got %v", report.Status)
	}
	
	if len(report.Errors) != 0 {
		t.Errorf("Expected no errors for empty flow, got %v", report.Errors)
	}
}

func TestSequentialFlow_Name(t *testing.T) {
	flow := NewSequentialFlow("test-name")
	
	if flow.Name() != "test-name" {
		t.Errorf("Expected name 'test-name', got '%s'", flow.Name())
	}
}

// MockError implements error interface for testing
type MockError struct {
	Message string
}

func (me *MockError) Error() string {
	return me.Message
}