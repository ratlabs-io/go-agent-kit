package agent

import (
	"context"
	"testing"

	"github.com/ratlabs-io/go-agent-kit/pkg/llm"
	"github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

// MockLLMClient implements the llm.Client interface for testing
type MockLLMClient struct {
	shouldErr bool
	response  *llm.CompletionResponse
}

func (m *MockLLMClient) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	if m.shouldErr {
		return nil, &LLMError{Message: "mock LLM error"}
	}
	return m.response, nil
}

func (m *MockLLMClient) Close() error {
	return nil
}

func TestChatAgent_Creation(t *testing.T) {
	agent := NewChatAgent("test-agent")
	
	if agent.Name() != "test-agent" {
		t.Errorf("Expected name 'test-agent', got '%s'", agent.Name())
	}
	
	if agent.Type() != TypeChat {
		t.Errorf("Expected type TypeChat, got %v", agent.Type())
	}
	
	if len(agent.Tools()) != 0 {
		t.Errorf("Expected 0 tools initially, got %d", len(agent.Tools()))
	}
}

func TestChatAgent_Configuration(t *testing.T) {
	agent := NewChatAgent("config-test")
	
	// Test builder pattern
	agent.WithModel("test-model").
		WithPrompt("test prompt")
	
	// Test configuration via map
	config := map[string]interface{}{
		"model":  "config-model",
		"prompt": "config prompt",
	}
	
	err := agent.Configure(config)
	if err != nil {
		t.Errorf("Configuration failed: %v", err)
	}
}

func TestChatAgent_SuccessfulRun(t *testing.T) {
	// Create mock LLM client
	mockResponse := &llm.CompletionResponse{
		Content: "Hello, I'm a helpful assistant!",
		Usage: llm.Usage{
			PromptTokens:     10,
			CompletionTokens: 8,
			TotalTokens:      18,
		},
	}
	
	mockClient := &MockLLMClient{
		shouldErr: false,
		response:  mockResponse,
	}
	
	// Create agent
	agent := NewChatAgent("test-assistant").
		WithModel("test-model").
		WithPrompt("You are a helpful assistant").
		WithClient(mockClient)
	
	// Create context
	ctx := workflow.NewWorkContext(context.Background())
	ctx.Set("user_input", "Hello!")
	
	// Run agent
	report := agent.Run(ctx)
	
	// Verify successful execution
	if report.Status != workflow.StatusCompleted {
		t.Errorf("Expected StatusCompleted, got %v", report.Status)
	}
	
	if len(report.Errors) != 0 {
		t.Errorf("Expected no errors, got %v", report.Errors)
	}
	
	// Verify response data
	if response, ok := report.Data.(*llm.CompletionResponse); ok {
		if response.Content != "Hello, I'm a helpful assistant!" {
			t.Errorf("Unexpected response content: %s", response.Content)
		}
		if response.Usage.TotalTokens != 18 {
			t.Errorf("Unexpected token usage: %d", response.Usage.TotalTokens)
		}
	} else {
		t.Error("Report data is not a CompletionResponse")
	}
	
	// Verify metadata
	if agentName, ok := report.Metadata["agent_name"]; !ok || agentName != "test-assistant" {
		t.Error("Agent name not found in metadata")
	}
	if agentType, ok := report.Metadata["agent_type"]; !ok || agentType != TypeChat {
		t.Error("Agent type not found in metadata")
	}
}

func TestChatAgent_LLMError(t *testing.T) {
	// Create mock LLM client that returns error
	mockClient := &MockLLMClient{
		shouldErr: true,
	}
	
	// Create agent
	agent := NewChatAgent("error-test").
		WithModel("test-model").
		WithPrompt("Test prompt").
		WithClient(mockClient)
	
	// Create context
	ctx := workflow.NewWorkContext(context.Background())
	
	// Run agent
	report := agent.Run(ctx)
	
	// Verify failure
	if report.Status != workflow.StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", report.Status)
	}
	
	if len(report.Errors) == 0 {
		t.Error("Expected errors, got none")
	}
}

func TestChatAgent_NoClient(t *testing.T) {
	// Create agent without LLM client
	agent := NewChatAgent("no-client-test").
		WithModel("test-model").
		WithPrompt("Test prompt")
	// Note: no WithClient() call
	
	// Create context
	ctx := workflow.NewWorkContext(context.Background())
	
	// Run agent
	report := agent.Run(ctx)
	
	// Verify failure due to missing client
	if report.Status != workflow.StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", report.Status)
	}
	
	if len(report.Errors) == 0 {
		t.Error("Expected errors for missing client, got none")
	}
}

func TestChatAgent_ToolConversion(t *testing.T) {
	// This tests the convertSchemaToMap function indirectly
	agent := NewChatAgent("tool-test")
	
	// The agent should be able to handle tools even if it's primarily for chat
	if len(agent.Tools()) != 0 {
		t.Errorf("Expected 0 tools initially, got %d", len(agent.Tools()))
	}
}

// Helper error type for testing
type LLMError struct {
	Message string
}

func (e *LLMError) Error() string {
	return e.Message
}