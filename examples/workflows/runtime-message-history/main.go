package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ratlabs-io/go-agent-kit/examples/integrations/openai"
	"github.com/ratlabs-io/go-agent-kit/pkg/agent"
	"github.com/ratlabs-io/go-agent-kit/pkg/constants"
	"github.com/ratlabs-io/go-agent-kit/pkg/llm"
	"github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

// RuntimeMessageHistoryExample demonstrates how to load message history at runtime
func main() {
	fmt.Println("=== Runtime Message History Example ===")
	
	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable not set")
		fmt.Println("Please set it with: export OPENAI_API_KEY=your-api-key-here")
		os.Exit(1)
	}
	
	// Create OpenAI client
	llmClient := openai.NewClient(apiKey)
	
	// Example 1: Runtime history with ChatAgent
	fmt.Println("\n--- Example 1: Runtime History with ChatAgent ---")
	runRuntimeChatHistory(llmClient)
	
	// Example 2: Simulating a conversation session
	fmt.Println("\n--- Example 2: Conversation Session Simulation ---")
	runConversationSession(llmClient)
	
	// Example 3: Dynamic history updates
	fmt.Println("\n--- Example 3: Dynamic History Updates ---")
	runDynamicHistoryUpdates(llmClient)
}

func runRuntimeChatHistory(llmClient llm.Client) {
	// Create agent without any compile-time history
	chatAgent := agent.NewChatAgent("assistant").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a helpful travel assistant.").
		WithClient(llmClient)
	
	// Simulate loading conversation history from database/storage at runtime
	conversationHistory := []llm.Message{
		{Role: "user", Content: "I'm planning a trip to Italy."},
		{Role: "assistant", Content: "Italy is a wonderful destination! What regions or cities are you most interested in visiting?"},
		{Role: "user", Content: "I'm thinking Rome and Florence."},
		{Role: "assistant", Content: "Excellent choices! Rome has amazing historical sites like the Colosseum and Vatican, while Florence is perfect for Renaissance art and architecture."},
	}
	
	// Create workflow context and set history at runtime
	ctx := workflow.NewWorkContext(context.Background())
	ctx.Set(constants.KeyMessageHistory, conversationHistory) // Runtime history loading
	ctx.Set(constants.KeyUserInput, "How many days should I spend in each city?")
	
	fmt.Println("Loaded conversation history at runtime:")
	for _, msg := range conversationHistory {
		fmt.Printf("  %s: %s\n", msg.Role, msg.Content)
	}
	fmt.Println("\nContinuing conversation...")
	fmt.Println("User: How many days should I spend in each city?")
	
	// Run agent - it will use the runtime-loaded history
	report := chatAgent.Run(ctx)
	
	if report.Status == workflow.StatusCompleted {
		fmt.Printf("✅ Agent completed successfully!\n")
		if response, ok := report.Data.(*llm.CompletionResponse); ok {
			fmt.Printf("Assistant: %s\n", response.Content)
		}
	} else {
		fmt.Printf("❌ Agent failed: %v\n", report.Errors)
	}
}

func runConversationSession(llmClient llm.Client) {
	// Simulate a conversation session where history is loaded from external source
	fmt.Println("Simulating a conversation session with external history management...")
	
	// Create agent
	chatAgent := agent.NewChatAgent("cooking-helper").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a helpful cooking assistant.").
		WithClient(llmClient)
	
	// Simulate conversation session with evolving history
	sessionHistory := []llm.Message{}
	
	// First interaction
	fmt.Println("\n--- Turn 1 ---")
	ctx1 := workflow.NewWorkContext(context.Background())
	ctx1.Set(constants.KeyMessageHistory, sessionHistory)
	ctx1.Set(constants.KeyUserInput, "What's a good recipe for beginners?")
	
	fmt.Println("User: What's a good recipe for beginners?")
	report1 := chatAgent.Run(ctx1)
	
	var response1 string
	if report1.Status == workflow.StatusCompleted {
		if resp, ok := report1.Data.(*llm.CompletionResponse); ok {
			response1 = resp.Content
			fmt.Printf("Assistant: %s\n", response1)
			
			// Update session history
			sessionHistory = append(sessionHistory,
				llm.Message{Role: "user", Content: "What's a good recipe for beginners?"},
				llm.Message{Role: "assistant", Content: response1},
			)
		}
	}
	
	// Second interaction with updated history
	fmt.Println("\n--- Turn 2 ---")
	ctx2 := workflow.NewWorkContext(context.Background())
	ctx2.Set(constants.KeyMessageHistory, sessionHistory) // Load updated history
	ctx2.Set(constants.KeyUserInput, "Can you give me the ingredients list?")
	
	fmt.Println("User: Can you give me the ingredients list?")
	report2 := chatAgent.Run(ctx2)
	
	if report2.Status == workflow.StatusCompleted {
		if resp, ok := report2.Data.(*llm.CompletionResponse); ok {
			fmt.Printf("Assistant: %s\n", resp.Content)
		}
	}
}

func runDynamicHistoryUpdates(llmClient llm.Client) {
	fmt.Println("Demonstrating dynamic history updates across multiple agent calls...")
	
	// Create agent
	plannerAgent := agent.NewChatAgent("trip-planner").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a helpful trip planning assistant.").
		WithClient(llmClient)
	
	// Initial conversation context
	history := []llm.Message{
		{Role: "user", Content: "I want to plan a weekend trip."},
		{Role: "assistant", Content: "I'd be happy to help you plan a weekend trip! Where are you thinking of going, and what kind of activities do you enjoy?"},
	}
	
	// Multiple interactions with dynamic history
	interactions := []string{
		"I'm interested in outdoor activities and I'm located in California.",
		"What about Yosemite? Is it good for a weekend?",
		"What should I pack for a Yosemite weekend trip?",
	}
	
	for i, userInput := range interactions {
		fmt.Printf("\n--- Interaction %d ---\n", i+1)
		
		// Create context with current history
		ctx := workflow.NewWorkContext(context.Background())
		ctx.Set(constants.KeyMessageHistory, history)
		ctx.Set(constants.KeyUserInput, userInput)
		
		fmt.Printf("User: %s\n", userInput)
		
		// Run agent
		report := plannerAgent.Run(ctx)
		
		if report.Status == workflow.StatusCompleted {
			if response, ok := report.Data.(*llm.CompletionResponse); ok {
				fmt.Printf("Assistant: %s\n", response.Content)
				
				// Dynamically update history for next interaction
				history = append(history,
					llm.Message{Role: "user", Content: userInput},
					llm.Message{Role: "assistant", Content: response.Content},
				)
				
				fmt.Printf("(History now contains %d messages)\n", len(history))
			}
		} else {
			fmt.Printf("❌ Agent failed: %v\n", report.Errors)
			break
		}
	}
	
	fmt.Printf("\n✅ Dynamic conversation completed with %d total messages in history!\n", len(history))
}