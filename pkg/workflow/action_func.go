package workflow

import "fmt"

// ActionFunc creates a simple Action from a function.
// This allows for quick action creation without implementing the full Action interface.
type ActionFunc struct {
	name string
	fn   func(*WorkContext) WorkReport
}

// NewActionFunc creates a new ActionFunc with the given name and function.
// The function should perform its work and return a WorkReport.
// If the function panics, it will be caught and turned into a failed report.
// Example:
//   action := workflow.NewActionFunc("process", func(ctx *workflow.WorkContext) workflow.WorkReport {
//       // Do something with the context
//       if input, ok := ctx.Get("input"); ok {
//           ctx.Set("output", process(input))
//           return workflow.NewCompletedWorkReport()
//       }
//       return workflow.NewFailedWorkReport(fmt.Errorf("no input provided"))
//   })
func NewActionFunc(name string, fn func(*WorkContext) WorkReport) *ActionFunc {
	return &ActionFunc{
		name: name,
		fn:   fn,
	}
}

// Name returns the name of the action.
func (af *ActionFunc) Name() string {
	return af.name
}

// Run executes the function with the given work context.
// Automatically handles success/failure and panic recovery.
func (af *ActionFunc) Run(wctx *WorkContext) (report WorkReport) {
	if af.fn == nil {
		return NewFailedWorkReport(fmt.Errorf("no function provided for action %s", af.name))
	}
	
	// Catch panics and turn them into failed reports
	defer func() {
		if r := recover(); r != nil {
			report = NewFailedWorkReport(fmt.Errorf("action %s failed: %v", af.name, r))
		}
	}()
	
	// Execute the function and return its report
	return af.fn(wctx)
}