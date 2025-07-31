package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ratlabs-io/go-agent-kit/examples/integrations/clients"
	"github.com/ratlabs-io/go-agent-kit/pkg/agent"
	"github.com/ratlabs-io/go-agent-kit/pkg/llm"
	"github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

// SimpleAgentWorkflow demonstrates a basic workflow with a single chat agent.
func main() {
	fmt.Println("=== Simple Agent Workflow Example ===")

	// Set up the router with available providers
	router := llm.NewRouterClient()
	var hasProvider bool

	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		router.Register("openai", clients.NewOpenAIClient(apiKey))
		hasProvider = true
		fmt.Println("✓ Registered OpenAI provider")
	}
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		router.Register("anthropic", clients.NewAnthropicClient(apiKey))
		hasProvider = true
		fmt.Println("✓ Registered Anthropic provider")
	}
	if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
		router.Register("gemini", clients.NewGeminiClient(apiKey))
		hasProvider = true
		fmt.Println("✓ Registered Gemini provider")
	}
	if apiKey := os.Getenv("GROK_API_KEY"); apiKey != "" {
		router.Register("grok", clients.NewGrokClient(apiKey))
		hasProvider = true
		fmt.Println("✓ Registered Grok provider")
	}

	if !hasProvider {
		fmt.Println("Error: No API keys found. Set one or more of:")
		fmt.Println("  OPENAI_API_KEY")
		fmt.Println("  ANTHROPIC_API_KEY")
		fmt.Println("  GEMINI_API_KEY")
		fmt.Println("  GROK_API_KEY")
		os.Exit(1)
	}

	// Determine which model to use (prefer OpenAI, fall back to others)
	var model string
	if router.IsProviderRegistered("openai") {
		model = "openai/gpt-4o-mini"
	} else if router.IsProviderRegistered("anthropic") {
		model = "anthropic/claude-3-haiku-20240307"
	} else if router.IsProviderRegistered("gemini") {
		model = "gemini/gemini-1.5-flash"
	} else if router.IsProviderRegistered("grok") {
		model = "grok/grok-beta"
	}

	fmt.Printf("Using model: %s\n", model)

	// Example 1: Basic Text Response
	fmt.Println("\n--- Example 1: Basic Text Response ---")
	runBasicTextExample(router, model)

	// Example 2: JSON Response (no schema)
	fmt.Println("\n--- Example 2: JSON Response ---")
	runJSONResponseExample(router, model)
}

func runBasicTextExample(llmClient llm.Client, model string) {
	// Create a simple chat agent
	chatAgent := agent.NewChatAgent("assistant").
		WithModel(model).
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

func runJSONResponseExample(llmClient llm.Client, model string) {
	// Create an agent that responds with JSON
	jsonAgent := agent.NewChatAgent("json-assistant").
		WithModel(model).
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
