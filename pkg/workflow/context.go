package workflow

import (
	"context"
	"sync"

	"github.com/ratlabs-io/go-agent-kit/pkg/constants"
)

// WorkContext provides a synchronized context for sharing data between actions in a workflow.
type WorkContext interface {
	// Context methods
	Context() context.Context
	Set(key, value interface{})
	Get(key interface{}) (value interface{}, ok bool)
	Logger() Logger
	WithLogger(logger Logger) WorkContext

	// Event and callback methods
	EmitEvent(event Event)
	EmitEventSync(event Event)
	Wait()
}

// DefaultWorkContext is the default implementation of WorkContext.
type DefaultWorkContext struct {
	ctx         context.Context
	mu          sync.RWMutex
	contextData map[interface{}]interface{}
	logger      Logger
	callbacks   *CallbackRegistry
}

// NewWorkContext creates a new WorkContext with the specified base context.
func NewWorkContext(ctx context.Context) WorkContext {
	return &DefaultWorkContext{
		ctx:         ctx,
		contextData: make(map[interface{}]interface{}),
		logger:      LoggerFromContext(ctx),
	}
}

// NewWorkContextWithCallbacks creates a new WorkContext with callbacks.
func NewWorkContextWithCallbacks(ctx context.Context, callbacks *CallbackRegistry) WorkContext {
	wc := &DefaultWorkContext{
		ctx:         ctx,
		contextData: make(map[interface{}]interface{}),
		logger:      LoggerFromContext(ctx),
		callbacks:   callbacks,
	}

	// Store a reference to this WorkContext in the context for event emission
	wc.ctx = context.WithValue(wc.ctx, constants.KeyWorkContext, wc)

	return wc
}

// NewWorkContextWithLogger creates a new WorkContext with a specific logger.
func NewWorkContextWithLogger(ctx context.Context, logger Logger) WorkContext {
	return &DefaultWorkContext{
		ctx:         ctx,
		contextData: make(map[interface{}]interface{}),
		logger:      logger,
	}
}

// Context returns the underlying context.
func (wc *DefaultWorkContext) Context() context.Context {
	return wc.ctx
}

// Set stores the value in the WorkContext for the given key.
func (wc *DefaultWorkContext) Set(key, value interface{}) {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	wc.contextData[key] = value
}

// Get retrieves the value from the WorkContext for the given key.
func (wc *DefaultWorkContext) Get(key interface{}) (value interface{}, ok bool) {
	wc.mu.RLock()
	defer wc.mu.RUnlock()
	value, ok = wc.contextData[key]
	return
}

// Logger returns the logger associated with this WorkContext.
func (wc *DefaultWorkContext) Logger() Logger {
	return wc.logger
}

// WithLogger returns a new WorkContext with the specified logger.
func (wc *DefaultWorkContext) WithLogger(logger Logger) WorkContext {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	newCtx := &DefaultWorkContext{
		ctx:         wc.ctx,
		contextData: make(map[interface{}]interface{}),
		logger:      logger,
		callbacks:   wc.callbacks,
	}

	// Copy existing data
	for k, v := range wc.contextData {
		newCtx.contextData[k] = v
	}

	return newCtx
}

// EmitEvent emits an event through the callback system.
func (wc *DefaultWorkContext) EmitEvent(event Event) {
	if wc.callbacks != nil {
		wc.callbacks.Emit(wc.ctx, event)
	}
}

// EmitEventSync emits an event synchronously through the callback system.
func (wc *DefaultWorkContext) EmitEventSync(event Event) {
	if wc.callbacks != nil {
		wc.callbacks.EmitSync(wc.ctx, event)
	}
}

// Wait blocks until all running callbacks have completed.
// This should be called before a workflow or agent finishes to ensure
// callbacks like logging, metrics, and notifications complete.
func (wc *DefaultWorkContext) Wait() {
	if wc.callbacks != nil {
		wc.callbacks.Wait()
	}
}
