package workflow

// WorkStatus represents the status of an action.
type WorkStatus int

const (
	// StatusCompleted indicates that the action has completed successfully.
	StatusCompleted WorkStatus = iota
	// StatusFailure indicates that the action has encountered a failure.
	StatusFailure
	// StatusSkipped indicates that the action was skipped (e.g., in conditional flows).
	StatusSkipped
)

// WorkReport holds information about the status and errors of an action.
// Enhanced for agent support with Events and Metadata fields.
type WorkReport struct {
	Status   WorkStatus
	Errors   []error
	Data     interface{}            // Generic data payload
	Events   []interface{}          // Events to publish (will be typed when events package is added)
	Metadata map[string]interface{} // Agent metadata (token usage, timing, etc.)
}

// NewCompletedWorkReport returns a WorkReport with a StatusCompleted status.
func NewCompletedWorkReport() WorkReport {
	return WorkReport{
		Status:   StatusCompleted,
		Errors:   []error{},
		Events:   []interface{}{},
		Metadata: make(map[string]interface{}),
	}
}

// NewFailedWorkReport returns a WorkReport with the given error.
func NewFailedWorkReport(err error) WorkReport {
	return WorkReport{
		Status:   StatusFailure,
		Errors:   []error{err},
		Events:   []interface{}{},
		Metadata: make(map[string]interface{}),
	}
}

// NewSkippedWorkReport returns a WorkReport with a StatusSkipped status.
func NewSkippedWorkReport() WorkReport {
	return WorkReport{
		Status:   StatusSkipped,
		Errors:   []error{},
		Events:   []interface{}{},
		Metadata: make(map[string]interface{}),
	}
}

// AddError appends an error to the WorkReport.
func (wr *WorkReport) AddError(err error) {
	wr.Errors = append(wr.Errors, err)
}

// AddEvent appends an event to the WorkReport.
func (wr *WorkReport) AddEvent(event interface{}) {
	wr.Events = append(wr.Events, event)
}

// SetMetadata sets a metadata key-value pair.
func (wr *WorkReport) SetMetadata(key string, value interface{}) {
	if wr.Metadata == nil {
		wr.Metadata = make(map[string]interface{})
	}
	wr.Metadata[key] = value
}
