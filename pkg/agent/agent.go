package agent

import (
	"github.com/ratlabs-io/go-agent-kit/pkg/tools"
	"github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

// AgentType represents the type of an agent.
type AgentType string

const (
	// TypeChat represents a simple chat agent that performs 1-hop LLM completions.
	TypeChat AgentType = "chat"

	// TypeTool represents a tool-calling agent that may use internal workflows
	// for multi-step tool execution.
	TypeTool AgentType = "tool"

	// TypeWorkflow represents an agent that wraps a complex internal workflow.
	TypeWorkflow AgentType = "workflow"
)

// Agent represents a specialized workflow action that adds agent-specific capabilities
// such as LLM integration, tool usage, and event publishing.
// All agents implement the workflow.Action interface to fit seamlessly into workflows.
type Agent interface {
	// Embed the workflow.Action interface
	workflow.Action

	// Type returns the type of the agent.
	Type() AgentType

	// Tools returns the list of tools available to this agent.
	Tools() []tools.Tool

	// Configure allows runtime configuration of the agent.
	Configure(config map[string]interface{}) error
}
