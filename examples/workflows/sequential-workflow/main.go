package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ratlabs-io/go-agent-kit/examples/integrations/openai"
	"github.com/ratlabs-io/go-agent-kit/pkg/agent"
	"github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

func main() {
	fmt.Println("=== Sequential Workflow Example ===")
	
	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable not set")
		os.Exit(1)
	}
	
	// Create OpenAI client
	llmClient := openai.NewClient(apiKey)
	
	// Create agents
	researcher := agent.NewChatAgent("researcher").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a research assistant. Provide key facts about the topic.").
		WithClient(llmClient)
	
	summarizer := agent.NewChatAgent("summarizer").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a summarizer. Create a concise summary of the provided content.").
		WithClient(llmClient)
	
	analyzer := agent.NewChatAgent("analyzer").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are an analyst. Extract the 3 most important insights.").
		WithClient(llmClient)
	
	// Create workflow - clean and simple!
	pipeline := workflow.NewSequentialFlow("research-pipeline").
		Then(researcher).
		ThenChain(summarizer).
		ThenChain(analyzer)
	
	// Run workflow
	ctx := context.Background()
	workflowCtx := workflow.NewWorkContext(ctx)
	workflowCtx.Set("user_input", "quantum computing in drug discovery")
	
	fmt.Println("Topic: quantum computing in drug discovery")
	fmt.Println("Pipeline: Research → Summarize → Analyze")
	fmt.Println()
	
	report := pipeline.Run(workflowCtx)
	
	// Display results
	if report.Status == workflow.StatusCompleted {
		fmt.Println("✅ Workflow completed successfully!")
		
		// Show final result
		if finalOutput, ok := workflowCtx.Get("previous_output"); ok {
			fmt.Println("\n🎯 Final Analysis:")
			fmt.Println("==================")
			fmt.Println(finalOutput)
		}
	} else {
		fmt.Printf("❌ Workflow failed: %v\n", report.Errors)
	}
}