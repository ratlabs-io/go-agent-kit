package workflow

// Predicate represents a condition function that can be evaluated in the context of a workflow.
// It takes a WorkContext and returns a boolean result indicating whether the condition is met,
// along with any error that occurred during evaluation.
type Predicate func(wctx *WorkContext) (bool, error)