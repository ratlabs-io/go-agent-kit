package workflow

import (
	"context"
	"testing"

	"github.com/ratlabs-io/go-agent-kit/pkg/constants"
)

// mockLoopAction is a test action that records execution
type mockLoopAction struct {
	name        string
	executions  int
	shouldFail  bool
	failAtIndex int
}

func (m *mockLoopAction) Name() string {
	return m.name
}

func (m *mockLoopAction) Run(wctx WorkContext) WorkReport {
	m.executions++

	if m.shouldFail && m.executions == m.failAtIndex {
		return NewFailedWorkReport(nil)
	}

	return NewCompletedWorkReport()
}

func TestNewLoop(t *testing.T) {
	loop := NewLoop("test-loop", 5)

	if loop.name != "test-loop" {
		t.Errorf("Expected name 'test-loop', got '%s'", loop.name)
	}

	if loop.count != 5 {
		t.Errorf("Expected count 5, got %d", loop.count)
	}

	if !loop.isCountLoop {
		t.Error("Expected isCountLoop to be true")
	}
}

func TestLoop_CountLoop(t *testing.T) {
	wctx := NewWorkContext(context.Background())
	action := &mockLoopAction{name: "test-action"}

	loop := NewLoop("count-loop", 3).WithAction(action)
	report := loop.Run(wctx)

	if report.Status != StatusCompleted {
		t.Errorf("Expected StatusCompleted, got %v", report.Status)
	}

	if action.executions != 3 {
		t.Errorf("Expected 3 executions, got %d", action.executions)
	}

	// Check that loop iteration was set correctly
	iteration, ok := wctx.Get(constants.KeyLoopIteration)
	if !ok {
		t.Error("Expected loop iteration to be set in context")
	}

	if iteration != 3 {
		t.Errorf("Expected final iteration to be 3, got %v", iteration)
	}
}

func TestLoop_CountLoop_WithFailure(t *testing.T) {
	wctx := NewWorkContext(context.Background())
	action := &mockLoopAction{
		name:        "failing-action",
		shouldFail:  true,
		failAtIndex: 2,
	}

	loop := NewLoop("failing-loop", 5).WithAction(action)
	report := loop.Run(wctx)

	if report.Status != StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", report.Status)
	}

	if action.executions != 2 {
		t.Errorf("Expected 2 executions before failure, got %d", action.executions)
	}
}

func TestNewLoopWhile(t *testing.T) {
	condition := func(wctx WorkContext) (bool, error) {
		return true, nil
	}

	loop := NewLoopWhile("while-loop", condition)

	if loop.name != "while-loop" {
		t.Errorf("Expected name 'while-loop', got '%s'", loop.name)
	}

	if !loop.isWhileLoop {
		t.Error("Expected isWhileLoop to be true")
	}
}

func TestLoop_WhileLoop(t *testing.T) {
	wctx := NewWorkContext(context.Background())
	action := &mockLoopAction{name: "while-action"}

	// Create a condition that stops after 2 executions
	condition := func(wctx WorkContext) (bool, error) {
		// Check executions directly since the condition is checked before setting the iteration
		return action.executions < 2, nil
	}

	loop := NewLoopWhile("while-loop", condition).WithAction(action)
	report := loop.Run(wctx)

	if report.Status != StatusCompleted {
		t.Errorf("Expected StatusCompleted, got %v", report.Status)
	}

	if action.executions != 2 {
		t.Errorf("Expected 2 executions, got %d", action.executions)
	}
}

func TestNewLoopUntil(t *testing.T) {
	condition := func(wctx WorkContext) (bool, error) {
		return false, nil
	}

	loop := NewLoopUntil("until-loop", condition)

	if loop.name != "until-loop" {
		t.Errorf("Expected name 'until-loop', got '%s'", loop.name)
	}

	if !loop.isUntilLoop {
		t.Error("Expected isUntilLoop to be true")
	}
}

func TestLoop_UntilLoop(t *testing.T) {
	wctx := NewWorkContext(context.Background())
	action := &mockLoopAction{name: "until-action"}

	// Create a condition that returns true after 3 iterations
	condition := func(wctx WorkContext) (bool, error) {
		iter, _ := wctx.Get(constants.KeyLoopIteration)
		if iter == nil {
			return false, nil
		}
		return iter.(int) >= 3, nil
	}

	loop := NewLoopUntil("until-loop", condition).WithAction(action)
	report := loop.Run(wctx)

	if report.Status != StatusCompleted {
		t.Errorf("Expected StatusCompleted, got %v", report.Status)
	}

	if action.executions != 3 {
		t.Errorf("Expected 3 executions, got %d", action.executions)
	}
}

func TestNewLoopOver_Slice(t *testing.T) {
	items := []string{"a", "b", "c"}
	loop := NewLoopOver("iter-loop", items)

	if loop.name != "iter-loop" {
		t.Errorf("Expected name 'iter-loop', got '%s'", loop.name)
	}

	if !loop.isIterLoop {
		t.Error("Expected isIterLoop to be true")
	}
}

func TestLoop_IterLoop_Slice(t *testing.T) {
	wctx := NewWorkContext(context.Background())
	action := &mockLoopAction{name: "iter-action"}

	items := []string{"apple", "banana", "cherry"}
	loop := NewLoopOver("slice-loop", items).WithAction(action)
	report := loop.Run(wctx)

	if report.Status != StatusCompleted {
		t.Errorf("Expected StatusCompleted, got %v", report.Status)
	}

	if action.executions != 3 {
		t.Errorf("Expected 3 executions, got %d", action.executions)
	}

	// Check final context values
	item, ok := wctx.Get(constants.KeyCurrentItem)
	if !ok {
		t.Error("Expected current item to be set in context")
	}

	if item != "cherry" {
		t.Errorf("Expected final item to be 'cherry', got %v", item)
	}

	index, ok := wctx.Get(constants.KeyCurrentIndex)
	if !ok {
		t.Error("Expected current index to be set in context")
	}

	if index != 2 {
		t.Errorf("Expected final index to be 2, got %v", index)
	}
}

func TestLoop_IterLoop_Map(t *testing.T) {
	wctx := NewWorkContext(context.Background())
	action := &mockLoopAction{name: "map-action"}

	items := map[string]int{"one": 1, "two": 2, "three": 3}
	loop := NewLoopOver("map-loop", items).WithAction(action)
	report := loop.Run(wctx)

	if report.Status != StatusCompleted {
		t.Errorf("Expected StatusCompleted, got %v", report.Status)
	}

	if action.executions != 3 {
		t.Errorf("Expected 3 executions, got %d", action.executions)
	}
}

func TestLoop_NoAction(t *testing.T) {
	wctx := NewWorkContext(context.Background())

	loop := NewLoop("no-action-loop", 3)
	report := loop.Run(wctx)

	if report.Status != StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", report.Status)
	}

	if len(report.Errors) == 0 {
		t.Error("Expected error in report")
	}
}

func TestLoop_InfiniteLoopProtection(t *testing.T) {
	wctx := NewWorkContext(context.Background())
	action := &mockLoopAction{name: "infinite-action"}

	// Create a condition that always returns true
	condition := func(wctx WorkContext) (bool, error) {
		return true, nil
	}

	loop := NewLoopWhile("infinite-loop", condition).WithAction(action)
	report := loop.Run(wctx)

	if report.Status != StatusFailure {
		t.Errorf("Expected StatusFailure for infinite loop, got %v", report.Status)
	}

	if len(report.Errors) == 0 {
		t.Error("Expected error in report for infinite loop")
	}
}
