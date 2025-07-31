package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/ratlabs-io/go-agent-kit/pkg/constants"
	"github.com/ratlabs-io/go-agent-kit/pkg/llm"
	"github.com/ratlabs-io/go-agent-kit/pkg/tools"
	"github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

// ToolAgent represents an agent that can use tools and may execute internal workflows
// for multi-step tool execution. It implements the Agent interface.
type ToolAgent struct {
	name         string
	agentType    AgentType
	model        string
	prompt       string
	tools        []tools.Tool
	client       llm.Client
	toolFlow     workflow.Action // Internal workflow for complex tool execution
	maxToolCalls int             // Maximum number of tool calls per execution
	jsonSchema   *llm.JSONSchema
	responseType llm.ResponseType
	maxTokens    int
	temperature  float64
	topP         float64
	log          *slog.Logger
}

// NewToolAgent creates a new ToolAgent with the given name.
func NewToolAgent(name string) *ToolAgent {
	return &ToolAgent{
		name:         name,
		agentType:    TypeTool,
		tools:        []tools.Tool{},
		maxToolCalls: 5,    // Default maximum tool calls
		maxTokens:    4000, // Default max tokens
		temperature:  0.7,  // Default temperature
		topP:         0.95, // Default top-p
		log:          slog.With("agent", "ToolAgent", "name", name),
	}
}

// Name returns the name of the ToolAgent.
func (ta *ToolAgent) Name() string {
	return ta.name
}

// Type returns the type of the agent.
func (ta *ToolAgent) Type() AgentType {
	return ta.agentType
}

// Tools returns the list of tools available to this agent.
func (ta *ToolAgent) Tools() []tools.Tool {
	return ta.tools
}

// Configure configures the ToolAgent with the provided settings.
func (ta *ToolAgent) Configure(config map[string]interface{}) error {
	if model, ok := config["model"].(string); ok {
		ta.model = model
	}
	if prompt, ok := config["prompt"].(string); ok {
		ta.prompt = prompt
	}
	if maxCalls, ok := config["max_tool_calls"].(int); ok {
		ta.maxToolCalls = maxCalls
	}
	if responseType, ok := config["response_type"].(string); ok {
		ta.responseType = llm.ResponseType(responseType)
	}
	if maxTokens, ok := config["max_tokens"].(int); ok {
		ta.maxTokens = maxTokens
	}
	if temperature, ok := config["temperature"].(float64); ok {
		ta.temperature = temperature
	}
	if topP, ok := config["top_p"].(float64); ok {
		ta.topP = topP
	}
	// Note: LLM client must be provided via WithClient() - no default implementation
	return nil
}

// WithModel sets the model for the ToolAgent.
func (ta *ToolAgent) WithModel(model string) *ToolAgent {
	ta.model = model
	return ta
}

// WithPrompt sets the prompt for the ToolAgent.
func (ta *ToolAgent) WithPrompt(prompt string) *ToolAgent {
	ta.prompt = prompt
	return ta
}

// WithTools adds tools to the ToolAgent.
func (ta *ToolAgent) WithTools(tools ...tools.Tool) *ToolAgent {
	ta.tools = append(ta.tools, tools...)
	return ta
}

// WithClient sets the LLM client for the ToolAgent.
func (ta *ToolAgent) WithClient(client llm.Client) *ToolAgent {
	ta.client = client
	return ta
}

// WithToolFlow sets an internal workflow for complex tool execution.
func (ta *ToolAgent) WithToolFlow(flow workflow.Action) *ToolAgent {
	ta.toolFlow = flow
	return ta
}

// WithMaxToolCalls sets the maximum number of tool calls per execution.
func (ta *ToolAgent) WithMaxToolCalls(max int) *ToolAgent {
	ta.maxToolCalls = max
	return ta
}

// WithJSONSchema sets the JSON schema for structured responses.
func (ta *ToolAgent) WithJSONSchema(schema *llm.JSONSchema) *ToolAgent {
	ta.jsonSchema = schema
	ta.responseType = llm.ResponseTypeJSONSchema
	return ta
}

// WithJSONResponse enables JSON object responses (no specific schema).
func (ta *ToolAgent) WithJSONResponse() *ToolAgent {
	ta.responseType = llm.ResponseTypeJSONObject
	ta.jsonSchema = nil
	return ta
}

// WithResponseType sets the response type for the agent.
func (ta *ToolAgent) WithResponseType(responseType llm.ResponseType) *ToolAgent {
	ta.responseType = responseType
	return ta
}

// WithMaxTokens sets the maximum number of tokens to generate.
func (ta *ToolAgent) WithMaxTokens(maxTokens int) *ToolAgent {
	ta.maxTokens = maxTokens
	return ta
}

// WithTemperature sets the sampling temperature (0.0 to 2.0).
func (ta *ToolAgent) WithTemperature(temperature float64) *ToolAgent {
	ta.temperature = temperature
	return ta
}

// WithTopP sets the nucleus sampling parameter (0.0 to 1.0).
func (ta *ToolAgent) WithTopP(topP float64) *ToolAgent {
	ta.topP = topP
	return ta
}

// Run executes the ToolAgent, potentially using tools and internal workflows.
func (ta *ToolAgent) Run(wctx workflow.WorkContext) workflow.WorkReport {
	startTime := time.Now()

	if ta.client == nil {
		ta.log.Error("no LLM client configured")
		return workflow.NewFailedWorkReport(fmt.Errorf("no LLM client configured for agent %s", ta.name))
	}

	// If we have an internal tool flow, use it for complex execution
	if ta.toolFlow != nil {
		ta.log.Info("executing internal tool flow", "flow", ta.toolFlow.Name())
		return ta.executeWithToolFlow(wctx, startTime)
	}

	// Otherwise, perform simple tool-calling execution
	return ta.executeSimpleToolCalling(wctx, startTime)
}

// executeWithToolFlow runs the internal workflow for complex tool execution.
func (ta *ToolAgent) executeWithToolFlow(wctx workflow.WorkContext, startTime time.Time) workflow.WorkReport {
	// Set up context for the internal flow
	wctx.Set("agent_name", ta.name)
	wctx.Set("available_tools", ta.tools)
	wctx.Set("llm_client", ta.client)
	wctx.Set("prompt", ta.prompt)

	// Execute the internal workflow
	flowReport := ta.toolFlow.Run(wctx)
	elapsed := time.Since(startTime)

	// Enhance the report with agent metadata
	if flowReport.Metadata == nil {
		flowReport.Metadata = make(map[string]interface{})
	}
	flowReport.SetMetadata("agent_name", ta.name)
	flowReport.SetMetadata("agent_type", ta.agentType)
	flowReport.SetMetadata("elapsed", elapsed)
	flowReport.SetMetadata("execution_type", "tool_flow")

	// Note: Event emission is now available through WorkContext interface

	ta.log.Info("tool flow execution completed", "status", flowReport.Status, "elapsed", elapsed)
	return flowReport
}

// executeSimpleToolCalling performs proper tool calling with conversation loop.
func (ta *ToolAgent) executeSimpleToolCalling(wctx workflow.WorkContext, startTime time.Time) workflow.WorkReport {
	// Convert tools to LLM tool definitions
	var toolDefs []llm.ToolDefinition
	for _, tool := range ta.tools {
		toolDefs = append(toolDefs, llm.ToolDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters:  convertSchemaToMap(tool.Parameters()),
		})
	}

	// Build initial messages for the conversation
	var messages []llm.Message

	// Check for runtime message history in context
	var messageHistory []llm.Message
	if runtimeHistory, ok := wctx.Get(constants.KeyMessageHistory); ok {
		if historySlice, ok := runtimeHistory.([]llm.Message); ok {
			messageHistory = historySlice
		}
	}

	// Add message history first
	if len(messageHistory) > 0 {
		messages = append(messages, messageHistory...)
	}

	// Add system prompt if provided (only if not already in history)
	if ta.prompt != "" {
		// Check if system prompt already exists in history
		systemPromptExists := false
		for _, msg := range messageHistory {
			if msg.Role == constants.RoleSystem && msg.Content == ta.prompt {
				systemPromptExists = true
				break
			}
		}
		if !systemPromptExists {
			messages = append(messages, llm.Message{
				Role:    constants.RoleSystem,
				Content: ta.prompt,
			})
		}
	}

	// Check for user input in context
	userInput := ""
	if input, ok := wctx.Get(constants.KeyUserInput); ok {
		if inputStr, ok := input.(string); ok && inputStr != "" {
			userInput = inputStr
		}
	}

	if userInput != "" {
		messages = append(messages, llm.Message{
			Role:    constants.RoleUser,
			Content: userInput,
		})
	}

	// If no messages were built, fall back to prompt-only mode
	var prompt string
	if len(messages) == 0 && ta.prompt != "" {
		prompt = ta.prompt
	}

	// Start the tool calling loop
	var finalResponse *llm.CompletionResponse
	var totalTokens int
	toolCallCount := 0

	for i := 0; i < ta.maxToolCalls; i++ {
		// Prepare the completion request
		req := llm.CompletionRequest{
			Model:        ta.model,
			Prompt:       prompt,
			Messages:     messages,
			Tools:        toolDefs,
			JSONSchema:   ta.jsonSchema,
			ResponseType: ta.responseType,
			MaxTokens:    ta.maxTokens,
			Temperature:  ta.temperature,
			TopP:         ta.topP,
			Metadata: map[string]interface{}{
				"agent_name":     ta.name,
				"agent_type":     ta.agentType,
				"loop_iteration": i + 1,
			},
		}

		// Make LLM completion call
		response, err := ta.client.Complete(wctx.Context(), req)
		if err != nil {
			elapsed := time.Since(startTime)
			ta.log.Error("LLM completion failed", "iteration", i+1, "elapsed", elapsed, "error", err)
			return workflow.NewFailedWorkReport(fmt.Errorf("LLM completion failed on iteration %d: %w", i+1, err))
		}

		totalTokens += response.Usage.TotalTokens
		finalResponse = response

		// If no tool calls, we're done
		if len(response.ToolCalls) == 0 {
			ta.log.Info("tool calling loop completed - no more tools requested", "iterations", i+1, "total_tokens", totalTokens)
			break
		}

		// Add assistant message with tool calls to conversation
		messages = append(messages, llm.Message{
			Role:    constants.RoleAssistant,
			Content: response.Content,
		})

		// Execute each tool call and add results to messages
		for _, toolCall := range response.ToolCalls {
			toolCallCount++

			result, err := ta.executeTool(wctx.Context(), toolCall)
			if err != nil {
				ta.log.Error("tool execution failed", "tool", toolCall.Name, "id", toolCall.ID, "error", err)
				// Add error message to conversation
				messages = append(messages, llm.Message{
					Role:    constants.RoleTool,
					Content: fmt.Sprintf("Error executing tool %s: %v", toolCall.Name, err),
					Name:    toolCall.Name,
				})
				continue
			}

			// Convert tool result to JSON string for the conversation
			resultJSON := ta.formatToolResult(result)

			// Add tool result message to conversation
			messages = append(messages, llm.Message{
				Role:    constants.RoleTool,
				Content: resultJSON,
				Name:    toolCall.Name,
			})

			ta.log.Info("tool executed successfully", "tool", toolCall.Name, "id", toolCall.ID, "iteration", i+1)
		}

		// Clear prompt for subsequent iterations (we have messages now)
		prompt = ""
	}

	// Check if we hit max tool calls limit
	if toolCallCount >= ta.maxToolCalls {
		ta.log.Warn("reached maximum tool calls limit", "max_calls", ta.maxToolCalls, "total_calls", toolCallCount)
	}

	// Create final report
	report := workflow.NewCompletedWorkReport()
	report.Data = finalResponse

	elapsed := time.Since(startTime)
	ta.log.Info("tool calling loop completed", "elapsed", elapsed, "total_tokens", totalTokens, "tool_calls", toolCallCount)

	// Add completion metadata
	ta.addCompletionMetadata(&report, finalResponse, startTime)

	// Override some metadata with loop-specific info
	report.SetMetadata("total_tokens", totalTokens)
	report.SetMetadata("tool_calls_count", toolCallCount)
	report.SetMetadata("execution_type", "tool_calling_loop")

	// Wait for callbacks to complete if WorkContext supports waiting
	if ctxValue := wctx.Context().Value(constants.KeyWorkContext); ctxValue != nil {
		if workCtx, ok := ctxValue.(workflow.WorkContext); ok {
			workCtx.Wait()
		}
	}

	return report
}

// formatToolResult converts a tool execution result to JSON string for LLM conversation.
func (ta *ToolAgent) formatToolResult(result interface{}) string {
	// Handle nil results
	if result == nil {
		return "null"
	}

	// Handle string results (already formatted)
	if str, ok := result.(string); ok {
		return str
	}

	// Handle structured results (convert to JSON)
	if resultData, err := json.Marshal(result); err == nil {
		return string(resultData)
	}

	// Fallback to string representation
	return fmt.Sprintf("%v", result)
}

// executeTool executes a single tool call.
func (ta *ToolAgent) executeTool(ctx context.Context, toolCall llm.ToolCall) (interface{}, error) {
	// Find the tool in our registry
	var targetTool tools.Tool
	for _, tool := range ta.tools {
		if tool.Name() == toolCall.Name {
			targetTool = tool
			break
		}
	}

	if targetTool == nil {
		return nil, fmt.Errorf("tool %s not found in agent registry", toolCall.Name)
	}

	// Execute the tool
	return targetTool.Execute(ctx, toolCall.Args)
}

// addCompletionMetadata adds completion metadata to the work report.
func (ta *ToolAgent) addCompletionMetadata(report *workflow.WorkReport, response *llm.CompletionResponse, startTime time.Time) {
	elapsed := time.Since(startTime)

	report.SetMetadata("agent_name", ta.name)
	report.SetMetadata("agent_type", ta.agentType)
	report.SetMetadata("elapsed", elapsed)
	report.SetMetadata("token_usage", response.Usage)
	report.SetMetadata("tool_calls_count", len(response.ToolCalls))
	report.SetMetadata("execution_type", "simple_tool_calling")

	// Add agent completion event
	event := workflow.Event{
		Type:      workflow.EventAgentCompleted,
		Source:    ta.name,
		Timestamp: time.Now(),
		Payload:   response,
		Metadata: map[string]interface{}{
			"agent_type":       ta.agentType,
			"elapsed":          elapsed,
			"tool_calls_count": len(response.ToolCalls),
		},
	}
	report.AddEvent(event)
}
