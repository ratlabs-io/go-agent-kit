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
	fmt.Println("=== Switch Workflow Example ===")

	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable not set")
		fmt.Println("Please set it with: export OPENAI_API_KEY=your-api-key-here")
		os.Exit(1)
	}

	llmClient := openai.NewClient(apiKey)

	// Create a sentiment analyzer
	sentimentAnalyzer := agent.NewChatAgent("sentiment-analyzer").
		WithModel("gpt-3.5-turbo").
		WithPrompt("Analyze the sentiment of this text. Respond with only one word: 'positive', 'negative', or 'neutral'.").
		WithClient(llmClient)

	// Create specialized response agents for different sentiments
	positiveAgent := agent.NewChatAgent("positive-responder").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are an enthusiastic assistant. Respond with energy and optimism!").
		WithClient(llmClient)

	negativeAgent := agent.NewChatAgent("empathetic-responder").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are an empathetic assistant. Respond with understanding and support.").
		WithClient(llmClient)

	neutralAgent := agent.NewChatAgent("professional-responder").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a professional assistant. Respond with clear, factual information.").
		WithClient(llmClient)

	urgentAgent := agent.NewChatAgent("urgent-responder").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are an urgent response assistant. Provide immediate, actionable help.").
		WithClient(llmClient)

	// Create predicates for different sentiment categories
	isPositive := func(ctx workflow.WorkContext) (bool, error) {
		if sentiment, ok := ctx.Get("previous_output"); ok {
			return strings.Contains(strings.ToLower(fmt.Sprintf("%v", sentiment)), "positive"), nil
		}
		return false, nil
	}

	isNegative := func(ctx workflow.WorkContext) (bool, error) {
		if sentiment, ok := ctx.Get("previous_output"); ok {
			return strings.Contains(strings.ToLower(fmt.Sprintf("%v", sentiment)), "negative"), nil
		}
		return false, nil
	}

	isUrgent := func(ctx workflow.WorkContext) (bool, error) {
		if input, ok := ctx.Get("user_input"); ok {
			inputStr := strings.ToLower(fmt.Sprintf("%v", input))
			urgentWords := []string{"urgent", "emergency", "help", "asap", "immediately", "crisis"}
			for _, word := range urgentWords {
				if strings.Contains(inputStr, word) {
					return true, nil
				}
			}
		}
		return false, nil
	}

	isNeutral := func(ctx workflow.WorkContext) (bool, error) {
		if sentiment, ok := ctx.Get("previous_output"); ok {
			return strings.Contains(strings.ToLower(fmt.Sprintf("%v", sentiment)), "neutral"), nil
		}
		return false, nil
	}

	// Create switch workflow using builder pattern
	sentimentRouter := workflow.NewSwitchFlowBuilder("sentiment-router").
		Case(isUrgent, urgentAgent).     // Check urgency first
		Case(isPositive, positiveAgent). // Then check positive sentiment
		Case(isNegative, negativeAgent). // Then check negative sentiment
		Case(isNeutral, neutralAgent).   // Then check neutral sentiment
		Default(neutralAgent).           // Default to neutral if nothing matches
		Build()

	// Create main workflow: analyze sentiment then route response
	sentimentWorkflow := workflow.NewSequentialFlow("sentiment-workflow").
		Then(sentimentAnalyzer).
		Then(sentimentRouter)

	// Test different sentiment inputs
	testInputs := []string{
		"I'm so excited about this new project!",
		"I'm really struggling with this problem and feeling frustrated.",
		"Can you explain how databases work?",
		"URGENT: I need help immediately with a server outage!",
		"This is terrible, nothing is working right.",
	}

	for i, input := range testInputs {
		fmt.Printf("\n--- Test %d ---\n", i+1)
		fmt.Printf("Input: %s\n", input)

		ctx := context.Background()
		workflowCtx := workflow.NewWorkContext(ctx)
		workflowCtx.Set("user_input", input)

		report := sentimentWorkflow.Run(workflowCtx)

		if report.Status == workflow.StatusCompleted {
			if finalOutput, ok := workflowCtx.Get("previous_output"); ok {
				fmt.Printf("Response: %s\n", finalOutput)
			}
		} else {
			fmt.Printf("Failed: %v\n", report.Errors)
		}
	}
}
