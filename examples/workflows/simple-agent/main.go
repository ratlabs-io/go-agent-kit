package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ratlabs-io/go-agent-kit/examples/integrations/openai"
	"github.com/ratlabs-io/go-agent-kit/pkg/agent"
	"github.com/ratlabs-io/go-agent-kit/pkg/llm"
	"github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

// SimpleAgentWorkflow demonstrates a basic workflow with a single chat agent.
func main() {
	fmt.Println("=== Simple Agent Workflow Example ===")
	
	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable not set")
		fmt.Println("Please set it with: export OPENAI_API_KEY=your-api-key-here")
		os.Exit(1)
	}
	
	// Create OpenAI client
	llmClient := openai.NewClient(apiKey)
	
	// Create a simple chat agent
	chatAgent := agent.NewChatAgent("assistant").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a helpful assistant. Respond concisely.").
		WithClient(llmClient)
	
	// Create workflow context
	ctx := context.Background()
	workflowCtx := workflow.NewWorkContext(ctx)
	workflowCtx.Set("user_input", "What is the capital of France?")
	
	// Run the agent
	fmt.Println("Running chat agent...")
	report := chatAgent.Run(workflowCtx)
	
	// Check results
	if report.Status == workflow.StatusCompleted {
		fmt.Printf("Agent completed successfully!\n")
		if response, ok := report.Data.(*llm.CompletionResponse); ok {
			fmt.Printf("Response: %s\n", response.Content)
		}
	} else {
		fmt.Printf("Agent failed: %v\n", report.Errors)
	}
}