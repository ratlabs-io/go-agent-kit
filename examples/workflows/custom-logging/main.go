package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/ratlabs-io/go-agent-kit/examples/integrations/clients"
	"github.com/ratlabs-io/go-agent-kit/pkg/agent"
	"github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

func main() {
	fmt.Println("=== Custom Logging Example ===")

	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable not set")
		os.Exit(1)
	}

	llmClient := clients.NewOpenAIClient(apiKey)

	// Create agent
	chatAgent := agent.NewChatAgent("assistant").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a helpful assistant.").
		WithClient(llmClient)

	fmt.Println("\n--- Test 1: Default Logging (to stderr) ---")
	ctx1 := workflow.NewWorkContext(context.Background())
	ctx1.Set("user_input", "Hello!")
	report1 := chatAgent.Run(ctx1)
	if report1.Status == workflow.StatusCompleted {
		fmt.Printf("✅ Response: %v\n", report1.Data)
	}

	fmt.Println("\n--- Test 2: No Logging ---")
	// Disable all logging
	workflow.SetDefaultLogger(workflow.NewNoOpLogger())
	ctx2 := workflow.NewWorkContext(context.Background())
	ctx2.Set("user_input", "Hello again!")
	report2 := chatAgent.Run(ctx2)
	if report2.Status == workflow.StatusCompleted {
		fmt.Printf("✅ Response: %v (no logs should appear)\n", report2.Data)
	}

	fmt.Println("\n--- Test 3: Custom Logger ---")
	// Create custom logger that writes to stdout with debug level
	customLogger := workflow.NewSlogLogger(
		slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})).With("app", "go-agent-kit-example"),
	)
	workflow.SetDefaultLogger(customLogger)

	ctx3 := workflow.NewWorkContext(context.Background())
	ctx3.Set("user_input", "Hello with custom logging!")
	report3 := chatAgent.Run(ctx3)
	if report3.Status == workflow.StatusCompleted {
		fmt.Printf("✅ Response: %v\n", report3.Data)
	}

	fmt.Println("\n--- Test 4: Per-Context Logger ---")
	// Create a specific logger for this context only
	contextLogger := workflow.NewSlogLogger(
		slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})).With("context", "special-context"),
	)

	ctx4 := workflow.NewWorkContextWithLogger(context.Background(), contextLogger)
	ctx4.Set("user_input", "Hello with per-context logging!")
	report4 := chatAgent.Run(ctx4)
	if report4.Status == workflow.StatusCompleted {
		fmt.Printf("✅ Response: %v\n", report4.Data)
	}
}
