package tools

import (
	"context"
	"testing"
)

// MockTool implements the Tool interface for testing
type MockTool struct {
	name        string
	description string
	schema      Schema
	result      interface{}
	shouldErr   bool
}

func (mt *MockTool) Name() string {
	return mt.name
}

func (mt *MockTool) Description() string {
	return mt.description
}

func (mt *MockTool) Parameters() Schema {
	return mt.schema
}

func (mt *MockTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	if mt.shouldErr {
		return nil, NewToolError("mock error", mt.name)
	}
	return mt.result, nil
}

// MockSimpleTool implements the SimpleTool interface for testing
type MockSimpleTool struct {
	name        string
	description string
	result      interface{}
	shouldErr   bool
}

func (mst *MockSimpleTool) Name() string {
	return mst.name
}

func (mst *MockSimpleTool) Description() string {
	return mst.description
}

func (mst *MockSimpleTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	if mst.shouldErr {
		return nil, NewToolError("simple mock error", mst.name)
	}
	return mst.result, nil
}

func TestDefaultToolRegistry_Register(t *testing.T) {
	registry := NewDefaultToolRegistry()

	tool := &MockTool{
		name:        "test-tool",
		description: "A test tool",
		schema:      Schema{Type: "object"},
		result:      "test-result",
	}

	// Test successful registration
	err := registry.Register(tool)
	if err != nil {
		t.Errorf("Failed to register tool: %v", err)
	}

	// Test retrieval
	retrievedTool, exists := registry.Get("test-tool")
	if !exists {
		t.Error("Tool not found after registration")
	}
	if retrievedTool.Name() != "test-tool" {
		t.Errorf("Retrieved tool has wrong name: %s", retrievedTool.Name())
	}

	// Test duplicate registration
	err = registry.Register(tool)
	if err == nil {
		t.Error("Expected error for duplicate registration")
	}
}

func TestDefaultToolRegistry_RegisterSimple(t *testing.T) {
	registry := NewDefaultToolRegistry()

	simpleTool := &MockSimpleTool{
		name:        "simple-tool",
		description: "A simple test tool",
		result:      "simple-result",
	}

	// Test simple tool registration
	err := registry.RegisterSimple(simpleTool)
	if err != nil {
		t.Errorf("Failed to register simple tool: %v", err)
	}

	// Test retrieval (should be wrapped)
	retrievedTool, exists := registry.Get("simple-tool")
	if !exists {
		t.Error("Simple tool not found after registration")
	}
	if retrievedTool.Name() != "simple-tool" {
		t.Errorf("Retrieved simple tool has wrong name: %s", retrievedTool.Name())
	}

	// Verify it implements full Tool interface
	schema := retrievedTool.Parameters()
	if schema.Type != "object" {
		t.Errorf("Simple tool wrapper should have object schema, got %s", schema.Type)
	}
}

func TestDefaultToolRegistry_List(t *testing.T) {
	registry := NewDefaultToolRegistry()

	// Add multiple tools
	tool1 := &MockTool{name: "tool1", description: "Tool 1", schema: Schema{Type: "object"}}
	tool2 := &MockTool{name: "tool2", description: "Tool 2", schema: Schema{Type: "object"}}
	simpleTool := &MockSimpleTool{name: "simple", description: "Simple tool"}

	if err := registry.Register(tool1); err != nil {
		t.Fatalf("Failed to register tool1: %v", err)
	}
	if err := registry.Register(tool2); err != nil {
		t.Fatalf("Failed to register tool2: %v", err)
	}
	if err := registry.RegisterSimple(simpleTool); err != nil {
		t.Fatalf("Failed to register simple tool: %v", err)
	}

	// Test list
	tools := registry.List()
	if len(tools) != 3 {
		t.Errorf("Expected 3 tools, got %d", len(tools))
	}

	// Verify all tools are present
	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.Name()] = true
	}

	if !names["tool1"] || !names["tool2"] || !names["simple"] {
		t.Error("Not all registered tools found in list")
	}
}

func TestDefaultToolRegistry_Unregister(t *testing.T) {
	registry := NewDefaultToolRegistry()

	tool := &MockTool{name: "temp-tool", description: "Temporary tool", schema: Schema{Type: "object"}}

	// Register and verify
	if err := registry.Register(tool); err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}
	if _, exists := registry.Get("temp-tool"); !exists {
		t.Error("Tool not found after registration")
	}

	// Unregister and verify removal
	err := registry.Unregister("temp-tool")
	if err != nil {
		t.Errorf("Failed to unregister tool: %v", err)
	}

	if _, exists := registry.Get("temp-tool"); exists {
		t.Error("Tool still found after unregistration")
	}

	// Test unregistering non-existent tool
	err = registry.Unregister("non-existent")
	if err == nil {
		t.Error("Expected error for unregistering non-existent tool")
	}
}

func TestDefaultToolRegistry_Count(t *testing.T) {
	registry := NewDefaultToolRegistry()

	// Test empty registry
	if registry.Count() != 0 {
		t.Errorf("Expected count 0, got %d", registry.Count())
	}

	// Add tools and test count
	tool1 := &MockTool{name: "tool1", description: "Tool 1", schema: Schema{Type: "object"}}
	tool2 := &MockTool{name: "tool2", description: "Tool 2", schema: Schema{Type: "object"}}

	if err := registry.Register(tool1); err != nil {
		t.Fatalf("Failed to register tool1: %v", err)
	}
	if registry.Count() != 1 {
		t.Errorf("Expected count 1, got %d", registry.Count())
	}

	if err := registry.Register(tool2); err != nil {
		t.Fatalf("Failed to register tool2: %v", err)
	}
	if registry.Count() != 2 {
		t.Errorf("Expected count 2, got %d", registry.Count())
	}

	if err := registry.Unregister("tool1"); err != nil {
		t.Fatalf("Failed to unregister tool1: %v", err)
	}
	if registry.Count() != 1 {
		t.Errorf("Expected count 1 after unregister, got %d", registry.Count())
	}
}

func TestWrapSimpleTool(t *testing.T) {
	simpleTool := &MockSimpleTool{
		name:        "wrapped-tool",
		description: "A tool to be wrapped",
		result:      "wrapped-result",
	}

	// Wrap the simple tool
	wrappedTool := WrapSimpleTool(simpleTool)

	// Test that it implements Tool interface
	if wrappedTool.Name() != "wrapped-tool" {
		t.Errorf("Wrapped tool name incorrect: %s", wrappedTool.Name())
	}

	if wrappedTool.Description() != "A tool to be wrapped" {
		t.Errorf("Wrapped tool description incorrect: %s", wrappedTool.Description())
	}

	// Test schema generation
	schema := wrappedTool.Parameters()
	if schema.Type != "object" {
		t.Errorf("Wrapped tool should have object schema, got %s", schema.Type)
	}

	// Test execution
	ctx := context.Background()
	result, err := wrappedTool.Execute(ctx, map[string]interface{}{})
	if err != nil {
		t.Errorf("Wrapped tool execution failed: %v", err)
	}
	if result != "wrapped-result" {
		t.Errorf("Wrapped tool result incorrect: %v", result)
	}
}

// Helper function to create tool errors
func NewToolError(message, toolName string) error {
	return &ToolError{Message: message, Tool: toolName}
}

// ToolError represents a tool execution error
type ToolError struct {
	Message string
	Tool    string
}

func (te *ToolError) Error() string {
	return "tool " + te.Tool + ": " + te.Message
}
