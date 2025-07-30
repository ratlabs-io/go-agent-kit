package workflow

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockAdvancedAction is a test action for advanced constructs testing
type mockAdvancedAction struct {
	name          string
	executions    int
	shouldFail    bool
	delay         time.Duration
	errorToReturn error
}

func (m *mockAdvancedAction) Name() string {
	return m.name
}

func (m *mockAdvancedAction) Run(wctx WorkContext) WorkReport {
	m.executions++

	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	if m.shouldFail {
		if m.errorToReturn != nil {
			return NewFailedWorkReport(m.errorToReturn)
		}
		return NewFailedWorkReport(errors.New("mock error"))
	}

	return NewCompletedWorkReport()
}

// Test Circuit Breaker

func TestNewCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker("test-cb", 3, 5*time.Second, 10*time.Second)

	if cb.name != "test-cb" {
		t.Errorf("Expected name 'test-cb', got '%s'", cb.name)
	}

	if cb.failureThreshold != 3 {
		t.Errorf("Expected failure threshold 3, got %d", cb.failureThreshold)
	}

	if cb.getState() != CircuitBreakerClosed {
		t.Errorf("Expected initial state to be Closed, got %v", cb.getState())
	}
}

func TestCircuitBreaker_Success(t *testing.T) {
	wctx := NewWorkContext(context.Background())
	action := &mockAdvancedAction{
		name:       "success-action",
		shouldFail: false,
	}

	cb := NewCircuitBreaker("success-cb", 3, 5*time.Second, 10*time.Second).
		WithAction(action)

	report := cb.Run(wctx)

	if report.Status != StatusCompleted {
		t.Errorf("Expected StatusCompleted, got %v", report.Status)
	}

	if action.executions != 1 {
		t.Errorf("Expected 1 execution, got %d", action.executions)
	}

	if cb.getState() != CircuitBreakerClosed {
		t.Errorf("Expected state to remain Closed, got %v", cb.getState())
	}
}

func TestCircuitBreaker_FailureThreshold(t *testing.T) {
	wctx := NewWorkContext(context.Background())
	action := &mockAdvancedAction{
		name:       "failing-action",
		shouldFail: true,
	}

	cb := NewCircuitBreaker("threshold-cb", 2, 5*time.Second, 10*time.Second).
		WithAction(action)

	// First failure
	report1 := cb.Run(wctx)
	if report1.Status != StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", report1.Status)
	}
	if cb.getState() != CircuitBreakerClosed {
		t.Errorf("Expected state to remain Closed after first failure, got %v", cb.getState())
	}

	// Second failure - should open circuit
	report2 := cb.Run(wctx)
	if report2.Status != StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", report2.Status)
	}
	if cb.getState() != CircuitBreakerOpen {
		t.Errorf("Expected state to be Open after reaching threshold, got %v", cb.getState())
	}

	// Third attempt should be rejected
	report3 := cb.Run(wctx)
	if report3.Status != StatusFailure {
		t.Errorf("Expected StatusFailure (rejected), got %v", report3.Status)
	}
	if action.executions != 2 {
		t.Errorf("Expected 2 executions (third should be rejected), got %d", action.executions)
	}
}

func TestCircuitBreaker_Recovery(t *testing.T) {
	wctx := NewWorkContext(context.Background())
	failingAction := &mockAdvancedAction{
		name:       "initially-failing-action",
		shouldFail: true,
	}

	cb := NewCircuitBreaker("recovery-cb", 1, 50*time.Millisecond, 100*time.Millisecond).
		WithAction(failingAction)

	// Trigger failure to open circuit
	cb.Run(wctx)
	if cb.getState() != CircuitBreakerOpen {
		t.Errorf("Expected state to be Open, got %v", cb.getState())
	}

	// Wait for recovery timeout
	time.Sleep(60 * time.Millisecond)

	// Now make action succeed
	failingAction.shouldFail = false

	// Next request should be allowed (half-open) and succeed
	report := cb.Run(wctx)
	if report.Status != StatusCompleted {
		t.Errorf("Expected StatusCompleted during recovery, got %v", report.Status)
	}

	if cb.getState() != CircuitBreakerClosed {
		t.Errorf("Expected state to be Closed after successful recovery, got %v", cb.getState())
	}
}

func TestCircuitBreaker_NoAction(t *testing.T) {
	wctx := NewWorkContext(context.Background())

	cb := NewCircuitBreaker("no-action-cb", 3, 5*time.Second, 10*time.Second)
	report := cb.Run(wctx)

	if report.Status != StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", report.Status)
	}
}

// Test Timeout Wrapper

func TestNewTimeoutWrapper(t *testing.T) {
	tw := NewTimeoutWrapper("test-timeout", 5*time.Second)

	if tw.name != "test-timeout" {
		t.Errorf("Expected name 'test-timeout', got '%s'", tw.name)
	}

	if tw.timeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", tw.timeout)
	}
}

func TestTimeoutWrapper_Success(t *testing.T) {
	wctx := NewWorkContext(context.Background())
	action := &mockAdvancedAction{
		name:  "fast-action",
		delay: 10 * time.Millisecond,
	}

	tw := NewTimeoutWrapper("success-timeout", 100*time.Millisecond).
		WithAction(action)

	start := time.Now()
	report := tw.Run(wctx)
	duration := time.Since(start)

	if report.Status != StatusCompleted {
		t.Errorf("Expected StatusCompleted, got %v", report.Status)
	}

	if action.executions != 1 {
		t.Errorf("Expected 1 execution, got %d", action.executions)
	}

	if duration >= 100*time.Millisecond {
		t.Errorf("Expected duration < 100ms, got %v", duration)
	}
}

func TestTimeoutWrapper_Timeout(t *testing.T) {
	wctx := NewWorkContext(context.Background())
	action := &mockAdvancedAction{
		name:  "slow-action",
		delay: 200 * time.Millisecond,
	}

	tw := NewTimeoutWrapper("timeout-test", 50*time.Millisecond).
		WithAction(action)

	start := time.Now()
	report := tw.Run(wctx)
	duration := time.Since(start)

	if report.Status != StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", report.Status)
	}

	if len(report.Errors) == 0 {
		t.Error("Expected timeout error in report")
	}

	// Should complete around the timeout duration
	if duration < 45*time.Millisecond || duration > 100*time.Millisecond {
		t.Errorf("Expected duration around 50ms, got %v", duration)
	}
}

func TestTimeoutWrapper_NoAction(t *testing.T) {
	wctx := NewWorkContext(context.Background())

	tw := NewTimeoutWrapper("no-action-timeout", 100*time.Millisecond)
	report := tw.Run(wctx)

	if report.Status != StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", report.Status)
	}
}

// Test Parallel Error Collector

func TestNewParallelErrorCollector(t *testing.T) {
	pec := NewParallelErrorCollector("test-parallel")

	if pec.name != "test-parallel" {
		t.Errorf("Expected name 'test-parallel', got '%s'", pec.name)
	}

	if len(pec.actions) != 0 {
		t.Errorf("Expected 0 actions, got %d", len(pec.actions))
	}
}

func TestParallelErrorCollector_AllSuccess(t *testing.T) {
	wctx := NewWorkContext(context.Background())

	action1 := &mockAdvancedAction{name: "action1"}
	action2 := &mockAdvancedAction{name: "action2"}
	action3 := &mockAdvancedAction{name: "action3"}

	pec := NewParallelErrorCollector("all-success").
		AddAction(action1).
		AddAction(action2).
		AddAction(action3)

	report := pec.Run(wctx)

	if report.Status != StatusCompleted {
		t.Errorf("Expected StatusCompleted, got %v", report.Status)
	}

	if len(report.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(report.Errors))
	}

	if action1.executions != 1 || action2.executions != 1 || action3.executions != 1 {
		t.Errorf("Expected all actions to execute once")
	}

	// Check metadata
	if report.Metadata["total_actions"] != 3 {
		t.Errorf("Expected total_actions to be 3, got %v", report.Metadata["total_actions"])
	}

	if report.Metadata["successful_actions"] != 3 {
		t.Errorf("Expected successful_actions to be 3, got %v", report.Metadata["successful_actions"])
	}

	if report.Metadata["failed_actions"] != 0 {
		t.Errorf("Expected failed_actions to be 0, got %v", report.Metadata["failed_actions"])
	}
}

func TestParallelErrorCollector_SomeFailures(t *testing.T) {
	wctx := NewWorkContext(context.Background())

	action1 := &mockAdvancedAction{name: "success1"}
	action2 := &mockAdvancedAction{name: "failure", shouldFail: true}
	action3 := &mockAdvancedAction{name: "success2"}

	pec := NewParallelErrorCollector("some-failures").
		AddActions(action1, action2, action3)

	report := pec.Run(wctx)

	if report.Status != StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", report.Status)
	}

	if len(report.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(report.Errors))
	}

	if action1.executions != 1 || action2.executions != 1 || action3.executions != 1 {
		t.Errorf("Expected all actions to execute once")
	}

	// Check metadata
	if report.Metadata["total_actions"] != 3 {
		t.Errorf("Expected total_actions to be 3, got %v", report.Metadata["total_actions"])
	}

	if report.Metadata["successful_actions"] != 2 {
		t.Errorf("Expected successful_actions to be 2, got %v", report.Metadata["successful_actions"])
	}

	if report.Metadata["failed_actions"] != 1 {
		t.Errorf("Expected failed_actions to be 1, got %v", report.Metadata["failed_actions"])
	}

	if report.Metadata["total_errors"] != 1 {
		t.Errorf("Expected total_errors to be 1, got %v", report.Metadata["total_errors"])
	}
}

func TestParallelErrorCollector_AllFailures(t *testing.T) {
	wctx := NewWorkContext(context.Background())

	action1 := &mockAdvancedAction{name: "failure1", shouldFail: true}
	action2 := &mockAdvancedAction{name: "failure2", shouldFail: true}

	pec := NewParallelErrorCollector("all-failures").
		AddAction(action1).
		AddAction(action2)

	report := pec.Run(wctx)

	if report.Status != StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", report.Status)
	}

	if len(report.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(report.Errors))
	}

	if report.Metadata["successful_actions"] != 0 {
		t.Errorf("Expected successful_actions to be 0, got %v", report.Metadata["successful_actions"])
	}

	if report.Metadata["failed_actions"] != 2 {
		t.Errorf("Expected failed_actions to be 2, got %v", report.Metadata["failed_actions"])
	}
}

func TestParallelErrorCollector_NoActions(t *testing.T) {
	wctx := NewWorkContext(context.Background())

	pec := NewParallelErrorCollector("no-actions")
	report := pec.Run(wctx)

	if report.Status != StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", report.Status)
	}

	if len(report.Errors) == 0 {
		t.Error("Expected error for no actions")
	}
}

func TestParallelErrorCollector_Concurrency(t *testing.T) {
	wctx := NewWorkContext(context.Background())

	// Create actions with different delays to test concurrency
	action1 := &mockAdvancedAction{name: "slow", delay: 50 * time.Millisecond}
	action2 := &mockAdvancedAction{name: "medium", delay: 30 * time.Millisecond}
	action3 := &mockAdvancedAction{name: "fast", delay: 10 * time.Millisecond}

	pec := NewParallelErrorCollector("concurrency-test").
		AddActions(action1, action2, action3)

	start := time.Now()
	report := pec.Run(wctx)
	duration := time.Since(start)

	if report.Status != StatusCompleted {
		t.Errorf("Expected StatusCompleted, got %v", report.Status)
	}

	// Should complete in roughly the time of the slowest action (50ms)
	// Allow some tolerance for goroutine scheduling
	if duration > 80*time.Millisecond {
		t.Errorf("Expected duration around 50ms (parallel execution), got %v", duration)
	}

	if duration < 45*time.Millisecond {
		t.Errorf("Expected duration at least 45ms (slowest action), got %v", duration)
	}
}
