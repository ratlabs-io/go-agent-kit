package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ratlabs-io/go-agent-kit/examples/integrations/clients"
	"github.com/ratlabs-io/go-agent-kit/pkg/agent"
	"github.com/ratlabs-io/go-agent-kit/pkg/llm"
	"github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

func main() {
	// Set up the router once
	router := llm.NewRouterClient()

	// Register available clients based on environment variables
	var registeredProviders []string

	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		openaiClient := clients.NewOpenAIClient(apiKey)
		router.Register("openai", openaiClient)
		registeredProviders = append(registeredProviders, "openai")
	}

	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		anthropicClient := clients.NewAnthropicClient(apiKey)
		router.Register("anthropic", anthropicClient)
		registeredProviders = append(registeredProviders, "anthropic")
	}

	if apiKey := os.Getenv("GROK_API_KEY"); apiKey != "" {
		grokClient := clients.NewGrokClient(apiKey)
		router.Register("grok", grokClient)
		registeredProviders = append(registeredProviders, "grok")
	}

	if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
		geminiClient := clients.NewGeminiClient(apiKey)
		router.Register("gemini", geminiClient)
		registeredProviders = append(registeredProviders, "gemini")
	}

	if len(registeredProviders) == 0 {
		fmt.Println("No API keys found. Set one or more of:")
		fmt.Println("  OPENAI_API_KEY")
		fmt.Println("  ANTHROPIC_API_KEY")
		fmt.Println("  GROK_API_KEY")
		fmt.Println("  GEMINI_API_KEY")
		return
	}

	fmt.Printf("Registered providers: %v\n", registeredProviders)

	// Example 1: Direct router usage
	fmt.Println("\n=== Direct Router Usage ===")
	testDirectUsage(router, registeredProviders)

	// Example 2: Agent with router
	fmt.Println("\n=== Agent with Router ===")
	testAgentUsage(router, registeredProviders)
}

func testDirectUsage(router *llm.RouterClient, providers []string) {
	ctx := context.Background()

	// Test each registered provider
	for _, provider := range providers {
		var model string
		switch provider {
		case "openai":
			model = "openai/gpt-4o-mini"
		case "anthropic":
			model = "anthropic/claude-3-haiku-20240307"
		case "grok":
			model = "grok/grok-beta"
		case "gemini":
			model = "gemini/gemini-1.5-flash"
		}

		fmt.Printf("Testing %s with model: %s\n", provider, model)
		
		response, err := router.Complete(ctx, llm.CompletionRequest{
			Model: model,
			Messages: []llm.Message{
				{Role: "user", Content: "Say hello in a friendly way"},
			},
		})
		
		if err != nil {
			fmt.Printf("Error with %s: %v\n", provider, err)
			continue
		}
		
		fmt.Printf("Response from %s: %s\n", provider, response.Content)
		fmt.Printf("Tokens used: %d total (%d prompt + %d completion)\n\n",
			response.Usage.TotalTokens,
			response.Usage.PromptTokens,
			response.Usage.CompletionTokens)
	}
}

func testAgentUsage(router *llm.RouterClient, providers []string) {
	if len(providers) == 0 {
		return
	}

	// Select the first available provider for the agent example
	var model string
	switch providers[0] {
	case "openai":
		model = "openai/gpt-4o-mini"
	case "anthropic":
		model = "anthropic/claude-3-haiku-20240307"
	case "grok":
		model = "grok/grok-beta"
	case "gemini":
		model = "gemini/gemini-1.5-flash"
	}

	fmt.Printf("Creating chat agent with model: %s\n", model)

	// Create a chat agent using the router
	chatAgent := agent.NewChatAgent("assistant").
		WithModel(model).
		WithClient(router).
		WithPrompt("You are a helpful assistant that explains things clearly and concisely.")

	// Create workflow context
	ctx := context.Background()
	workflowCtx := workflow.NewWorkContext(ctx)
	workflowCtx.Set("user_input", "Explain what a router pattern is in software architecture")

	// Run the agent directly
	fmt.Println("Question: Explain what a router pattern is in software architecture")
	fmt.Println("Running chat agent with router...")
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