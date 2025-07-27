package workflow

import (
	"context"
	"sync"
	"time"
)

// EventType represents the type of workflow/agent event.
type EventType string

const (
	EventAgentStarted   EventType = "agent.started"
	EventAgentCompleted EventType = "agent.completed"
	EventAgentFailed    EventType = "agent.failed"
	EventWorkflowStarted EventType = "workflow.started"
	EventWorkflowCompleted EventType = "workflow.completed"
	EventWorkflowFailed  EventType = "workflow.failed"
	EventToolCalled     EventType = "tool.called"
	EventToolCompleted  EventType = "tool.completed"
	EventToolFailed     EventType = "tool.failed"
)

// Event represents something that happened during workflow/agent execution.
type Event struct {
	Type      EventType              `json:"type"`
	Source    string                 `json:"source"`    // Agent or workflow name
	Timestamp time.Time              `json:"timestamp"`
	Payload   interface{}            `json:"payload"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// EventCallback is a function that gets called when events occur.
// Users can implement this to handle events however they want (logging, metrics, etc.)
type EventCallback func(ctx context.Context, event Event)

// CallbackRegistry holds event callbacks for workflows and agents.
type CallbackRegistry struct {
	callbacks []EventCallback
	wg        sync.WaitGroup
}

// NewCallbackRegistry creates a new callback registry.
func NewCallbackRegistry() *CallbackRegistry {
	return &CallbackRegistry{
		callbacks: make([]EventCallback, 0),
	}
}

// Add registers a new event callback.
func (cr *CallbackRegistry) Add(callback EventCallback) {
	if callback != nil {
		cr.callbacks = append(cr.callbacks, callback)
	}
}

// Emit sends an event to all registered callbacks.
func (cr *CallbackRegistry) Emit(ctx context.Context, event Event) {
	for _, callback := range cr.callbacks {
		cr.wg.Add(1)
		// Execute callbacks concurrently to avoid blocking
		go func(cb EventCallback) {
			defer cr.wg.Done()
			defer func() {
				// Recover from panics in user callbacks
				if r := recover(); r != nil {
					// Could log this if we had a logger, but keeping it simple
				}
			}()
			cb(ctx, event)
		}(callback)
	}
}

// EmitSync sends an event to all registered callbacks synchronously.
func (cr *CallbackRegistry) EmitSync(ctx context.Context, event Event) {
	for _, callback := range cr.callbacks {
		func(cb EventCallback) {
			defer func() {
				// Recover from panics in user callbacks
				if r := recover(); r != nil {
					// Could log this if we had a logger, but keeping it simple
				}
			}()
			cb(ctx, event)
		}(callback)
	}
}

// Wait blocks until all running callbacks have completed.
func (cr *CallbackRegistry) Wait() {
	cr.wg.Wait()
}