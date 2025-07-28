package agent

import (
	"context"
	"fmt"
	"log/slog"
	"time"

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
	maxToolCalls int            // Maximum number of tool calls per execution
	log          *slog.Logger
}

// NewToolAgent creates a new ToolAgent with the given name.
func NewToolAgent(name string) *ToolAgent {
	return &ToolAgent{
		name:         name,
		agentType:    TypeTool,
		tools:        []tools.Tool{},
		maxToolCalls: 5, // Default maximum tool calls
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

// executeSimpleToolCalling performs simple LLM completion with tool calling.
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
	
	// Build messages for the completion request
	var messages []llm.Message
	
	// Add system prompt if provided
	if ta.prompt != "" {
		messages = append(messages, llm.Message{
			Role:    "system",
			Content: ta.prompt,
		})
	}
	
	// Check for user input in context (check both "user_input" and "task" for compatibility)
	userInput := ""
	if input, ok := wctx.Get("user_input"); ok {
		if inputStr, ok := input.(string); ok && inputStr != "" {
			userInput = inputStr
		}
	} else if task, ok := wctx.Get("task"); ok {
		if taskStr, ok := task.(string); ok && taskStr != "" {
			userInput = taskStr
		}
	}
	
	if userInput != "" {
		messages = append(messages, llm.Message{
			Role:    "user",
			Content: userInput,
		})
	}
	
	// If no messages were built, fall back to prompt-only mode
	var prompt string
	if len(messages) == 0 && ta.prompt != "" {
		prompt = ta.prompt
	}
	
	// Prepare the completion request
	req := llm.CompletionRequest{
		Model:    ta.model,
		Prompt:   prompt,
		Messages: messages,
		Tools:    toolDefs,
		Metadata: map[string]interface{}{
			"agent_name": ta.name,
			"agent_type": ta.agentType,
		},
	}
	
	// For now, perform a simple completion
	// TODO: Implement proper tool calling loop when gollm supports it
	response, err := ta.client.Complete(wctx.Context(), req)
	if err != nil {
		elapsed := time.Since(startTime)
		ta.log.Error("LLM completion failed", "elapsed", elapsed, "error", err)
		return workflow.NewFailedWorkReport(fmt.Errorf("LLM completion failed: %w", err))
	}
	
	// Process any tool calls in the response
	report := ta.processToolCalls(wctx, response, startTime)
	
	elapsed := time.Since(startTime)
	ta.log.Info("simple tool calling completed", "elapsed", elapsed, "tokens", response.Usage.TotalTokens)
	
	// Wait for callbacks to complete if WorkContext supports waiting
	if ctxValue := wctx.Context().Value("work_context"); ctxValue != nil {
		if workCtx, ok := ctxValue.(workflow.WorkContext); ok {
			workCtx.Wait()
		}
	}
	
	return report
}

// processToolCalls handles tool calls from the LLM response.
func (ta *ToolAgent) processToolCalls(wctx workflow.WorkContext, response *llm.CompletionResponse, startTime time.Time) workflow.WorkReport {
	report := workflow.NewCompletedWorkReport()
	report.Data = response
	
	// If no tool calls, return the response as-is
	if len(response.ToolCalls) == 0 {
		ta.addCompletionMetadata(&report, response, startTime)
		return report
	}
	
	// Execute tool calls
	toolResults := make(map[string]interface{})
	
	for _, toolCall := range response.ToolCalls {
		result, err := ta.executeTool(wctx.Context(), toolCall)
		if err != nil {
			ta.log.Error("tool execution failed", "tool", toolCall.Name, "error", err)
			report.AddError(fmt.Errorf("tool %s failed: %w", toolCall.Name, err))
			continue
		}
		
		toolResults[toolCall.ID] = result
		ta.log.Info("tool executed successfully", "tool", toolCall.Name, "id", toolCall.ID)
	}
	
	// Add tool results to response data
	if responseData, ok := report.Data.(*llm.CompletionResponse); ok {
		if responseData.Metadata == nil {
			responseData.Metadata = make(map[string]interface{})
		}
		responseData.Metadata["tool_results"] = toolResults
	}
	
	ta.addCompletionMetadata(&report, response, startTime)
	return report
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