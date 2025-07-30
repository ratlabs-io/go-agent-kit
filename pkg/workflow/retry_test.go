package workflow

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockRetryAction is a test action for retry testing
type mockRetryAction struct {
	name             string
	executions       int
	failUntilAttempt int
	errorToReturn    error
}

func (m *mockRetryAction) Name() string {
	return m.name
}

func (m *mockRetryAction) Run(wctx WorkContext) WorkReport {
	m.executions++

	if m.executions <= m.failUntilAttempt {
		if m.errorToReturn != nil {
			return NewFailedWorkReport(m.errorToReturn)
		}
		return NewFailedWorkReport(errors.New("mock error"))
	}

	return NewCompletedWorkReport()
}

func TestNewRetry(t *testing.T) {
	retry := NewRetry("test-retry", 3)

	if retry.name != "test-retry" {
		t.Errorf("Expected name 'test-retry', got '%s'", retry.name)
	}

	if retry.maxAttempts != 3 {
		t.Errorf("Expected maxAttempts 3, got %d", retry.maxAttempts)
	}

	if retry.backoffStrategy == nil {
		t.Error("Expected default backoff strategy to be set")
	}

	if retry.retryCondition == nil {
		t.Error("Expected default retry condition to be set")
	}
}

func TestRetry_SuccessOnFirstAttempt(t *testing.T) {
	wctx := NewWorkContext(context.Background())
	action := &mockRetryAction{
		name:             "success-action",
		failUntilAttempt: 0, // Never fail
	}

	retry := NewRetry("success-retry", 3).WithAction(action)
	report := retry.Run(wctx)

	if report.Status != StatusCompleted {
		t.Errorf("Expected StatusCompleted, got %v", report.Status)
	}

	if action.executions != 1 {
		t.Errorf("Expected 1 execution, got %d", action.executions)
	}
}

func TestRetry_SuccessOnSecondAttempt(t *testing.T) {
	wctx := NewWorkContext(context.Background())
	action := &mockRetryAction{
		name:             "retry-once-action",
		failUntilAttempt: 1, // Fail first attempt, succeed on second
	}

	retry := NewRetry("retry-once", 3).
		WithAction(action).
		WithBackoffStrategy(NewFixedBackoff(10 * time.Millisecond))

	start := time.Now()
	report := retry.Run(wctx)
	duration := time.Since(start)

	if report.Status != StatusCompleted {
		t.Errorf("Expected StatusCompleted, got %v", report.Status)
	}

	if action.executions != 2 {
		t.Errorf("Expected 2 executions, got %d", action.executions)
	}

	// Should have waited at least once
	if duration < 10*time.Millisecond {
		t.Errorf("Expected duration >= 10ms, got %v", duration)
	}
}

func TestRetry_AllAttemptsFail(t *testing.T) {
	wctx := NewWorkContext(context.Background())
	action := &mockRetryAction{
		name:             "always-fail-action",
		failUntilAttempt: 10, // Always fail
	}

	retry := NewRetry("always-fail", 3).
		WithAction(action).
		WithBackoffStrategy(NewFixedBackoff(1 * time.Millisecond))

	report := retry.Run(wctx)

	if report.Status != StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", report.Status)
	}

	if action.executions != 3 {
		t.Errorf("Expected 3 executions, got %d", action.executions)
	}

	// Check metadata
	if report.Metadata["retry_attempts"] != 3 {
		t.Errorf("Expected retry_attempts metadata to be 3, got %v", report.Metadata["retry_attempts"])
	}

	if report.Metadata["retry_name"] != "always-fail" {
		t.Errorf("Expected retry_name metadata to be 'always-fail', got %v", report.Metadata["retry_name"])
	}
}

func TestRetry_WithCustomRetryCondition(t *testing.T) {
	wctx := NewWorkContext(context.Background())
	action := &mockRetryAction{
		name:             "timeout-action",
		failUntilAttempt: 10, // Always fail
		errorToReturn:    errors.New("connection timeout"),
	}

	retry := NewRetry("timeout-retry", 3).
		WithAction(action).
		WithRetryCondition(RetryOnTimeoutCondition).
		WithBackoffStrategy(NewFixedBackoff(1 * time.Millisecond))

	report := retry.Run(wctx)

	if report.Status != StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", report.Status)
	}

	if action.executions != 3 {
		t.Errorf("Expected 3 executions, got %d", action.executions)
	}
}

func TestRetry_WithCustomRetryCondition_NoRetry(t *testing.T) {
	wctx := NewWorkContext(context.Background())
	action := &mockRetryAction{
		name:             "non-timeout-action",
		failUntilAttempt: 10, // Always fail
		errorToReturn:    errors.New("validation error"),
	}

	retry := NewRetry("validation-retry", 3).
		WithAction(action).
		WithRetryCondition(RetryOnTimeoutCondition). // Only retry on timeout
		WithBackoffStrategy(NewFixedBackoff(1 * time.Millisecond))

	report := retry.Run(wctx)

	if report.Status != StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", report.Status)
	}

	// Should only execute once since retry condition is not met
	if action.executions != 1 {
		t.Errorf("Expected 1 execution, got %d", action.executions)
	}
}

func TestRetry_WithStopCondition(t *testing.T) {
	wctx := NewWorkContext(context.Background())
	action := &mockRetryAction{
		name:             "stop-condition-action",
		failUntilAttempt: 10, // Always fail
	}

	// Stop condition that triggers after first execution
	stopCondition := func(wctx WorkContext) bool {
		// Check if action has been executed
		return action.executions >= 1
	}

	retry := NewRetry("stop-condition-retry", 5).
		WithAction(action).
		WithStopCondition(stopCondition).
		WithBackoffStrategy(NewFixedBackoff(1 * time.Millisecond))

	report := retry.Run(wctx)

	if report.Status != StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", report.Status)
	}

	// Should only execute once due to stop condition
	if action.executions != 1 {
		t.Errorf("Expected 1 execution, got %d", action.executions)
	}
}

func TestRetry_NoAction(t *testing.T) {
	wctx := NewWorkContext(context.Background())

	retry := NewRetry("no-action-retry", 3)
	report := retry.Run(wctx)

	if report.Status != StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", report.Status)
	}

	if len(report.Errors) == 0 {
		t.Error("Expected error in report")
	}
}

// Test backoff strategies

func TestFixedBackoff(t *testing.T) {
	backoff := NewFixedBackoff(100 * time.Millisecond)

	delay1 := backoff.CalculateDelay(0)
	delay2 := backoff.CalculateDelay(5)
	delay3 := backoff.CalculateDelay(10)

	if delay1 != 100*time.Millisecond {
		t.Errorf("Expected 100ms, got %v", delay1)
	}

	if delay2 != 100*time.Millisecond {
		t.Errorf("Expected 100ms, got %v", delay2)
	}

	if delay3 != 100*time.Millisecond {
		t.Errorf("Expected 100ms, got %v", delay3)
	}
}

func TestLinearBackoff(t *testing.T) {
	backoff := NewLinearBackoff(100*time.Millisecond, 50*time.Millisecond)

	delay0 := backoff.CalculateDelay(0)
	delay1 := backoff.CalculateDelay(1)
	delay2 := backoff.CalculateDelay(2)

	if delay0 != 100*time.Millisecond {
		t.Errorf("Expected 100ms, got %v", delay0)
	}

	if delay1 != 150*time.Millisecond {
		t.Errorf("Expected 150ms, got %v", delay1)
	}

	if delay2 != 200*time.Millisecond {
		t.Errorf("Expected 200ms, got %v", delay2)
	}
}

func TestExponentialBackoff(t *testing.T) {
	backoff := NewExponentialBackoff(100*time.Millisecond, 1*time.Second, 2.0)

	delay0 := backoff.CalculateDelay(0)
	delay1 := backoff.CalculateDelay(1)
	delay2 := backoff.CalculateDelay(2)
	delay10 := backoff.CalculateDelay(10) // Should be capped at maxDelay

	if delay0 != 0 {
		t.Errorf("Expected 0ms, got %v", delay0)
	}

	if delay1 != 200*time.Millisecond {
		t.Errorf("Expected 200ms, got %v", delay1)
	}

	if delay2 != 400*time.Millisecond {
		t.Errorf("Expected 400ms, got %v", delay2)
	}

	if delay10 != 1*time.Second {
		t.Errorf("Expected 1s (capped), got %v", delay10)
	}
}

// Test retry conditions

func TestRetryConditions(t *testing.T) {
	tests := []struct {
		name      string
		condition RetryConditionFunc
		err       error
		expected  bool
	}{
		{
			name:      "DefaultRetryCondition with error",
			condition: DefaultRetryCondition,
			err:       errors.New("some error"),
			expected:  true,
		},
		{
			name:      "DefaultRetryCondition with nil error",
			condition: DefaultRetryCondition,
			err:       nil,
			expected:  false,
		},
		{
			name:      "NeverRetryCondition with error",
			condition: NeverRetryCondition,
			err:       errors.New("some error"),
			expected:  false,
		},
		{
			name:      "RetryOnTimeoutCondition with timeout error",
			condition: RetryOnTimeoutCondition,
			err:       errors.New("connection timeout"),
			expected:  true,
		},
		{
			name:      "RetryOnTimeoutCondition with other error",
			condition: RetryOnTimeoutCondition,
			err:       errors.New("validation error"),
			expected:  false,
		},
		{
			name:      "RetryOnRateLimitCondition with rate limit error",
			condition: RetryOnRateLimitCondition,
			err:       errors.New("rate limit exceeded"),
			expected:  true,
		},
		{
			name:      "RetryOnNetworkCondition with network error",
			condition: RetryOnNetworkCondition,
			err:       errors.New("connection refused"),
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.condition(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCombineRetryConditions(t *testing.T) {
	combined := CombineRetryConditions(
		RetryOnTimeoutCondition,
		RetryOnRateLimitCondition,
	)

	tests := []struct {
		err      error
		expected bool
	}{
		{errors.New("connection timeout"), true},
		{errors.New("rate limit exceeded"), true},
		{errors.New("validation error"), false},
		{nil, false},
	}

	for _, tt := range tests {
		result := combined(tt.err)
		if result != tt.expected {
			t.Errorf("For error %v, expected %v, got %v", tt.err, tt.expected, result)
		}
	}
}
