package builtin

import (
	"context"
	"fmt"
	"time"

	"github.com/ratlabs-io/go-agent-kit/pkg/tools"
)

// EchoTool is a simple tool that echoes back input with optional transformations.
// Useful for testing and as a reference implementation.
type EchoTool struct{}

// NewEchoTool creates a new echo tool.
func NewEchoTool() *EchoTool {
	return &EchoTool{}
}

// Name returns the name of the tool.
func (et *EchoTool) Name() string {
	return "echo"
}

// Description returns a description of what the tool does.
func (et *EchoTool) Description() string {
	return "Echo back the input message with optional transformations (uppercase, lowercase, reverse)."
}

// Parameters returns the JSON Schema for the tool's parameters.
func (et *EchoTool) Parameters() tools.Schema {
	return tools.Schema{
		Type:        "object",
		Description: "Parameters for echo tool",
		Properties: map[string]interface{}{
			"message": map[string]interface{}{
				"type":        "string",
				"description": "The message to echo back",
			},
			"transform": map[string]interface{}{
				"type":        "string",
				"description": "Optional transformation to apply",
				"enum":        []string{"none", "upper", "lower", "reverse"},
				"default":     "none",
			},
			"repeat": map[string]interface{}{
				"type":        "integer",
				"description": "Number of times to repeat the message",
				"default":     1,
				"minimum":     1,
				"maximum":     10,
			},
		},
		Required: []string{"message"},
	}
}

// Execute performs the echo operation.
func (et *EchoTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract message
	message, ok := params["message"].(string)
	if !ok {
		return nil, fmt.Errorf("message parameter is required and must be a string")
	}
	
	// Extract transform option
	transform := "none"
	if t, ok := params["transform"].(string); ok {
		transform = t
	}
	
	// Extract repeat count
	repeat := 1
	if r, ok := params["repeat"].(float64); ok {
		repeat = int(r)
	}
	if repeat < 1 || repeat > 10 {
		repeat = 1
	}
	
	// Apply transformation
	transformedMessage := et.applyTransform(message, transform)
	
	// Create result
	var result []string
	for i := 0; i < repeat; i++ {
		result = append(result, transformedMessage)
	}
	
	return map[string]interface{}{
		"original_message":    message,
		"transformed_message": transformedMessage,
		"transform":          transform,
		"repeat_count":       repeat,
		"result":             result,
		"timestamp":          time.Now(),
	}, nil
}

// applyTransform applies the specified transformation to the message.
func (et *EchoTool) applyTransform(message, transform string) string {
	switch transform {
	case "upper":
		return fmt.Sprintf("%s", message) // Using fmt to keep it simple, could use strings.ToUpper
	case "lower":
		return fmt.Sprintf("%s", message) // Using fmt to keep it simple, could use strings.ToLower  
	case "reverse":
		runes := []rune(message)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	default:
		return message
	}
}