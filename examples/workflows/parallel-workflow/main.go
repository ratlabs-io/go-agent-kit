package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ratlabs-io/go-agent-kit/examples/integrations/clients"
	"github.com/ratlabs-io/go-agent-kit/pkg/agent"
	"github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

func main() {
	fmt.Println("=== Parallel Workflow Example ===")

	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable not set")
		os.Exit(1)
	}

	// Create OpenAI client
	llmClient := clients.NewOpenAIClient(apiKey)

	// Create multiple specialized agents
	techAnalyst := agent.NewChatAgent("tech-analyst").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a technology analyst. Analyze the technical aspects of the topic.").
		WithClient(llmClient)

	marketAnalyst := agent.NewChatAgent("market-analyst").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a market analyst. Analyze the business and market implications.").
		WithClient(llmClient)

	riskAnalyst := agent.NewChatAgent("risk-analyst").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a risk analyst. Identify potential risks and challenges.").
		WithClient(llmClient)

	// Create synthesis agent to combine results
	synthesisAgent := agent.NewChatAgent("synthesizer").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a synthesizer. Combine all the previous analysis into a comprehensive summary.").
		WithClient(llmClient)

	// Create parallel workflow
	parallelFlow := workflow.NewParallelFlow("multi-perspective-analysis").
		Execute(techAnalyst).
		Execute(marketAnalyst).
		Execute(riskAnalyst)

	// Combine parallel + synthesis: parallel analysis → synthesis
	overallFlow := workflow.NewSequentialFlow("comprehensive-analysis").
		Then(parallelFlow).
		ThenChain(synthesisAgent)

	// Run workflow
	ctx := context.Background()
	workflowCtx := workflow.NewWorkContext(ctx)
	workflowCtx.Set("user_input", "implementing AI in small businesses")

	fmt.Println("Topic: implementing AI in small businesses")
	fmt.Println("Running: Tech Analysis || Market Analysis || Risk Analysis → Synthesis")
	fmt.Println()

	report := overallFlow.Run(workflowCtx)

	// Display results
	if report.Status == workflow.StatusCompleted {
		fmt.Println("✅ Parallel workflow completed successfully!")

		// Show final synthesis
		if finalOutput, ok := workflowCtx.Get("previous_output"); ok {
			fmt.Println("\n🎯 Comprehensive Analysis:")
			fmt.Println("=========================")
			fmt.Println(finalOutput)
		}
	} else {
		fmt.Printf("❌ Workflow failed: %v\n", report.Errors)
	}
}
