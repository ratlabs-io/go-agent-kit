package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ratlabs-io/go-agent-kit/examples/integrations/openai"
	"github.com/ratlabs-io/go-agent-kit/pkg/agent"
	"github.com/ratlabs-io/go-agent-kit/pkg/llm"
	"github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

// TaskAnalysis represents the structured output we expect
type TaskAnalysis struct {
	Category      string   `json:"category"`
	Complexity    string   `json:"complexity"`
	EstimatedTime int      `json:"estimated_time"`
	Requirements  []string `json:"requirements"`
}

// StructuredJSONAgentWorkflow demonstrates how to get structured JSON responses from agents.
func main() {
	fmt.Println("=== Structured JSON Agent Workflow Example ===")

	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable not set")
		fmt.Println("Please set it with: export OPENAI_API_KEY=your-api-key-here")
		os.Exit(1)
	}

	// Create OpenAI client
	llmClient := openai.NewClient(apiKey)

	// Example 1: Structured JSON with Schema
	fmt.Println("\n--- Example 1: Task Analysis with JSON Schema ---")
	runTaskAnalysisExample(llmClient)

	// Example 2: Generic JSON Response
	fmt.Println("\n--- Example 2: Generic JSON Response ---")
	runGenericJSONExample(llmClient)

	// Example 3: Multiple Analysis Workflow
	fmt.Println("\n--- Example 3: Multi-Step Analysis Workflow ---")
	runMultiStepAnalysisWorkflow(llmClient)
}

func runTaskAnalysisExample(llmClient llm.Client) {
	// Define a JSON schema for task analysis
	schema := &llm.JSONSchema{
		Name:        "task_analysis",
		Description: "Analysis of a user task with structured categorization",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"category": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"question", "request", "task", "planning", "other"},
					"description": "The type of user input",
				},
				"complexity": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"low", "medium", "high"},
					"description": "Estimated complexity level",
				},
				"estimated_time": map[string]interface{}{
					"type":        "integer",
					"description": "Estimated time in minutes",
					"minimum":     1,
					"maximum":     1440,
				},
				"requirements": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "List of requirements or steps needed",
				},
			},
			"required":             []string{"category", "complexity", "estimated_time", "requirements"},
			"additionalProperties": false,
		},
		Strict: true,
	}

	// Create agent with JSON schema
	// Note: JSON schema requires gpt-4o, gpt-4o-mini, or gpt-4-turbo-preview
	// gpt-3.5-turbo only supports json_object, not json_schema
	analysisAgent := agent.NewChatAgent("task-analyzer").
		WithModel("gpt-4o-mini").
		WithPrompt(`You are a task analysis expert. Analyze the user's input and provide a structured breakdown. 
Consider the type of request, complexity level, time requirements, and necessary steps.
Be precise and realistic in your estimates.`).
		WithJSONSchema(schema).
		WithClient(llmClient)

	// Test with different types of inputs
	testInputs := []string{
		"Help me plan a vacation to Japan for 2 weeks",
		"What is the capital of France?",
		"I need to learn Python programming from scratch",
		"Fix the bug in my authentication system that's causing login failures",
	}

	for i, input := range testInputs {
		fmt.Printf("\nInput %d: %s\n", i+1, input)

		ctx := workflow.NewWorkContext(context.Background())
		ctx.Set("user_input", input)

		report := analysisAgent.Run(ctx)

		if report.Status == workflow.StatusCompleted {
			if response, ok := report.Data.(*llm.CompletionResponse); ok {
				// Parse the structured JSON response
				var analysis TaskAnalysis
				if err := json.Unmarshal([]byte(response.Content), &analysis); err != nil {
					fmt.Printf("Error parsing JSON: %v\n", err)
					fmt.Printf("Raw response: %s\n", response.Content)
				} else {
					fmt.Printf("✅ Category: %s\n", analysis.Category)
					fmt.Printf("✅ Complexity: %s\n", analysis.Complexity)
					fmt.Printf("✅ Estimated Time: %d minutes\n", analysis.EstimatedTime)
					fmt.Printf("✅ Requirements: %v\n", analysis.Requirements)
				}
			}
		} else {
			fmt.Printf("❌ Analysis failed: %v\n", report.Errors)
		}
	}
}

func runGenericJSONExample(llmClient llm.Client) {
	// Create agent for generic JSON responses (no specific schema)
	jsonAgent := agent.NewChatAgent("json-responder").
		WithModel("gpt-3.5-turbo").
		WithPrompt(`Respond with a JSON object containing key insights about the user's input. 
Include fields like 'summary', 'key_points', 'recommendations', and any other relevant information.`).
		WithJSONResponse().
		WithClient(llmClient)

	ctx := workflow.NewWorkContext(context.Background())
	ctx.Set("user_input", "I'm starting a new business selling handmade jewelry online")

	fmt.Println("Input: I'm starting a new business selling handmade jewelry online")

	report := jsonAgent.Run(ctx)

	if report.Status == workflow.StatusCompleted {
		if response, ok := report.Data.(*llm.CompletionResponse); ok {
			// Parse the generic JSON response
			var insights map[string]interface{}
			if err := json.Unmarshal([]byte(response.Content), &insights); err != nil {
				fmt.Printf("Error parsing JSON: %v\n", err)
			} else {
				fmt.Println("✅ Business Insights (JSON):")
				// Pretty print the JSON
				prettyJSON, _ := json.MarshalIndent(insights, "", "  ")
				fmt.Println(string(prettyJSON))
			}
		}
	} else {
		fmt.Printf("❌ JSON response failed: %v\n", report.Errors)
	}
}

func runMultiStepAnalysisWorkflow(llmClient llm.Client) {
	// Step 1: Initial analysis with JSON schema
	initialSchema := &llm.JSONSchema{
		Name:        "initial_analysis",
		Description: "Initial breakdown of a complex task",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"main_goal": map[string]interface{}{
					"type":        "string",
					"description": "The primary objective",
				},
				"sub_tasks": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name":        map[string]interface{}{"type": "string"},
							"description": map[string]interface{}{"type": "string"},
							"priority":    map[string]interface{}{"type": "integer", "minimum": 1, "maximum": 5},
						},
						"required":             []string{"name", "description", "priority"},
						"additionalProperties": false,
					},
				},
				"estimated_total_time": map[string]interface{}{
					"type":        "integer",
					"description": "Total estimated time in hours",
				},
			},
			"required":             []string{"main_goal", "sub_tasks", "estimated_total_time"},
			"additionalProperties": false,
		},
		Strict: true,
	}

	// Step 2: Detailed planning with JSON schema
	planningSchema := &llm.JSONSchema{
		Name:        "detailed_plan",
		Description: "Detailed execution plan with timeline",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"execution_phases": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"phase_name":   map[string]interface{}{"type": "string"},
							"duration":     map[string]interface{}{"type": "string"},
							"deliverables": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
						},
						"required":             []string{"phase_name", "duration", "deliverables"},
						"additionalProperties": false,
					},
				},
				"success_metrics": map[string]interface{}{
					"type":  "array",
					"items": map[string]interface{}{"type": "string"},
				},
			},
			"required":             []string{"execution_phases", "success_metrics"},
			"additionalProperties": false,
		},
		Strict: true,
	}

	// Create agents
	analysisAgent := agent.NewChatAgent("analyzer").
		WithModel("gpt-4o-mini").
		WithPrompt("Break down this complex task into manageable components with priorities.").
		WithJSONSchema(initialSchema).
		WithClient(llmClient)

	planningAgent := agent.NewChatAgent("planner").
		WithModel("gpt-4o-mini").
		WithPrompt("Create a detailed execution plan based on the analysis. Focus on phases, deliverables, and success metrics.").
		WithJSONSchema(planningSchema).
		WithClient(llmClient)

	// Create sequential workflow
	planningWorkflow := workflow.NewSequentialFlow("project-planning").
		Then(analysisAgent).
		ThenChain(planningAgent) // Chain passes output from previous step as input

	// Execute workflow
	ctx := workflow.NewWorkContext(context.Background())
	ctx.Set("user_input", "Launch a mobile app for food delivery in my city")

	fmt.Println("Input: Launch a mobile app for food delivery in my city")
	fmt.Println("Running multi-step analysis workflow...")

	report := planningWorkflow.Run(ctx)

	if report.Status == workflow.StatusCompleted {
		// The final output will be from the planning agent
		if response, ok := report.Data.(*llm.CompletionResponse); ok {
			var plan map[string]interface{}
			if err := json.Unmarshal([]byte(response.Content), &plan); err != nil {
				fmt.Printf("Error parsing final plan: %v\n", err)
			} else {
				fmt.Println("✅ Detailed Execution Plan:")
				prettyJSON, _ := json.MarshalIndent(plan, "", "  ")
				fmt.Println(string(prettyJSON))
			}
		}

		// Also show the intermediate analysis result if available
		if analysisData, ok := ctx.Get("analyzer"); ok {
			if analysisResponse, ok := analysisData.(*llm.CompletionResponse); ok {
				var analysis map[string]interface{}
				if err := json.Unmarshal([]byte(analysisResponse.Content), &analysis); err == nil {
					fmt.Println("\n📋 Initial Analysis:")
					prettyJSON, _ := json.MarshalIndent(analysis, "", "  ")
					fmt.Println(string(prettyJSON))
				}
			}
		}
	} else {
		fmt.Printf("❌ Workflow failed: %v\n", report.Errors)
	}
}
