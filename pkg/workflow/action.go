package workflow

// Action represents a single unit of work within a workflow.
// Implementations of Action define specific tasks or operations that can be executed
// as part of a larger workflow, such as sequential or parallel flows.
type Action interface {
	// Name returns the name of the action, used for identification and logging.
	Name() string
	// Run executes the action with the given work context and returns a report.
	// The work context provides synchronized data sharing and cancellation capabilities.
	// All errors should be reported via the WorkReport.Status and WorkReport.Errors fields.
	Run(wctx *WorkContext) WorkReport
}