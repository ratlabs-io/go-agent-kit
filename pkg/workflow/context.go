package workflow

import (
	"context"
	"sync"
)

// WorkContext provides a synchronized context for sharing data between actions in a workflow.
type WorkContext struct {
	Ctx         context.Context
	mu          sync.RWMutex
	ContextData map[interface{}]interface{}
	logger      Logger
}

// NewWorkContext creates a new WorkContext with the specified base context.
func NewWorkContext(ctx context.Context) *WorkContext {
	return &WorkContext{
		Ctx:         ctx,
		ContextData: make(map[interface{}]interface{}),
		logger:      LoggerFromContext(ctx),
	}
}

// NewWorkContextWithLogger creates a new WorkContext with a specific logger.
func NewWorkContextWithLogger(ctx context.Context, logger Logger) *WorkContext {
	return &WorkContext{
		Ctx:         ctx,
		ContextData: make(map[interface{}]interface{}),
		logger:      logger,
	}
}

// Set stores the value in the WorkContext for the given key.
func (wc *WorkContext) Set(key, value interface{}) {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	wc.ContextData[key] = value
}

// Get retrieves the value from the WorkContext for the given key.
func (wc *WorkContext) Get(key interface{}) (value interface{}, ok bool) {
	wc.mu.RLock()
	defer wc.mu.RUnlock()
	value, ok = wc.ContextData[key]
	return
}

// Logger returns the logger associated with this WorkContext.
func (wc *WorkContext) Logger() Logger {
	return wc.logger
}

// WithLogger returns a new WorkContext with the specified logger.
func (wc *WorkContext) WithLogger(logger Logger) *WorkContext {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	
	newCtx := &WorkContext{
		Ctx:         wc.Ctx,
		ContextData: make(map[interface{}]interface{}),
		logger:      logger,
	}
	
	// Copy existing data
	for k, v := range wc.ContextData {
		newCtx.ContextData[k] = v
	}
	
	return newCtx
}