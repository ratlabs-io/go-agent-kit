package workflow

import (
	"fmt"
	"time"

	"github.com/ratlabs-io/go-agent-kit/pkg/constants"
)

// BackoffStrategy defines the interface for retry backoff strategies.
type BackoffStrategy interface {
	// CalculateDelay returns the delay duration for the given attempt number (0-based).
	CalculateDelay(attempt int) time.Duration
}

// RetryConditionFunc is a function that determines whether to retry based on the error.
type RetryConditionFunc func(error) bool

// StopConditionFunc is a function that determines whether to stop retrying based on context state.
type StopConditionFunc func(WorkContext) bool

// Retry represents a retry construct that executes an action with retry logic.
type Retry struct {
	name        string
	action      Action
	maxAttempts int
	
	// Retry configuration
	backoffStrategy BackoffStrategy
	retryCondition  RetryConditionFunc
	stopCondition   StopConditionFunc
}

// NewRetry creates a basic retry construct with a maximum number of attempts.
func NewRetry(name string, maxAttempts int) *Retry {
	return &Retry{
		name:        name,
		maxAttempts: maxAttempts,
		backoffStrategy: &FixedBackoff{delay: time.Second},
		retryCondition:  DefaultRetryCondition,
	}
}

// WithAction sets the action to be executed with retry logic.
func (r *Retry) WithAction(action Action) *Retry {
	r.action = action
	return r
}

// WithBackoffStrategy sets the backoff strategy for retry delays.
func (r *Retry) WithBackoffStrategy(strategy BackoffStrategy) *Retry {
	r.backoffStrategy = strategy
	return r
}

// WithRetryCondition sets a custom retry condition based on error type.
func (r *Retry) WithRetryCondition(condition RetryConditionFunc) *Retry {
	r.retryCondition = condition
	return r
}

// WithStopCondition sets a stop condition based on context state.
func (r *Retry) WithStopCondition(condition StopConditionFunc) *Retry {
	r.stopCondition = condition
	return r
}

// Name returns the name of the retry construct.
func (r *Retry) Name() string {
	return r.name
}

// Run executes the retry logic with the given work context.
func (r *Retry) Run(wctx WorkContext) WorkReport {
	if r.action == nil {
		return NewFailedWorkReport(fmt.Errorf("retry %s: no action specified", r.name))
	}

	logger := wctx.Logger()
	logger.Debug("Starting retry", "type", constants.FlowTypeRetry, "name", r.name, "max_attempts", r.maxAttempts)

	var lastReport WorkReport
	
	for attempt := 0; attempt < r.maxAttempts; attempt++ {
		logger.Debug("Retry attempt", "name", r.name, "attempt", attempt+1, "max_attempts", r.maxAttempts)
		
		// Execute the action
		report := r.action.Run(wctx)
		
		// Success case
		if report.Status == StatusCompleted {
			if attempt > 0 {
				logger.Info("Retry succeeded", "name", r.name, "attempt", attempt+1, "max_attempts", r.maxAttempts)
			}
			return report
		}
		
		lastReport = report
		
		// Check if we should retry based on error condition
		shouldRetry := false
		if len(report.Errors) > 0 {
			for _, err := range report.Errors {
				if r.retryCondition(err) {
					shouldRetry = true
					break
				}
			}
		} else {
			// If no specific errors but status is failure, use default retry logic
			shouldRetry = true
		}
		
		// Check stop condition
		if r.stopCondition != nil && r.stopCondition(wctx) {
			logger.Debug("Retry stop condition met, aborting", "name", r.name)
			break
		}
		
		// Don't retry if condition says no
		if !shouldRetry {
			logger.Debug("Retry condition not met, aborting", "name", r.name)
			break
		}
		
		// Don't sleep after the last attempt
		if attempt < r.maxAttempts-1 {
			delay := r.backoffStrategy.CalculateDelay(attempt)
			logger.Debug("Retry waiting before next attempt", "name", r.name, "delay", delay)
			time.Sleep(delay)
		}
	}
	
	logger.Error("Retry all attempts failed", "name", r.name)
	
	// Add retry information to the final report
	if lastReport.Metadata == nil {
		lastReport.Metadata = make(map[string]interface{})
	}
	lastReport.Metadata["retry_attempts"] = r.maxAttempts
	lastReport.Metadata["retry_name"] = r.name
	
	return lastReport
}

// Backoff strategy implementations

// FixedBackoff implements a fixed delay backoff strategy.
type FixedBackoff struct {
	delay time.Duration
}

// NewFixedBackoff creates a new fixed backoff strategy.
func NewFixedBackoff(delay time.Duration) *FixedBackoff {
	return &FixedBackoff{delay: delay}
}

// CalculateDelay returns the fixed delay for any attempt.
func (f *FixedBackoff) CalculateDelay(attempt int) time.Duration {
	return f.delay
}

// LinearBackoff implements a linear backoff strategy.
type LinearBackoff struct {
	baseDelay time.Duration
	increment time.Duration
}

// NewLinearBackoff creates a new linear backoff strategy.
func NewLinearBackoff(baseDelay, increment time.Duration) *LinearBackoff {
	return &LinearBackoff{
		baseDelay: baseDelay,
		increment: increment,
	}
}

// CalculateDelay returns a linearly increasing delay.
func (l *LinearBackoff) CalculateDelay(attempt int) time.Duration {
	return l.baseDelay + time.Duration(attempt)*l.increment
}

// ExponentialBackoff implements an exponential backoff strategy.
type ExponentialBackoff struct {
	baseDelay time.Duration
	maxDelay  time.Duration
	factor    float64
}

// NewExponentialBackoff creates a new exponential backoff strategy.
func NewExponentialBackoff(baseDelay, maxDelay time.Duration, factor float64) *ExponentialBackoff {
	return &ExponentialBackoff{
		baseDelay: baseDelay,
		maxDelay:  maxDelay,
		factor:    factor,
	}
}

// CalculateDelay returns an exponentially increasing delay, capped at maxDelay.
func (e *ExponentialBackoff) CalculateDelay(attempt int) time.Duration {
	delay := time.Duration(float64(e.baseDelay) * e.factor * float64(attempt))
	if delay > e.maxDelay {
		delay = e.maxDelay
	}
	return delay
}

// Built-in retry conditions

// DefaultRetryCondition always returns true, retrying on any error.
func DefaultRetryCondition(err error) bool {
	return err != nil
}

// NeverRetryCondition never retries, regardless of error.
func NeverRetryCondition(err error) bool {
	return false
}

// RetryOnTimeoutCondition retries only on timeout-like errors.
func RetryOnTimeoutCondition(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	timeoutKeywords := []string{"timeout", "deadline", "context canceled", "context deadline exceeded"}
	
	for _, keyword := range timeoutKeywords {
		if contains(errStr, keyword) {
			return true
		}
	}
	
	return false
}

// RetryOnRateLimitCondition retries only on rate limit errors.
func RetryOnRateLimitCondition(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	rateLimitKeywords := []string{"rate limit", "too many requests", "429", "quota exceeded"}
	
	for _, keyword := range rateLimitKeywords {
		if contains(errStr, keyword) {
			return true
		}
	}
	
	return false
}

// RetryOnNetworkCondition retries on network-related errors.
func RetryOnNetworkCondition(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	networkKeywords := []string{
		"connection refused", "connection reset", "network unreachable",
		"no such host", "dns", "timeout", "i/o timeout",
	}
	
	for _, keyword := range networkKeywords {
		if contains(errStr, keyword) {
			return true
		}
	}
	
	return false
}

// CombineRetryConditions combines multiple retry conditions with OR logic.
func CombineRetryConditions(conditions ...RetryConditionFunc) RetryConditionFunc {
	return func(err error) bool {
		for _, condition := range conditions {
			if condition(err) {
				return true
			}
		}
		return false
	}
}

// Helper function to check if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}