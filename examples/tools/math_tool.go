package builtin

import (
	"context"
	"fmt"
	"math"

	"github.com/ratlabs-io/go-agent-kit/pkg/tools"
)

// MathTool provides basic mathematical operations.
// Useful for testing and as a reference implementation.
type MathTool struct{}

// NewMathTool creates a new math tool.
func NewMathTool() *MathTool {
	return &MathTool{}
}

// Name returns the name of the tool.
func (mt *MathTool) Name() string {
	return "math"
}

// Description returns a description of what the tool does.
func (mt *MathTool) Description() string {
	return "Perform basic mathematical operations: add, subtract, multiply, divide, power, sqrt."
}

// Parameters returns the JSON Schema for the tool's parameters.
func (mt *MathTool) Parameters() tools.Schema {
	return tools.Schema{
		Type:        "object",
		Description: "Parameters for mathematical operations",
		Properties: map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "The mathematical operation to perform",
				"enum":        []string{"add", "subtract", "multiply", "divide", "power", "sqrt", "abs"},
			},
			"a": map[string]interface{}{
				"type":        "number",
				"description": "First number (or the only number for unary operations)",
			},
			"b": map[string]interface{}{
				"type":        "number",
				"description": "Second number (for binary operations)",
			},
		},
		Required: []string{"operation", "a"},
	}
}

// Execute performs the mathematical operation.
func (mt *MathTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract operation
	operation, ok := params["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation parameter is required and must be a string")
	}

	// Extract first number
	a, ok := params["a"].(float64)
	if !ok {
		return nil, fmt.Errorf("parameter 'a' is required and must be a number")
	}

	// Extract second number (for binary operations)
	var b float64
	var hasBinary bool
	if bVal, ok := params["b"].(float64); ok {
		b = bVal
		hasBinary = true
	}

	// Perform operation
	var result float64

	switch operation {
	case "add":
		if !hasBinary {
			return nil, fmt.Errorf("add operation requires parameter 'b'")
		}
		result = a + b

	case "subtract":
		if !hasBinary {
			return nil, fmt.Errorf("subtract operation requires parameter 'b'")
		}
		result = a - b

	case "multiply":
		if !hasBinary {
			return nil, fmt.Errorf("multiply operation requires parameter 'b'")
		}
		result = a * b

	case "divide":
		if !hasBinary {
			return nil, fmt.Errorf("divide operation requires parameter 'b'")
		}
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = a / b

	case "power":
		if !hasBinary {
			return nil, fmt.Errorf("power operation requires parameter 'b'")
		}
		result = math.Pow(a, b)

	case "sqrt":
		if a < 0 {
			return nil, fmt.Errorf("cannot take square root of negative number")
		}
		result = math.Sqrt(a)

	case "abs":
		result = math.Abs(a)

	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}

	// Check for invalid results
	if math.IsNaN(result) {
		return nil, fmt.Errorf("operation resulted in NaN")
	}
	if math.IsInf(result, 0) {
		return nil, fmt.Errorf("operation resulted in infinity")
	}

	// Prepare response
	response := map[string]interface{}{
		"operation": operation,
		"a":         a,
		"result":    result,
	}

	if hasBinary {
		response["b"] = b
	}

	return response, nil
}
