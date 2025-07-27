package workflow

import (
	"fmt"
	"log/slog"
	"time"
)

// Case represents a condition-action pair in a SwitchFlow.
type Case struct {
	Condition Predicate
	Action    Action
}

// SwitchFlow is an action that executes different actions based on multiple conditions.
// It evaluates conditions in order and runs the action for the first condition that evaluates to true.
type SwitchFlow struct {
	FlowName     string
	Cases        []Case
	DefaultAction Action // Optional - can be nil
	Log          *slog.Logger
}

// NewSwitchFlow creates a new SwitchFlow with the given name, cases, and optional default action.
// The defaultAction is executed if no case conditions evaluate to true.
func NewSwitchFlow(name string, cases []Case, defaultAction Action) *SwitchFlow {
	log := slog.With("flow", "SwitchFlow")
	return &SwitchFlow{
		FlowName:     name,
		Cases:        cases,
		DefaultAction: defaultAction,
		Log:          log,
	}
}

// NewSwitchFlowBuilder creates a new SwitchFlow builder for fluent construction.
func NewSwitchFlowBuilder(name string) *SwitchFlowBuilder {
	return &SwitchFlowBuilder{
		flow: &SwitchFlow{
			FlowName: name,
			Cases:    []Case{},
			Log:      slog.With("flow", "SwitchFlow"),
		},
	}
}

// SwitchFlowBuilder provides a fluent interface for building SwitchFlow instances.
type SwitchFlowBuilder struct {
	flow *SwitchFlow
}

// Case adds a condition-action case to the SwitchFlow.
func (b *SwitchFlowBuilder) Case(condition Predicate, action Action) *SwitchFlowBuilder {
	b.flow.Cases = append(b.flow.Cases, Case{
		Condition: condition,
		Action:    action,
	})
	return b
}

// Default sets the default action to execute if no cases match.
func (b *SwitchFlowBuilder) Default(action Action) *SwitchFlowBuilder {
	b.flow.DefaultAction = action
	return b
}

// Build returns the constructed SwitchFlow.
func (b *SwitchFlowBuilder) Build() *SwitchFlow {
	return b.flow
}

// Name returns the name of the SwitchFlow, used for identification and logging.
func (sf *SwitchFlow) Name() string {
	return sf.FlowName
}

// Run executes the SwitchFlow by evaluating conditions in order and running the first matching action.
// If no conditions match and a default action is provided, runs the default action.
// If no conditions match and no default action is provided, returns a skipped report.
func (sf *SwitchFlow) Run(wctx *WorkContext) WorkReport {
	startTime := time.Now()
	
	// Evaluate each case in order
	for i, caseItem := range sf.Cases {
		conditionStart := time.Now()
		result, err := caseItem.Condition(wctx)
		conditionElapsed := time.Since(conditionStart)
		
		if err != nil {
			elapsed := time.Since(startTime)
			sf.Log.Error("case condition evaluation failed", 
				"flow", sf.FlowName, 
				"case", i, 
				"elapsed", elapsed, 
				"condition_elapsed", conditionElapsed,
				"error", err)
			return WorkReport{
				Status:   StatusFailure,
				Errors:   []error{fmt.Errorf("case %d condition evaluation failed: %w", i, err)},
				Events:   []interface{}{},
				Metadata: make(map[string]interface{}),
			}
		}

		sf.Log.Debug("case condition evaluated", 
			"flow", sf.FlowName, 
			"case", i, 
			"result", result, 
			"condition_elapsed", conditionElapsed)

		if result {
			elapsed := time.Since(startTime)
			sf.Log.Info("case matched, executing action", 
				"flow", sf.FlowName, 
				"case", i, 
				"action", caseItem.Action.Name(),
				"evaluation_elapsed", elapsed)
			return caseItem.Action.Run(wctx)
		}
	}

	// No cases matched, try default action
	if sf.DefaultAction != nil {
		elapsed := time.Since(startTime)
		sf.Log.Info("no cases matched, executing default action", 
			"flow", sf.FlowName, 
			"action", sf.DefaultAction.Name(),
			"evaluation_elapsed", elapsed)
		return sf.DefaultAction.Run(wctx)
	}

	// No cases matched and no default action
	elapsed := time.Since(startTime)
	sf.Log.Info("no cases matched and no default action, skipping", 
		"flow", sf.FlowName,
		"evaluation_elapsed", elapsed)
	return NewSkippedWorkReport()
}