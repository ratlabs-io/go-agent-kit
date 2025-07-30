package workflow

import (
	"fmt"
	"reflect"

	"github.com/ratlabs-io/go-agent-kit/pkg/constants"
)

// ErrorTypeMatcherFunc is a function that determines if an error matches a specific type or condition.
type ErrorTypeMatcherFunc func(error) bool

// ErrorHandlerAction is an action that handles errors.
type ErrorHandlerAction interface {
	Action
	// HandleError is called with the original error that was caught.
	HandleError(wctx WorkContext, err error) WorkReport
}

// DefaultErrorHandlerAction is a simple error handler that just logs and returns.
type DefaultErrorHandlerAction struct {
	name        string
	handlerFunc func(WorkContext, error) WorkReport
}

func NewDefaultErrorHandlerAction(name string, handlerFunc func(WorkContext, error) WorkReport) *DefaultErrorHandlerAction {
	return &DefaultErrorHandlerAction{
		name:        name,
		handlerFunc: handlerFunc,
	}
}

func (d *DefaultErrorHandlerAction) Name() string {
	return d.name
}

func (d *DefaultErrorHandlerAction) Run(wctx WorkContext) WorkReport {
	// This should not be called directly
	return NewCompletedWorkReport()
}

func (d *DefaultErrorHandlerAction) HandleError(wctx WorkContext, err error) WorkReport {
	if d.handlerFunc != nil {
		return d.handlerFunc(wctx, err)
	}
	return NewCompletedWorkReport()
}

// CatchHandler represents a catch block in the try-catch construct.
type CatchHandler struct {
	matcher ErrorTypeMatcherFunc
	action  ErrorHandlerAction
}

// TryCatch represents a try-catch-finally construct.
type TryCatch struct {
	name           string
	tryAction      Action
	catchHandlers  []CatchHandler
	catchAllAction ErrorHandlerAction
	finallyAction  Action
}

// NewTryCatch creates a new try-catch construct.
func NewTryCatch(name string) *TryCatch {
	return &TryCatch{
		name:          name,
		catchHandlers: make([]CatchHandler, 0),
	}
}

// WithTryAction sets the action to execute in the try block.
func (tc *TryCatch) WithTryAction(action Action) *TryCatch {
	tc.tryAction = action
	return tc
}

// Catch adds a catch handler for specific error types or conditions.
func (tc *TryCatch) Catch(matcher ErrorTypeMatcherFunc, action ErrorHandlerAction) *TryCatch {
	tc.catchHandlers = append(tc.catchHandlers, CatchHandler{
		matcher: matcher,
		action:  action,
	})
	return tc
}

// CatchAny adds a catch-all handler that handles any unhandled errors.
func (tc *TryCatch) CatchAny(action ErrorHandlerAction) *TryCatch {
	tc.catchAllAction = action
	return tc
}

// Finally adds a finally block that always executes regardless of success/failure.
func (tc *TryCatch) Finally(action Action) *TryCatch {
	tc.finallyAction = action
	return tc
}

// Name returns the name of the try-catch construct.
func (tc *TryCatch) Name() string {
	return tc.name
}

// Run executes the try-catch-finally logic.
func (tc *TryCatch) Run(wctx WorkContext) WorkReport {
	if tc.tryAction == nil {
		return NewFailedWorkReport(fmt.Errorf("try-catch %s: no try action specified", tc.name))
	}

	logger := wctx.Logger()
	logger.Debug("Starting try-catch", "type", constants.FlowTypeTryCatch, "name", tc.name)

	var finalReport WorkReport

	// Execute try block
	logger.Debug("Try-catch executing try block", "name", tc.name)
	tryReport := tc.tryAction.Run(wctx)

	// Handle success case
	if tryReport.Status == StatusCompleted {
		logger.Debug("Try-catch try block completed successfully", "name", tc.name)
		finalReport = tryReport
	} else {
		// Handle errors
		logger.Debug("Try-catch try block failed, checking catch handlers", "name", tc.name)
		finalReport = tc.handleErrors(wctx, tryReport)
	}

	// Execute finally block if present
	if tc.finallyAction != nil {
		logger.Debug("Try-catch executing finally block", "name", tc.name)
		finallyReport := tc.finallyAction.Run(wctx)

		// If finally block fails, that overrides the previous result
		if finallyReport.Status == StatusFailure {
			logger.Error("Try-catch finally block failed", "name", tc.name)
			// Combine errors from both try and finally
			if len(finalReport.Errors) > 0 {
				finallyReport.Errors = append(finallyReport.Errors, finalReport.Errors...)
			}
			finalReport = finallyReport
		}

		// Merge metadata from finally block
		if len(finallyReport.Metadata) > 0 {
			if finalReport.Metadata == nil {
				finalReport.Metadata = make(map[string]interface{})
			}
			for k, v := range finallyReport.Metadata {
				finalReport.Metadata["finally_"+k] = v
			}
		}
	}

	logger.Debug("Completed try-catch", "type", constants.FlowTypeTryCatch, "name", tc.name, "status", finalReport.Status)
	return finalReport
}

// handleErrors processes errors through the catch handlers.
func (tc *TryCatch) handleErrors(wctx WorkContext, tryReport WorkReport) WorkReport {
	logger := wctx.Logger()

	// Try specific catch handlers first
	for i, handler := range tc.catchHandlers {
		for _, err := range tryReport.Errors {
			if handler.matcher(err) {
				logger.Debug("Try-catch error matched catch handler", "name", tc.name, "handler_index", i)
				return handler.action.HandleError(wctx, err)
			}
		}
	}

	// Try catch-all handler
	if tc.catchAllAction != nil {
		logger.Debug("Try-catch using catch-all handler", "name", tc.name)
		// Pass the first error or create a generic one
		var err error
		if len(tryReport.Errors) > 0 {
			err = tryReport.Errors[0]
		} else {
			err = fmt.Errorf("action failed with status: %v", tryReport.Status)
		}
		return tc.catchAllAction.HandleError(wctx, err)
	}

	// No handlers matched, return original report
	logger.Debug("Try-catch no catch handlers matched, returning original error", "name", tc.name)
	return tryReport
}

// Error type matchers and predicates

// ErrorTypeEquals matches errors of a specific type.
func ErrorTypeEquals(targetType reflect.Type) ErrorTypeMatcherFunc {
	return func(err error) bool {
		if err == nil {
			return false
		}
		return reflect.TypeOf(err) == targetType
	}
}

// ErrorMessageContains matches errors whose message contains a specific substring.
func ErrorMessageContains(substring string) ErrorTypeMatcherFunc {
	return func(err error) bool {
		if err == nil {
			return false
		}
		return contains(err.Error(), substring)
	}
}

// ErrorMessageEquals matches errors with an exact message.
func ErrorMessageEquals(message string) ErrorTypeMatcherFunc {
	return func(err error) bool {
		if err == nil {
			return false
		}
		return err.Error() == message
	}
}

// AnyError matches any non-nil error.
func AnyError(err error) bool {
	return err != nil
}

// NoError matches nil errors (useful for testing).
func NoError(err error) bool {
	return err == nil
}

// CombineErrorMatchers combines multiple error matchers with OR logic.
func CombineErrorMatchers(matchers ...ErrorTypeMatcherFunc) ErrorTypeMatcherFunc {
	return func(err error) bool {
		for _, matcher := range matchers {
			if matcher(err) {
				return true
			}
		}
		return false
	}
}

// TimeoutError matches timeout-related errors.
func TimeoutError(err error) bool {
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

// NetworkError matches network-related errors.
func NetworkError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	networkKeywords := []string{
		"connection refused", "connection reset", "network unreachable",
		"no such host", "dns", "i/o timeout",
	}

	for _, keyword := range networkKeywords {
		if contains(errStr, keyword) {
			return true
		}
	}

	return false
}

// ValidationError matches validation-related errors.
func ValidationError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	validationKeywords := []string{"validation", "invalid", "malformed", "bad request"}

	for _, keyword := range validationKeywords {
		if contains(errStr, keyword) {
			return true
		}
	}

	return false
}
