package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ratlabs-io/go-agent-kit/examples/integrations/openai"
	"github.com/ratlabs-io/go-agent-kit/pkg/agent"
	"github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

func main() {
	fmt.Println("=== Conditional Workflow Example ===")

	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable not set")
		fmt.Println("Please set it with: export OPENAI_API_KEY=your-api-key-here")
		os.Exit(1)
	}

	llmClient := openai.NewClient(apiKey)

	// Create a classifier agent to determine content type
	classifier := agent.NewChatAgent("classifier").
		WithModel("gpt-3.5-turbo").
		WithPrompt("Classify this input as 'technical', 'creative', or 'general'. Respond with only one word.").
		WithClient(llmClient)

	// Create specialized response agents
	technicalAgent := agent.NewChatAgent("technical-expert").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a technical expert. Provide detailed, accurate technical information.").
		WithClient(llmClient)

	creativeAgent := agent.NewChatAgent("creative-writer").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a creative writer. Respond with imaginative and engaging content.").
		WithClient(llmClient)

	generalAgent := agent.NewChatAgent("general-assistant").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a helpful general assistant. Provide clear, helpful responses.").
		WithClient(llmClient)

	// Create predicate functions
	isTechnical := func(ctx workflow.WorkContext) (bool, error) {
		if classification, ok := ctx.Get("previous_output"); ok {
			return strings.Contains(strings.ToLower(fmt.Sprintf("%v", classification)), "technical"), nil
		}
		return false, nil
	}

	isCreative := func(ctx workflow.WorkContext) (bool, error) {
		if classification, ok := ctx.Get("previous_output"); ok {
			return strings.Contains(strings.ToLower(fmt.Sprintf("%v", classification)), "creative"), nil
		}
		return false, nil
	}

	// Create conditional workflows
	technicalCheck := workflow.NewConditionalFlow("technical-check", isTechnical, technicalAgent, nil)
	creativeCheck := workflow.NewConditionalFlow("creative-check", isCreative, creativeAgent, nil)
	generalFallback := workflow.NewConditionalFlow("general-fallback",
		func(ctx workflow.WorkContext) (bool, error) { return true, nil }, // Always true as fallback
		generalAgent, nil)

	// Create the main workflow
	classifyAndRespond := workflow.NewSequentialFlow("classify-and-respond").
		Then(classifier).
		Then(technicalCheck).
		Then(creativeCheck).
		Then(generalFallback)

	// Test different types of inputs
	testInputs := []string{
		"How does machine learning work?",
		"Write a story about a robot who dreams",
		"What's the weather like today?",
	}

	for i, input := range testInputs {
		fmt.Printf("\n--- Test %d ---\n", i+1)
		fmt.Printf("Input: %s\n", input)

		ctx := context.Background()
		workflowCtx := workflow.NewWorkContext(ctx)
		workflowCtx.Set("user_input", input)

		report := classifyAndRespond.Run(workflowCtx)

		if report.Status == workflow.StatusCompleted {
			if finalOutput, ok := workflowCtx.Get("previous_output"); ok {
				fmt.Printf("Response: %s\n", finalOutput)
			}
		} else {
			fmt.Printf("Failed: %v\n", report.Errors)
		}
	}
}
