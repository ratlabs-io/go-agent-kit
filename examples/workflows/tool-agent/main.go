package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ratlabs-io/go-agent-kit/examples/integrations/openai"
	builtin "github.com/ratlabs-io/go-agent-kit/examples/tools"
	"github.com/ratlabs-io/go-agent-kit/pkg/agent"
	"github.com/ratlabs-io/go-agent-kit/pkg/llm"
	"github.com/ratlabs-io/go-agent-kit/pkg/tools"
	"github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

// ToolAgentWorkflow demonstrates an agent using tools to complete tasks.
func main() {
	fmt.Println("=== Tool Agent Workflow Example ===")

	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable not set")
		fmt.Println("Please set it with: export OPENAI_API_KEY=your-api-key-here")
		os.Exit(1)
	}

	// Create OpenAI client
	llmClient := openai.NewClient(apiKey)

	// Create tools
	echoTool := builtin.NewEchoTool()
	mathTool := builtin.NewMathTool()
	simpleTool := builtin.NewSimpleEchoTool()

	// Create tool agent with multiple tools
	toolAgent := agent.NewToolAgent("calculator").
		WithModel("gpt-3.5-turbo").
		WithPrompt("You are a helpful assistant with access to tools. Use them to complete tasks.").
		WithClient(llmClient).
		WithTools(echoTool, mathTool, tools.WrapSimpleTool(simpleTool))

	// Create workflow context
	ctx := context.Background()
	workflowCtx := workflow.NewWorkContext(ctx)
	workflowCtx.Set("user_input", "Calculate 15 * 23 and echo the result")

	// Run the tool agent
	fmt.Println("Task: Calculate 15 * 23 and echo the result")
	fmt.Println("Running tool agent...")
	fmt.Println()

	report := toolAgent.Run(workflowCtx)

	// Display results
	if report.Status == workflow.StatusCompleted {
		fmt.Println("✅ Tool agent completed successfully!")

		// Show the final result more clearly
		if response, ok := report.Data.(*llm.CompletionResponse); ok {
			if response.Content != "" {
				fmt.Println("\n🤖 Agent Response:")
				fmt.Println("==================")
				fmt.Println(response.Content)
			}

			// Show tool results if available
			if response.Metadata != nil {
				if toolResults, ok := response.Metadata["tool_results"]; ok {
					fmt.Println("\n🔧 Tool Results:")
					fmt.Println("================")
					for _, result := range toolResults.(map[string]interface{}) {
						if resultMap, ok := result.(map[string]interface{}); ok {
							if operation, hasOp := resultMap["operation"]; hasOp {
								if a, hasA := resultMap["a"]; hasA {
									if b, hasB := resultMap["b"]; hasB {
										if finalResult, hasResult := resultMap["result"]; hasResult {
											fmt.Printf("Math: %v %v %v = %v\n", a, operation, b, finalResult)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	} else {
		fmt.Printf("❌ Tool agent failed: %v\n", report.Errors)
	}
}
