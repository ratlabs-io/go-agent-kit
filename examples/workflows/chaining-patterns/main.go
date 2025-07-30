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
	fmt.Println("=== Chaining Patterns Example ===")

	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable not set")
		os.Exit(1)
	}

	llmClient := openai.NewClient(apiKey)

	// Create agents
	researcher := agent.NewChatAgent("researcher").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a researcher. Provide facts about the topic.").
		WithClient(llmClient)

	summarizer := agent.NewChatAgent("summarizer").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a summarizer. Make it concise.").
		WithClient(llmClient)

	reviewer := agent.NewChatAgent("reviewer").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a reviewer. Give feedback on the content.").
		WithClient(llmClient)

	// Pattern 1: ThenChain - each agent gets only the previous output
	fmt.Println("🔗 Pattern 1: ThenChain (output → input)")
	chainFlow := workflow.NewSequentialFlow("chain-pattern").
		Then(researcher).
		ThenChain(summarizer)

	ctx := context.Background()
	chainCtx := workflow.NewWorkContext(ctx)
	chainCtx.Set("user_input", "machine learning")

	chainReport := chainFlow.Run(chainCtx)
	if chainReport.Status == workflow.StatusCompleted {
		if final, ok := chainCtx.Get("previous_output"); ok {
			fmt.Printf("Final: %s\n\n", final)
		}
	}

	// Pattern 2: ThenAccumulate - each agent gets original + all previous outputs
	fmt.Println("📚 Pattern 2: ThenAccumulate (snowball effect)")
	accumFlow := workflow.NewSequentialFlow("accumulate-pattern").
		Then(researcher).
		ThenAccumulate(summarizer).
		ThenAccumulate(reviewer)

	accumCtx := workflow.NewWorkContext(ctx)
	accumCtx.Set("user_input", "machine learning")

	accumReport := accumFlow.Run(accumCtx)
	if accumReport.Status == workflow.StatusCompleted {
		if final, ok := accumCtx.Get("previous_output"); ok {
			fmt.Printf("Final: %s\n", final)
		}
	}
}
