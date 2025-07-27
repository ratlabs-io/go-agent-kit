package workflow

import (
	"context"
	"sync"

	"github.com/ratlabs-io/go-agent-kit/pkg/tools"
)

// AgentContext extends WorkContext with agent-specific capabilities.
// It provides access to event callbacks, tool registry, and shared data
// for communication between agents in a workflow.
type AgentContext struct {
	*WorkContext
	Callbacks    *CallbackRegistry
	ToolRegistry tools.ToolRegistry
	SharedData   *sync.Map // Thread-safe map for agent-to-agent data sharing
}

// NewAgentContext creates a new AgentContext with the given base context,
// callback registry, and tool registry.
func NewAgentContext(ctx context.Context, callbacks *CallbackRegistry, toolRegistry tools.ToolRegistry) *AgentContext {
	agentCtx := &AgentContext{
		WorkContext:  NewWorkContext(ctx),
		Callbacks:    callbacks,
		ToolRegistry: toolRegistry,
		SharedData:   &sync.Map{},
	}
	
	// Store a reference to this AgentContext in the context for event emission
	agentCtx.WorkContext.Ctx = context.WithValue(agentCtx.WorkContext.Ctx, "agent_context", agentCtx)
	
	return agentCtx
}

// EmitEvent emits an event through the callback system.
func (ac *AgentContext) EmitEvent(event Event) {
	if ac.Callbacks != nil {
		ac.Callbacks.Emit(ac.Ctx, event)
	}
}

// EmitEventSync emits an event synchronously through the callback system.
func (ac *AgentContext) EmitEventSync(event Event) {
	if ac.Callbacks != nil {
		ac.Callbacks.EmitSync(ac.Ctx, event)
	}
}

// GetTool retrieves a tool from the tool registry.
func (ac *AgentContext) GetTool(name string) (tools.Tool, bool) {
	return ac.ToolRegistry.Get(name)
}

// ListTools returns all available tools from the registry.
func (ac *AgentContext) ListTools() []tools.Tool {
	return ac.ToolRegistry.List()
}

// SetSharedData stores a value in the shared data map using the given key.
func (ac *AgentContext) SetSharedData(key, value interface{}) {
	ac.SharedData.Store(key, value)
}

// GetSharedData retrieves a value from the shared data map using the given key.
func (ac *AgentContext) GetSharedData(key interface{}) (interface{}, bool) {
	return ac.SharedData.Load(key)
}

// DeleteSharedData removes a key-value pair from the shared data map.
func (ac *AgentContext) DeleteSharedData(key interface{}) {
	ac.SharedData.Delete(key)
}

// RangeSharedData iterates over all key-value pairs in the shared data map.
func (ac *AgentContext) RangeSharedData(fn func(key, value interface{}) bool) {
	ac.SharedData.Range(fn)
}

// Wait blocks until all running callbacks have completed.
// This should be called before a workflow or agent finishes to ensure
// callbacks like logging, metrics, and notifications complete.
func (ac *AgentContext) Wait() {
	if ac.Callbacks != nil {
		ac.Callbacks.Wait()
	}
}