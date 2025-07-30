package main

import (
	"context"
	"encoding/json"
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

	// Example 1: Basic Text Response
	fmt.Println("\n--- Example 1: Basic Text Response ---")
	runBasicTextExample(llmClient)

	// Example 2: JSON Response (no schema)
	fmt.Println("\n--- Example 2: JSON Response ---")
	runJSONResponseExample(llmClient)
}

func runBasicTextExample(llmClient llm.Client) {
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
	fmt.Println("Question: What is the capital of France?")
	fmt.Println("Running chat agent...")
	report := chatAgent.Run(workflowCtx)

	// Check results
	if report.Status == workflow.StatusCompleted {
		fmt.Printf("✅ Agent completed successfully!\n")
		if response, ok := report.Data.(*llm.CompletionResponse); ok {
			fmt.Printf("Answer: %s\n", response.Content)
		}
	} else {
		fmt.Printf("❌ Agent failed: %v\n", report.Errors)
	}
}

func runJSONResponseExample(llmClient llm.Client) {
	// Create an agent that responds with JSON
	jsonAgent := agent.NewChatAgent("json-assistant").
		WithModel("gpt-3.5-turbo").
		WithPrompt(`You are a helpful assistant. Respond with a JSON object containing:
- "answer": the direct answer to the question
- "confidence": your confidence level (low/medium/high)  
- "category": the type of question (geography/science/history/other)
- "additional_info": any relevant extra information`).
		WithJSONResponse().
		WithClient(llmClient)

	// Create workflow context
	ctx := context.Background()
	workflowCtx := workflow.NewWorkContext(ctx)
	workflowCtx.Set("user_input", "What is the capital of France?")

	// Run the agent
	fmt.Println("Question: What is the capital of France?")
	fmt.Println("Running JSON chat agent...")
	report := jsonAgent.Run(workflowCtx)

	// Check results
	if report.Status == workflow.StatusCompleted {
		fmt.Printf("✅ Agent completed successfully!\n")
		if response, ok := report.Data.(*llm.CompletionResponse); ok {
			// Pretty print the JSON response
			var jsonResponse map[string]interface{}
			if err := json.Unmarshal([]byte(response.Content), &jsonResponse); err != nil {
				fmt.Printf("Raw JSON: %s\n", response.Content)
			} else {
				fmt.Println("JSON Response:")
				prettyJSON, _ := json.MarshalIndent(jsonResponse, "", "  ")
				fmt.Println(string(prettyJSON))
			}
		}
	} else {
		fmt.Printf("❌ Agent failed: %v\n", report.Errors)
	}
}
