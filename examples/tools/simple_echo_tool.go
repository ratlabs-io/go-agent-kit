package builtin

import (
	"context"
	"fmt"
	"time"

	"github.com/ratlabs-io/go-agent-kit/pkg/tools"
)

// SimpleEchoTool demonstrates the SimpleTool interface with minimal code.
// This is much simpler than the full Tool interface but less flexible.
type SimpleEchoTool struct{}

// NewSimpleEchoTool creates a new simple echo tool.
func NewSimpleEchoTool() *SimpleEchoTool {
	return &SimpleEchoTool{}
}

// Name returns the name of the tool.
func (set *SimpleEchoTool) Name() string {
	return "simple_echo"
}

// Description returns a description of what the tool does.
func (set *SimpleEchoTool) Description() string {
	return "Simple echo tool that returns the input message with a timestamp."
}

// Execute performs the echo operation with minimal parameter handling.
func (set *SimpleEchoTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract message (required parameter)
	message, ok := params["message"].(string)
	if !ok {
		return nil, fmt.Errorf("message parameter is required and must be a string")
	}
	
	// Return simple result
	return map[string]interface{}{
		"message":   message,
		"timestamp": time.Now(),
		"tool":      "simple_echo",
	}, nil
}

// ToTool converts this SimpleTool to a full Tool interface.
// This demonstrates how to use the adapter.
func (set *SimpleEchoTool) ToTool() tools.Tool {
	return tools.WrapSimpleTool(set)
}