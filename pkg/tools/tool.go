package tools

import (
	"context"
	"encoding/json"
)

// Schema represents a JSON Schema for tool parameters.
type Schema struct {
	Type        string                 `json:"type"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Description string                 `json:"description,omitempty"`
}

// SimpleTool defines a minimal interface for native tools with automatic schema generation.
// This is the recommended interface for simple native tools.
type SimpleTool interface {
	// Name returns the unique name of the tool.
	Name() string

	// Description returns a human-readable description of what the tool does.
	Description() string

	// Execute runs the tool with the given parameters and returns the result.
	// The params will be automatically validated against the tool's parameter struct.
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// Tool defines the interface for MCP-style tools that can be executed by agents.
// This interface is compatible with the Model Context Protocol and provides
// full control over parameter schemas.
type Tool interface {
	// Name returns the unique name of the tool.
	Name() string

	// Description returns a human-readable description of what the tool does.
	Description() string

	// Parameters returns the JSON Schema for the tool's parameters.
	Parameters() Schema

	// Execute runs the tool with the given parameters and returns the result.
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// ToolRegistry manages the registration and discovery of tools.
type ToolRegistry interface {
	// Register adds a tool to the registry.
	Register(tool Tool) error

	// Get retrieves a tool by name.
	Get(name string) (Tool, bool)

	// List returns all registered tools.
	List() []Tool

	// Unregister removes a tool from the registry.
	Unregister(name string) error
}

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	Data     interface{}            `json:"data"`
	Error    error                  `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for ToolResult.
func (tr *ToolResult) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{
		"data": tr.Data,
	}

	if tr.Error != nil {
		result["error"] = tr.Error.Error()
	}

	if tr.Metadata != nil {
		result["metadata"] = tr.Metadata
	}

	return json.Marshal(result)
}

// SimpleToolAdapter wraps a SimpleTool to implement the full Tool interface
// by automatically generating JSON schemas from struct tags.
type SimpleToolAdapter struct {
	simple SimpleTool
	schema Schema
}

// WrapSimpleTool converts a SimpleTool to a Tool by automatically generating
// the parameter schema. The schema generation can be enhanced in the future
// to use struct tags or reflection.
func WrapSimpleTool(simple SimpleTool) Tool {
	return &SimpleToolAdapter{
		simple: simple,
		schema: Schema{
			Type:        "object",
			Description: "Parameters for " + simple.Name(),
			Properties:  map[string]interface{}{},
			Required:    []string{},
		},
	}
}

// Name returns the name from the wrapped SimpleTool.
func (sta *SimpleToolAdapter) Name() string {
	return sta.simple.Name()
}

// Description returns the description from the wrapped SimpleTool.
func (sta *SimpleToolAdapter) Description() string {
	return sta.simple.Description()
}

// Parameters returns the auto-generated schema.
func (sta *SimpleToolAdapter) Parameters() Schema {
	return sta.schema
}

// Execute delegates to the wrapped SimpleTool.
func (sta *SimpleToolAdapter) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return sta.simple.Execute(ctx, params)
}
