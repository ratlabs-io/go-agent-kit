package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ratlabs-io/go-agent-kit/examples/integrations/openai"
	"github.com/ratlabs-io/go-agent-kit/pkg/agent"
	"github.com/ratlabs-io/go-agent-kit/pkg/llm"
	"github.com/ratlabs-io/go-agent-kit/pkg/tools"
	"github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

// ChatWithHistoryExample demonstrates how to use message history with agents
func main() {
	fmt.Println("=== Chat Agent with Message History Example ===")
	
	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable not set")
		fmt.Println("Please set it with: export OPENAI_API_KEY=your-api-key-here")
		os.Exit(1)
	}
	
	// Create OpenAI client
	llmClient := openai.NewClient(apiKey)
	
	// Example 1: Basic Chat with History
	fmt.Println("\n--- Example 1: Basic Chat with History ---")
	runBasicChatWithHistory(llmClient)
	
	// Example 2: Tool Agent with History
	fmt.Println("\n--- Example 2: Tool Agent with History ---")
	runToolAgentWithHistory(llmClient)
	
	// Example 3: Multi-turn Conversation
	fmt.Println("\n--- Example 3: Multi-turn Conversation ---")
	runMultiTurnConversation(llmClient)
}

func runBasicChatWithHistory(llmClient llm.Client) {
	// Simulate a previous conversation about travel
	messageHistory := []llm.Message{
		{Role: "user", Content: "I'm planning a trip to Japan."},
		{Role: "assistant", Content: "That's exciting! Japan is a wonderful destination. When are you planning to visit, and what aspects of Japanese culture or attractions are you most interested in?"},
		{Role: "user", Content: "I'm thinking about going in April for the cherry blossoms."},
		{Role: "assistant", Content: "April is an excellent choice! The cherry blossom (sakura) season is magical. The blooms typically peak in early to mid-April in popular spots like Tokyo and Kyoto. I'd recommend booking accommodations early as it's a very popular time to visit."},
	}
	
	// Create agent with conversation history
	chatAgent := agent.NewChatAgent("travel-assistant").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a helpful travel assistant with knowledge of previous conversations. Continue helping the user with their travel plans.").
		WithClient(llmClient).
		WithMessageHistory(messageHistory)
	
	// Continue the conversation
	ctx := workflow.NewWorkContext(context.Background())
	ctx.Set("user_input", "What are some must-see places in Kyoto?")
	
	fmt.Println("Previous conversation:")
	for _, msg := range messageHistory {
		fmt.Printf("  %s: %s\n", msg.Role, msg.Content)
	}
	fmt.Println("\nContinuing conversation...")
	fmt.Println("User: What are some must-see places in Kyoto?")
	
	// Run the agent
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

func runToolAgentWithHistory(llmClient llm.Client) {
	// Create a simple math tool
	mathTool := &SimpleMathTool{}
	
	// Previous conversation with tool usage
	toolHistory := []llm.Message{
		{Role: "user", Content: "I need help with some calculations for my budget."},
		{Role: "assistant", Content: "I'd be happy to help you with budget calculations. What specific calculations do you need?"},
		{Role: "user", Content: "My monthly income is $5000. My rent is $1500, utilities are $200, and groceries are $400."},
		{Role: "assistant", Content: "Let me calculate your expenses and remaining budget. Your total fixed expenses are $2100 ($1500 + $200 + $400), leaving you with $2900 per month for other expenses and savings."},
	}
	
	// Create tool agent with history
	toolAgent := agent.NewToolAgent("budget-assistant").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a helpful financial assistant that can perform calculations. You have access to previous conversations about the user's budget.").
		WithClient(llmClient).
		WithTools(mathTool).
		WithMessageHistory(toolHistory)
	
	ctx := workflow.NewWorkContext(context.Background())
	ctx.Set("user_input", "If I save 20% of my remaining budget, how much would that be?")
	
	fmt.Println("Previous conversation:")
	for _, msg := range toolHistory {
		fmt.Printf("  %s: %s\n", msg.Role, msg.Content)
	}
	fmt.Println("\nContinuing conversation...")
	fmt.Println("User: If I save 20% of my remaining budget, how much would that be?")
	
	report := toolAgent.Run(ctx)
	
	if report.Status == workflow.StatusCompleted {
		fmt.Printf("✅ Agent completed successfully!\n")
		if response, ok := report.Data.(*llm.CompletionResponse); ok {
			fmt.Printf("Assistant: %s\n", response.Content)
		}
	} else {
		fmt.Printf("❌ Agent failed: %v\n", report.Errors)
	}
}

func runMultiTurnConversation(llmClient llm.Client) {
	fmt.Println("Starting a multi-turn conversation about cooking...")
	
	// Initialize with system prompt only
	var history []llm.Message
	
	// Create the agent
	cookingAgent := agent.NewChatAgent("cooking-assistant").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a helpful cooking assistant. Help users with recipes, cooking techniques, and meal planning.").
		WithClient(llmClient)
	
	// Simulate multiple turns of conversation
	questions := []string{
		"I want to learn how to make pasta from scratch.",
		"What type of flour should I use?",
		"How long should I knead the dough?",
		"Can I make it without a pasta machine?",
	}
	
	for i, question := range questions {
		fmt.Printf("\nTurn %d - User: %s\n", i+1, question)
		
		// Update agent with current history
		cookingAgent = cookingAgent.WithMessageHistory(history)
		
		// Create context for this turn
		ctx := workflow.NewWorkContext(context.Background())
		ctx.Set("user_input", question)
		
		// Run the agent
		report := cookingAgent.Run(ctx)
		
		if report.Status == workflow.StatusCompleted {
			if response, ok := report.Data.(*llm.CompletionResponse); ok {
				fmt.Printf("Assistant: %s\n", response.Content)
				
				// Add this exchange to history
				history = append(history,
					llm.Message{Role: "user", Content: question},
					llm.Message{Role: "assistant", Content: response.Content},
				)
			}
		} else {
			fmt.Printf("❌ Agent failed: %v\n", report.Errors)
			break
		}
	}
	
	fmt.Printf("\n✅ Completed %d turns of conversation!\n", len(questions))
}

// SimpleMathTool is a basic calculator tool
type SimpleMathTool struct{}

func (t *SimpleMathTool) Name() string {
	return "calculator"
}

func (t *SimpleMathTool) Description() string {
	return "Performs basic math operations: add, subtract, multiply, divide"
}

func (t *SimpleMathTool) Parameters() tools.Schema {
	return tools.Schema{
		Type: "object",
		Properties: map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "The operation to perform: add, subtract, multiply, divide",
				"enum":        []string{"add", "subtract", "multiply", "divide"},
			},
			"a": map[string]interface{}{
				"type":        "number",
				"description": "First number",
			},
			"b": map[string]interface{}{
				"type":        "number",
				"description": "Second number",
			},
		},
		Required: []string{"operation", "a", "b"},
	}
}

func (t *SimpleMathTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	op, _ := params["operation"].(string)
	a, _ := params["a"].(float64)
	b, _ := params["b"].(float64)
	
	var result float64
	switch op {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		result = a / b
	default:
		return nil, fmt.Errorf("unknown operation: %s", op)
	}
	
	return map[string]interface{}{
		"result": result,
		"expression": fmt.Sprintf("%g %s %g = %g", a, op, b, result),
	}, nil
}