package workflow

import (
	"context"
	"errors"
	"testing"
)

// mockTryCatchAction is a test action for try-catch testing
type mockTryCatchAction struct {
	name          string
	executions    int
	shouldFail    bool
	errorToReturn error
}

func (m *mockTryCatchAction) Name() string {
	return m.name
}

func (m *mockTryCatchAction) Run(wctx WorkContext) WorkReport {
	m.executions++

	if m.shouldFail {
		if m.errorToReturn != nil {
			return NewFailedWorkReport(m.errorToReturn)
		}
		return NewFailedWorkReport(errors.New("mock error"))
	}

	return NewCompletedWorkReport()
}

// mockErrorHandlerAction is a test error handler
type mockErrorHandlerAction struct {
	name         string
	executions   int
	handledError error
}

func (m *mockErrorHandlerAction) Name() string {
	return m.name
}

func (m *mockErrorHandlerAction) Run(wctx WorkContext) WorkReport {
	m.executions++
	return NewCompletedWorkReport()
}

func (m *mockErrorHandlerAction) HandleError(wctx WorkContext, err error) WorkReport {
	m.executions++
	m.handledError = err
	return NewCompletedWorkReport()
}

func TestNewTryCatch(t *testing.T) {
	tc := NewTryCatch("test-try-catch")

	if tc.name != "test-try-catch" {
		t.Errorf("Expected name 'test-try-catch', got '%s'", tc.name)
	}

	if len(tc.catchHandlers) != 0 {
		t.Errorf("Expected 0 catch handlers, got %d", len(tc.catchHandlers))
	}
}

func TestTryCatch_Success(t *testing.T) {
	wctx := NewWorkContext(context.Background())

	tryAction := &mockTryCatchAction{
		name:       "success-action",
		shouldFail: false,
	}

	tc := NewTryCatch("success-try-catch").WithTryAction(tryAction)
	report := tc.Run(wctx)

	if report.Status != StatusCompleted {
		t.Errorf("Expected StatusCompleted, got %v", report.Status)
	}

	if tryAction.executions != 1 {
		t.Errorf("Expected 1 execution, got %d", tryAction.executions)
	}
}

func TestTryCatch_WithCatch(t *testing.T) {
	wctx := NewWorkContext(context.Background())

	tryAction := &mockTryCatchAction{
		name:          "failing-action",
		shouldFail:    true,
		errorToReturn: errors.New("timeout error"),
	}

	catchAction := &mockErrorHandlerAction{name: "timeout-handler"}

	tc := NewTryCatch("timeout-try-catch").
		WithTryAction(tryAction).
		Catch(TimeoutError, catchAction)

	report := tc.Run(wctx)

	if report.Status != StatusCompleted {
		t.Errorf("Expected StatusCompleted, got %v", report.Status)
	}

	if tryAction.executions != 1 {
		t.Errorf("Expected 1 try execution, got %d", tryAction.executions)
	}

	if catchAction.executions != 1 {
		t.Errorf("Expected 1 catch execution, got %d", catchAction.executions)
	}

	if catchAction.handledError == nil {
		t.Error("Expected catch handler to receive error")
	}

	if catchAction.handledError.Error() != "timeout error" {
		t.Errorf("Expected handled error to be 'timeout error', got '%s'", catchAction.handledError.Error())
	}
}

func TestTryCatch_WithCatchNoMatch(t *testing.T) {
	wctx := NewWorkContext(context.Background())

	tryAction := &mockTryCatchAction{
		name:          "failing-action",
		shouldFail:    true,
		errorToReturn: errors.New("validation error"),
	}

	catchAction := &mockErrorHandlerAction{name: "timeout-handler"}

	tc := NewTryCatch("nomatch-try-catch").
		WithTryAction(tryAction).
		Catch(TimeoutError, catchAction) // Only catches timeout errors

	report := tc.Run(wctx)

	if report.Status != StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", report.Status)
	}

	if tryAction.executions != 1 {
		t.Errorf("Expected 1 try execution, got %d", tryAction.executions)
	}

	if catchAction.executions != 0 {
		t.Errorf("Expected 0 catch executions, got %d", catchAction.executions)
	}
}

func TestTryCatch_WithCatchAny(t *testing.T) {
	wctx := NewWorkContext(context.Background())

	tryAction := &mockTryCatchAction{
		name:          "failing-action",
		shouldFail:    true,
		errorToReturn: errors.New("any error"),
	}

	catchAnyAction := &mockErrorHandlerAction{name: "catch-any-handler"}

	tc := NewTryCatch("catchany-try-catch").
		WithTryAction(tryAction).
		CatchAny(catchAnyAction)

	report := tc.Run(wctx)

	if report.Status != StatusCompleted {
		t.Errorf("Expected StatusCompleted, got %v", report.Status)
	}

	if tryAction.executions != 1 {
		t.Errorf("Expected 1 try execution, got %d", tryAction.executions)
	}

	if catchAnyAction.executions != 1 {
		t.Errorf("Expected 1 catch-any execution, got %d", catchAnyAction.executions)
	}
}

func TestTryCatch_WithFinally(t *testing.T) {
	wctx := NewWorkContext(context.Background())

	tryAction := &mockTryCatchAction{
		name:       "success-action",
		shouldFail: false,
	}

	finallyAction := &mockTryCatchAction{name: "finally-action"}

	tc := NewTryCatch("finally-try-catch").
		WithTryAction(tryAction).
		Finally(finallyAction)

	report := tc.Run(wctx)

	if report.Status != StatusCompleted {
		t.Errorf("Expected StatusCompleted, got %v", report.Status)
	}

	if tryAction.executions != 1 {
		t.Errorf("Expected 1 try execution, got %d", tryAction.executions)
	}

	if finallyAction.executions != 1 {
		t.Errorf("Expected 1 finally execution, got %d", finallyAction.executions)
	}
}

func TestTryCatch_FinallyWithFailingTry(t *testing.T) {
	wctx := NewWorkContext(context.Background())

	tryAction := &mockTryCatchAction{
		name:       "failing-action",
		shouldFail: true,
	}

	finallyAction := &mockTryCatchAction{name: "finally-action"}

	tc := NewTryCatch("finally-fail-try-catch").
		WithTryAction(tryAction).
		Finally(finallyAction)

	report := tc.Run(wctx)

	if report.Status != StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", report.Status)
	}

	if tryAction.executions != 1 {
		t.Errorf("Expected 1 try execution, got %d", tryAction.executions)
	}

	if finallyAction.executions != 1 {
		t.Errorf("Expected 1 finally execution, got %d", finallyAction.executions)
	}
}

func TestTryCatch_FinallyOverrideResult(t *testing.T) {
	wctx := NewWorkContext(context.Background())

	tryAction := &mockTryCatchAction{
		name:       "success-action",
		shouldFail: false,
	}

	finallyAction := &mockTryCatchAction{
		name:       "failing-finally-action",
		shouldFail: true,
	}

	tc := NewTryCatch("finally-override-try-catch").
		WithTryAction(tryAction).
		Finally(finallyAction)

	report := tc.Run(wctx)

	if report.Status != StatusFailure {
		t.Errorf("Expected StatusFailure (finally override), got %v", report.Status)
	}

	if tryAction.executions != 1 {
		t.Errorf("Expected 1 try execution, got %d", tryAction.executions)
	}

	if finallyAction.executions != 1 {
		t.Errorf("Expected 1 finally execution, got %d", finallyAction.executions)
	}
}

func TestTryCatch_ComplexFlow(t *testing.T) {
	wctx := NewWorkContext(context.Background())

	tryAction := &mockTryCatchAction{
		name:          "complex-failing-action",
		shouldFail:    true,
		errorToReturn: errors.New("connection refused"),
	}

	timeoutCatchAction := &mockErrorHandlerAction{name: "timeout-handler"}
	networkCatchAction := &mockErrorHandlerAction{name: "network-handler"}
	catchAnyAction := &mockErrorHandlerAction{name: "catch-any-handler"}
	finallyAction := &mockTryCatchAction{name: "finally-action"}

	tc := NewTryCatch("complex-try-catch").
		WithTryAction(tryAction).
		Catch(TimeoutError, timeoutCatchAction).
		Catch(NetworkError, networkCatchAction).
		CatchAny(catchAnyAction).
		Finally(finallyAction)

	report := tc.Run(wctx)

	if report.Status != StatusCompleted {
		t.Errorf("Expected StatusCompleted, got %v", report.Status)
	}

	if tryAction.executions != 1 {
		t.Errorf("Expected 1 try execution, got %d", tryAction.executions)
	}

	if timeoutCatchAction.executions != 0 {
		t.Errorf("Expected 0 timeout catch executions, got %d", timeoutCatchAction.executions)
	}

	if networkCatchAction.executions != 1 {
		t.Errorf("Expected 1 network catch execution, got %d", networkCatchAction.executions)
	}

	if catchAnyAction.executions != 0 {
		t.Errorf("Expected 0 catch-any executions, got %d", catchAnyAction.executions)
	}

	if finallyAction.executions != 1 {
		t.Errorf("Expected 1 finally execution, got %d", finallyAction.executions)
	}
}

func TestTryCatch_NoTryAction(t *testing.T) {
	wctx := NewWorkContext(context.Background())

	tc := NewTryCatch("no-try-action")
	report := tc.Run(wctx)

	if report.Status != StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", report.Status)
	}

	if len(report.Errors) == 0 {
		t.Error("Expected error in report")
	}
}

// Test error matchers

func TestErrorMatchers(t *testing.T) {
	tests := []struct {
		name     string
		matcher  ErrorTypeMatcherFunc
		err      error
		expected bool
	}{
		{
			name:     "TimeoutError with timeout",
			matcher:  TimeoutError,
			err:      errors.New("connection timeout"),
			expected: true,
		},
		{
			name:     "TimeoutError with other error",
			matcher:  TimeoutError,
			err:      errors.New("validation error"),
			expected: false,
		},
		{
			name:     "NetworkError with network error",
			matcher:  NetworkError,
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "ValidationError with validation error",
			matcher:  ValidationError,
			err:      errors.New("validation failed"),
			expected: true,
		},
		{
			name:     "AnyError with error",
			matcher:  AnyError,
			err:      errors.New("any error"),
			expected: true,
		},
		{
			name:     "AnyError with nil",
			matcher:  AnyError,
			err:      nil,
			expected: false,
		},
		{
			name:     "NoError with nil",
			matcher:  NoError,
			err:      nil,
			expected: true,
		},
		{
			name:     "NoError with error",
			matcher:  NoError,
			err:      errors.New("error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.matcher(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestErrorMessageContains(t *testing.T) {
	matcher := ErrorMessageContains("timeout")

	tests := []struct {
		err      error
		expected bool
	}{
		{errors.New("connection timeout"), true},
		{errors.New("timeout occurred"), true},
		{errors.New("validation error"), false},
		{nil, false},
	}

	for _, tt := range tests {
		result := matcher(tt.err)
		if result != tt.expected {
			t.Errorf("For error %v, expected %v, got %v", tt.err, tt.expected, result)
		}
	}
}

func TestErrorMessageEquals(t *testing.T) {
	matcher := ErrorMessageEquals("exact error")

	tests := []struct {
		err      error
		expected bool
	}{
		{errors.New("exact error"), true},
		{errors.New("different error"), false},
		{errors.New("exact error with more"), false},
		{nil, false},
	}

	for _, tt := range tests {
		result := matcher(tt.err)
		if result != tt.expected {
			t.Errorf("For error %v, expected %v, got %v", tt.err, tt.expected, result)
		}
	}
}

func TestCombineErrorMatchers(t *testing.T) {
	combined := CombineErrorMatchers(
		TimeoutError,
		NetworkError,
	)

	tests := []struct {
		err      error
		expected bool
	}{
		{errors.New("connection timeout"), true},
		{errors.New("connection refused"), true},
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
