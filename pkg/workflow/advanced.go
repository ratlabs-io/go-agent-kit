package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ratlabs-io/go-agent-kit/pkg/constants"
)

// CircuitBreakerState represents the state of a circuit breaker.
type CircuitBreakerState int

const (
	// CircuitBreakerClosed means the circuit breaker is allowing requests through.
	CircuitBreakerClosed CircuitBreakerState = iota
	// CircuitBreakerOpen means the circuit breaker is rejecting requests.
	CircuitBreakerOpen
	// CircuitBreakerHalfOpen means the circuit breaker is testing if the service has recovered.
	CircuitBreakerHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern for retry operations.
type CircuitBreaker struct {
	name                string
	action              Action
	failureThreshold    int
	recoveryTimeout     time.Duration
	resetTimeout        time.Duration
	
	// State tracking
	mu             sync.RWMutex
	state          CircuitBreakerState
	failures       int
	lastFailTime   time.Time
	lastResetTime  time.Time
}

// NewCircuitBreaker creates a new circuit breaker.
func NewCircuitBreaker(name string, failureThreshold int, recoveryTimeout, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:             name,
		failureThreshold: failureThreshold,
		recoveryTimeout:  recoveryTimeout,
		resetTimeout:     resetTimeout,
		state:           CircuitBreakerClosed,
	}
}

// WithAction sets the action to be protected by the circuit breaker.
func (cb *CircuitBreaker) WithAction(action Action) *CircuitBreaker {
	cb.action = action
	return cb
}

// Name returns the name of the circuit breaker.
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

// Run executes the action with circuit breaker protection.
func (cb *CircuitBreaker) Run(wctx WorkContext) WorkReport {
	if cb.action == nil {
		return NewFailedWorkReport(fmt.Errorf("circuit breaker %s: no action specified", cb.name))
	}

	logger := wctx.Logger()
	
	// Check if we should allow the request
	if !cb.allowRequest() {
		logger.Debug("Circuit breaker request rejected", "name", cb.name, "state", cb.getState())
		return NewFailedWorkReport(fmt.Errorf("circuit breaker %s is open", cb.name))
	}
	
	logger.Debug("Circuit breaker executing action", "name", cb.name, "state", cb.getState())
	
	// Execute the action
	report := cb.action.Run(wctx)
	
	// Update circuit breaker state based on result
	if report.Status == StatusCompleted {
		cb.onSuccess()
		logger.Debug("Circuit breaker action succeeded", "name", cb.name)
	} else {
		cb.onFailure()
		logger.Debug("Circuit breaker action failed", "name", cb.name)
	}
	
	return report
}

// allowRequest determines if a request should be allowed through.
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	now := time.Now()
	
	switch cb.state {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		if now.Sub(cb.lastFailTime) > cb.recoveryTimeout {
			cb.state = CircuitBreakerHalfOpen
			cb.lastResetTime = now
			return true
		}
		return false
	case CircuitBreakerHalfOpen:
		return true
	default:
		return false
	}
}

// onSuccess is called when the action succeeds.
func (cb *CircuitBreaker) onSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.failures = 0
	if cb.state == CircuitBreakerHalfOpen {
		cb.state = CircuitBreakerClosed
	}
}

// onFailure is called when the action fails.
func (cb *CircuitBreaker) onFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.failures++
	cb.lastFailTime = time.Now()
	
	if cb.failures >= cb.failureThreshold {
		cb.state = CircuitBreakerOpen
	}
}

// getState returns the current state of the circuit breaker.
func (cb *CircuitBreaker) getState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetMetrics returns current metrics for the circuit breaker.
func (cb *CircuitBreaker) GetMetrics() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	return map[string]interface{}{
		"state":            cb.state,
		"failures":         cb.failures,
		"failure_threshold": cb.failureThreshold,
		"last_fail_time":   cb.lastFailTime,
		"last_reset_time":  cb.lastResetTime,
	}
}

// TimeoutWrapper wraps any action with a timeout.
type TimeoutWrapper struct {
	name    string
	action  Action
	timeout time.Duration
}

// NewTimeoutWrapper creates a new timeout wrapper.
func NewTimeoutWrapper(name string, timeout time.Duration) *TimeoutWrapper {
	return &TimeoutWrapper{
		name:    name,
		timeout: timeout,
	}
}

// WithAction sets the action to be wrapped with timeout.
func (tw *TimeoutWrapper) WithAction(action Action) *TimeoutWrapper {
	tw.action = action
	return tw
}

// Name returns the name of the timeout wrapper.
func (tw *TimeoutWrapper) Name() string {
	return tw.name
}

// Run executes the action with timeout protection.
func (tw *TimeoutWrapper) Run(wctx WorkContext) WorkReport {
	if tw.action == nil {
		return NewFailedWorkReport(fmt.Errorf("timeout wrapper %s: no action specified", tw.name))
	}

	logger := wctx.Logger()
	logger.Debug("Timeout wrapper executing action", "name", tw.name, "timeout", tw.timeout)
	
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(wctx.Context(), tw.timeout)
	defer cancel()
	
	// Create a new work context with the timeout context
	timeoutWctx := NewWorkContext(ctx)
	
	// Copy data from original context
	if originalWctx, ok := wctx.(*DefaultWorkContext); ok {
		originalWctx.mu.RLock()
		for k, v := range originalWctx.contextData {
			timeoutWctx.Set(k, v)
		}
		originalWctx.mu.RUnlock()
	}
	
	// Use a channel to capture the result
	resultChan := make(chan WorkReport, 1)
	
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Timeout wrapper action panicked", "name", tw.name, "panic", r)
				resultChan <- NewFailedWorkReport(fmt.Errorf("action panicked: %v", r))
			}
		}()
		
		result := tw.action.Run(timeoutWctx)
		resultChan <- result
	}()
	
	select {
	case result := <-resultChan:
		logger.Debug("Timeout wrapper action completed", "name", tw.name)
		return result
	case <-ctx.Done():
		logger.Error("Timeout wrapper action timed out", "name", tw.name, "timeout", tw.timeout)
		return NewFailedWorkReport(fmt.Errorf("action timed out after %v", tw.timeout))
	}
}

// ParallelErrorCollector runs multiple actions in parallel and collects all errors.
type ParallelErrorCollector struct {
	name    string
	actions []Action
}

// NewParallelErrorCollector creates a new parallel error collector.
func NewParallelErrorCollector(name string) *ParallelErrorCollector {
	return &ParallelErrorCollector{
		name:    name,
		actions: make([]Action, 0),
	}
}

// AddAction adds an action to be executed in parallel.
func (pec *ParallelErrorCollector) AddAction(action Action) *ParallelErrorCollector {
	pec.actions = append(pec.actions, action)
	return pec
}

// AddActions adds multiple actions to be executed in parallel.
func (pec *ParallelErrorCollector) AddActions(actions ...Action) *ParallelErrorCollector {
	pec.actions = append(pec.actions, actions...)
	return pec
}

// Name returns the name of the parallel error collector.
func (pec *ParallelErrorCollector) Name() string {
	return pec.name
}

// Run executes all actions in parallel and collects all errors.
func (pec *ParallelErrorCollector) Run(wctx WorkContext) WorkReport {
	if len(pec.actions) == 0 {
		return NewFailedWorkReport(fmt.Errorf("parallel error collector %s: no actions specified", pec.name))
	}

	logger := wctx.Logger()
	logger.Debug("Starting parallel error collector", "type", constants.FlowTypeParallel, "name", pec.name, "actions", len(pec.actions))
	
	var wg sync.WaitGroup
	resultChan := make(chan WorkReport, len(pec.actions))
	
	// Execute all actions in parallel
	for i, action := range pec.actions {
		wg.Add(1)
		go func(index int, act Action) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Parallel error collector action panicked", "name", pec.name, "index", index, "panic", r)
					resultChan <- NewFailedWorkReport(fmt.Errorf("action %d panicked: %v", index, r))
				}
			}()
			
			logger.Debug("Parallel error collector starting action", "name", pec.name, "index", index, "action", act.Name())
			result := act.Run(wctx)
			logger.Debug("Parallel error collector completed action", "name", pec.name, "index", index, "action", act.Name(), "status", result.Status)
			resultChan <- result
		}(i, action)
	}
	
	// Wait for all actions to complete
	wg.Wait()
	close(resultChan)
	
	// Collect all results
	var allErrors []error
	var allEvents []interface{}
	var allMetadata = make(map[string]interface{})
	successCount := 0
	
	actionIndex := 0
	for result := range resultChan {
		if result.Status == StatusCompleted {
			successCount++
		}
		
		// Collect errors
		allErrors = append(allErrors, result.Errors...)
		
		// Collect events
		allEvents = append(allEvents, result.Events...)
		
		// Collect metadata with action prefix
		for k, v := range result.Metadata {
			allMetadata[fmt.Sprintf("action_%d_%s", actionIndex, k)] = v
		}
		actionIndex++
	}
	
	// Add summary metadata
	allMetadata["total_actions"] = len(pec.actions)
	allMetadata["successful_actions"] = successCount
	allMetadata["failed_actions"] = len(pec.actions) - successCount
	allMetadata["total_errors"] = len(allErrors)
	
	logger.Debug("Completed parallel error collector", "type", constants.FlowTypeParallel, "name", pec.name, "successful", successCount, "total", len(pec.actions))
	
	// Determine overall status
	finalReport := WorkReport{
		Errors:   allErrors,
		Events:   allEvents,
		Metadata: allMetadata,
	}
	
	if len(allErrors) == 0 {
		finalReport.Status = StatusCompleted
	} else {
		finalReport.Status = StatusFailure
	}
	
	return finalReport
}