package workflow

import (
	"strings"
	"sync"
	"time"
)

// ParallelFlow is an action that executes a series of actions concurrently.
// It runs all actions simultaneously, collecting their results and reporting any failures.
type ParallelFlow struct {
	FlowName string
	Actions  []Action
}

// NewParallelFlow creates a new ParallelFlow with the given name and initial actions.
// The name is used for identification in logs and reports.
func NewParallelFlow(name string, actions ...Action) *ParallelFlow {
	return &ParallelFlow{
		FlowName: name,
		Actions:  actions,
	}
}

// Name returns the name of the ParallelFlow, used for identification and logging.
func (pf *ParallelFlow) Name() string {
	return pf.FlowName
}

// Execute adds an action to the ParallelFlow.
// This method allows for adding actions that will be executed concurrently.
func (pf *ParallelFlow) Execute(action Action) *ParallelFlow {
	pf.Actions = append(pf.Actions, action)
	return pf
}

// Run performs the actions in the ParallelFlow concurrently using the given work context.
// It executes all actions simultaneously, waiting for all to complete before returning a combined report.
// The work context provides synchronized data sharing and cancellation capabilities.
func (pf *ParallelFlow) Run(wctx WorkContext) WorkReport {
	var wg sync.WaitGroup
	reports := make(chan WorkReport, len(pf.Actions))
	logger := wctx.Logger().With("flow", "ParallelFlow", "name", pf.FlowName)

	for _, action := range pf.Actions {
		wg.Add(1)
		go func(action Action) {
			defer wg.Done()
			startTime := time.Now()
			report := action.Run(wctx)
			elapsed := time.Since(startTime)

			switch report.Status {
			case StatusFailure:
				logger.Error("action failed", "action", action.Name(), "elapsed", elapsed, "errors", report.Errors)
			case StatusSkipped:
				logger.Info("action skipped", "action", action.Name(), "elapsed", elapsed)
			default:
				logger.Info("action completed", "action", action.Name(), "elapsed", elapsed)
			}
			reports <- report
		}(action)
	}

	wg.Wait()
	close(reports)

	combinedReport := NewCompletedWorkReport()
	var hasCompleted bool
	var outputs []string

	for report := range reports {
		// Merge events and metadata from all reports
		combinedReport.Events = append(combinedReport.Events, report.Events...)
		for k, v := range report.Metadata {
			combinedReport.SetMetadata(k, v)
		}

		switch report.Status {
		case StatusFailure:
			combinedReport.Status = StatusFailure
			combinedReport.Errors = append(combinedReport.Errors, report.Errors...)
		case StatusCompleted:
			hasCompleted = true
			// Collect outputs from all successful actions
			if report.Data != nil {
				content := extractContent(report.Data)
				if content != "" {
					outputs = append(outputs, content)
				}
			}
		}
	}

	// If no actions completed successfully, mark as skipped
	if !hasCompleted && combinedReport.Status != StatusFailure {
		combinedReport.Status = StatusSkipped
	}

	// Combine all outputs into a single response for chaining
	if len(outputs) > 0 {
		combinedContent := strings.Join(outputs, "\n\n---\n\n")
		// Create a simple response-like structure
		combinedData := struct{ Content string }{Content: combinedContent}
		combinedReport.Data = &combinedData
	}

	return combinedReport
}
