package workflow

import (
	"fmt"
	"time"

	"github.com/ratlabs-io/go-agent-kit/pkg/constants"
)

// SequentialFlow is an action that executes a series of actions in order.
// It runs each action one after another, stopping if any action fails.
type SequentialFlow struct {
	FlowName string
	Actions  []Action
}

// NewSequentialFlow creates a new SequentialFlow with the given name and initial actions.
// The name is used for identification in logs and reports.
// Actions can be added later using the Then or Execute methods.
func NewSequentialFlow(name string, actions ...Action) *SequentialFlow {
	return &SequentialFlow{
		FlowName: name,
		Actions:  actions,
	}
}

// Name returns the name of the SequentialFlow, used for identification and logging.
func (sf *SequentialFlow) Name() string {
	return sf.FlowName
}

// Then appends a subsequent action to the SequentialFlow.
// This method allows for chaining actions to be executed in sequence.
func (sf *SequentialFlow) Then(action Action) *SequentialFlow {
	sf.Actions = append(sf.Actions, action)
	return sf
}

// Execute adds an action to the SequentialFlow.
// It is a synonym for Then, providing an alternative naming for adding actions.
func (sf *SequentialFlow) Execute(action Action) *SequentialFlow {
	return sf.Then(action)
}

// ThenChain adds an action where the previous action's output becomes the user_input.
// This rotates the context so each agent gets the previous agent's output as input.
func (sf *SequentialFlow) ThenChain(action Action) *SequentialFlow {
	chainAction := NewActionFunc(action.Name()+"_chain", func(ctx WorkContext) WorkReport {
		// Get previous output and set as user_input
		if prevOutput, ok := ctx.Get(constants.KeyPreviousOutput); ok {
			ctx.Set(constants.KeyUserInput, prevOutput)
		}

		// Run the action
		report := action.Run(ctx)
		if report.Status != StatusCompleted {
			return report
		}

		// Store this action's output for the next action
		if report.Data != nil {
			content := extractContent(report.Data)
			ctx.Set(constants.KeyPreviousOutput, content)
		}

		return report
	})

	return sf.Then(chainAction)
}

// ThenAccumulate adds an action where the previous output is accumulated with the original input.
// This creates a snowball effect where each agent gets more context.
func (sf *SequentialFlow) ThenAccumulate(action Action) *SequentialFlow {
	accumulateAction := NewActionFunc(action.Name()+"_accumulate", func(ctx WorkContext) WorkReport {
		// Get original input if not already stored
		if _, ok := ctx.Get(constants.KeyOriginalInput); !ok {
			if userInput, exists := ctx.Get(constants.KeyUserInput); exists {
				ctx.Set(constants.KeyOriginalInput, userInput)
			}
		}

		// Build accumulated input
		var accumulated string
		if original, ok := ctx.Get(constants.KeyOriginalInput); ok {
			accumulated = fmt.Sprintf("Original request: %v", original)
		}

		if prevOutput, ok := ctx.Get(constants.KeyPreviousOutput); ok {
			accumulated += fmt.Sprintf("\n\nPrevious output: %v", prevOutput)
		}

		ctx.Set(constants.KeyUserInput, accumulated)

		// Run the action
		report := action.Run(ctx)
		if report.Status != StatusCompleted {
			return report
		}

		// Store this action's output for the next action
		if report.Data != nil {
			content := extractContent(report.Data)
			ctx.Set(constants.KeyPreviousOutput, content)
		}

		return report
	})

	return sf.Then(accumulateAction)
}

// Run performs the actions in the SequentialFlow using the given work context.
// It executes each action in sequence, stopping if any action fails.
// The work context provides synchronized data sharing and cancellation capabilities.
func (sf *SequentialFlow) Run(wctx WorkContext) WorkReport {
	report := NewCompletedWorkReport()
	logger := wctx.Logger().With("flow", "SequentialFlow", "name", sf.FlowName)

	for _, action := range sf.Actions {
		startTime := time.Now()
		actionReport := action.Run(wctx)
		elapsed := time.Since(startTime)

		// Merge events and metadata from action report
		report.Events = append(report.Events, actionReport.Events...)
		for k, v := range actionReport.Metadata {
			report.SetMetadata(k, v)
		}

		if actionReport.Status == StatusFailure {
			logger.Error("action failed", "action", action.Name(), "elapsed", elapsed, "errors", actionReport.Errors)
			report.Status = StatusFailure
			report.Errors = append(report.Errors, fmt.Errorf("%s failed", action.Name()))
			report.Errors = append(report.Errors, actionReport.Errors...)
			return report
		}

		if actionReport.Status == StatusSkipped {
			logger.Info("action skipped", "action", action.Name(), "elapsed", elapsed)
			// Continue with next action for skipped actions
			continue
		}

		logger.Info("action completed", "action", action.Name(), "elapsed", elapsed)

		// Store action output for potential chaining
		if actionReport.Data != nil {
			content := extractContent(actionReport.Data)
			wctx.Set(constants.KeyPreviousOutput, content)
			// For the last successful action, preserve its data
			report.Data = actionReport.Data
		}
	}

	return report
}
