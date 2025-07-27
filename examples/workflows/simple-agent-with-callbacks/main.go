package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ratlabs-io/go-agent-kit/examples/integrations/openai"
	"github.com/ratlabs-io/go-agent-kit/pkg/agent"
	"github.com/ratlabs-io/go-agent-kit/pkg/llm"
	"github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

// SimpleAgentWithCallbacks demonstrates a basic workflow with callbacks for monitoring.
func main() {
	fmt.Println("=== Simple Agent with Callbacks Example ===")
	
	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable not set")
		fmt.Println("Please set it with: export OPENAI_API_KEY=your-api-key-here")
		os.Exit(1)
	}
	
	// Create OpenAI client
	llmClient := openai.NewClient(apiKey)
	
	// Create callback registry for monitoring
	callbacks := workflow.NewCallbackRegistry()
	
	// Register callback for agent completed events
	callbacks.Add(func(ctx context.Context, event workflow.Event) {
		if event.Type == workflow.EventAgentCompleted {
			fmt.Printf("📊 Event: Agent '%s' completed\n", event.Source)
			if metadata, ok := event.Metadata["elapsed"]; ok {
				fmt.Printf("⏱️  Execution time: %v\n", metadata)
			}
			if metadata, ok := event.Metadata["agent_type"]; ok {
				fmt.Printf("🤖 Agent type: %v\n", metadata)
			}
		}
	})
	
	// Register callback for token usage monitoring
	callbacks.Add(func(ctx context.Context, event workflow.Event) {
		if event.Type == workflow.EventAgentCompleted {
			if response, ok := event.Payload.(*llm.CompletionResponse); ok {
				fmt.Printf("🎯 Token usage - Prompt: %d, Completion: %d, Total: %d\n", 
					response.Usage.PromptTokens, 
					response.Usage.CompletionTokens, 
					response.Usage.TotalTokens)
			}
		}
	})
	
	// Register callback for general monitoring
	callbacks.Add(func(ctx context.Context, event workflow.Event) {
		timestamp := event.Timestamp.Format("15:04:05")
		fmt.Printf("🔔 [%s] Event: %s from %s\n", timestamp, event.Type, event.Source)
	})
	
	// Create a simple chat agent
	chatAgent := agent.NewChatAgent("assistant").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a helpful assistant. Respond concisely and helpfully.").
		WithClient(llmClient)
	
	// Create AgentContext with callbacks (instead of plain WorkContext)
	ctx := context.Background()
	agentCtx := workflow.NewAgentContext(ctx, callbacks, nil) // nil for tool registry since we don't need tools
	agentCtx.Set("user_input", "What are the three largest cities in Japan?")
	
	// Run the agent
	fmt.Println("🚀 Running chat agent with callback monitoring...")
	fmt.Println("📝 Question: What are the three largest cities in Japan?")
	fmt.Println()
	
	start := time.Now()
	report := chatAgent.Run(agentCtx.WorkContext) // Pass the embedded WorkContext
	elapsed := time.Since(start)
	
	fmt.Println()
	fmt.Printf("⏰ Total execution time: %v\n", elapsed)
	
	// Check results
	if report.Status == workflow.StatusCompleted {
		fmt.Printf("✅ Agent completed successfully!\n")
		if response, ok := report.Data.(*llm.CompletionResponse); ok {
			fmt.Printf("💬 Response: %s\n", response.Content)
		}
		
		// Show metadata from the report
		fmt.Println("\n📋 Report Metadata:")
		for key, value := range report.Metadata {
			fmt.Printf("   %s: %v\n", key, value)
		}
	} else {
		fmt.Printf("❌ Agent failed: %v\n", report.Errors)
	}
	
	fmt.Println("\n🎉 Callback monitoring demonstration complete!")
}