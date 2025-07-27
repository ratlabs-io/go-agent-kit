package tools

import (
	"fmt"
	"sync"
)

// DefaultToolRegistry provides a thread-safe implementation of ToolRegistry.
type DefaultToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// NewDefaultToolRegistry creates a new DefaultToolRegistry.
func NewDefaultToolRegistry() *DefaultToolRegistry {
	return &DefaultToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry.
func (r *DefaultToolRegistry) Register(tool Tool) error {
	if tool == nil {
		return fmt.Errorf("tool cannot be nil")
	}
	
	name := tool.Name()
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool %s is already registered", name)
	}
	
	r.tools[name] = tool
	return nil
}

// RegisterSimple adds a SimpleTool to the registry by automatically wrapping it.
// This is a convenience method for registering tools that implement SimpleTool.
func (r *DefaultToolRegistry) RegisterSimple(simple SimpleTool) error {
	return r.Register(WrapSimpleTool(simple))
}

// Get retrieves a tool by name.
func (r *DefaultToolRegistry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	tool, exists := r.tools[name]
	return tool, exists
}

// List returns all registered tools.
func (r *DefaultToolRegistry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	
	return tools
}

// Unregister removes a tool from the registry.
func (r *DefaultToolRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.tools[name]; !exists {
		return fmt.Errorf("tool %s is not registered", name)
	}
	
	delete(r.tools, name)
	return nil
}

// Names returns all registered tool names.
func (r *DefaultToolRegistry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	
	return names
}

// Count returns the number of registered tools.
func (r *DefaultToolRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	return len(r.tools)
}