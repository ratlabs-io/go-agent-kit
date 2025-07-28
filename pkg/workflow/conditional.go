package workflow

import (
	"fmt"
	"log/slog"
	"time"
)

// ConditionalFlow is an action that executes different actions based on a condition.
// It evaluates a predicate and runs either the ifTrue or ifFalse action accordingly.
type ConditionalFlow struct {
	FlowName  string
	Condition Predicate
	IfTrue    Action
	IfFalse   Action // Optional - can be nil
	Log       *slog.Logger
}

// NewConditionalFlow creates a new ConditionalFlow with the given name, condition, and actions.
// The ifFalse action is optional and can be nil.
func NewConditionalFlow(name string, condition Predicate, ifTrue Action, ifFalse Action) *ConditionalFlow {
	log := slog.With("flow", "ConditionalFlow")
	return &ConditionalFlow{
		FlowName:  name,
		Condition: condition,
		IfTrue:    ifTrue,
		IfFalse:   ifFalse,
		Log:       log,
	}
}

// Name returns the name of the ConditionalFlow, used for identification and logging.
func (cf *ConditionalFlow) Name() string {
	return cf.FlowName
}

// Run executes the ConditionalFlow by evaluating the condition and running the appropriate action.
// If the condition evaluates to true, runs the ifTrue action.
// If the condition evaluates to false and ifFalse is provided, runs the ifFalse action.
// If the condition evaluates to false and ifFalse is nil, returns a skipped report.
func (cf *ConditionalFlow) Run(wctx WorkContext) WorkReport {
	startTime := time.Now()
	
	// Evaluate the condition
	result, err := cf.Condition(wctx)
	if err != nil {
		elapsed := time.Since(startTime)
		cf.Log.Error("condition evaluation failed", "flow", cf.FlowName, "elapsed", elapsed, "error", err)
		return WorkReport{
			Status:   StatusFailure,
			Errors:   []error{fmt.Errorf("condition evaluation failed: %w", err)},
			Events:   []interface{}{},
			Metadata: make(map[string]interface{}),
		}
	}

	elapsed := time.Since(startTime)
	cf.Log.Info("condition evaluated", "flow", cf.FlowName, "result", result, "elapsed", elapsed)

	// Execute the appropriate action based on the condition result
	if result {
		if cf.IfTrue != nil {
			cf.Log.Info("executing ifTrue action", "flow", cf.FlowName, "action", cf.IfTrue.Name())
			return cf.IfTrue.Run(wctx)
		}
		// If no ifTrue action is provided, skip
		cf.Log.Info("no ifTrue action provided, skipping", "flow", cf.FlowName)
		return NewSkippedWorkReport()
	} else {
		if cf.IfFalse != nil {
			cf.Log.Info("executing ifFalse action", "flow", cf.FlowName, "action", cf.IfFalse.Name())
			return cf.IfFalse.Run(wctx)
		}
		// If no ifFalse action is provided, skip
		cf.Log.Info("no ifFalse action provided, skipping", "flow", cf.FlowName)
		return NewSkippedWorkReport()
	}
}