package agent

import (
	"fmt"
	"time"

	"github.com/ratlabs-io/go-agent-kit/pkg/llm"
	"github.com/ratlabs-io/go-agent-kit/pkg/tools"
	"github.com/ratlabs-io/go-agent-kit/pkg/workflow"
)

// ChatAgent represents a simple agent that performs 1-hop LLM completions.
// It implements the Agent interface and can be used as a node in workflows.
// For tool-calling capabilities, use ToolAgent instead.
type ChatAgent struct {
	name         string
	agentType    AgentType
	model        string
	prompt       string
	client       llm.Client
	jsonSchema   *llm.JSONSchema
	responseType llm.ResponseType
}

// NewChatAgent creates a new ChatAgent with the given name.
func NewChatAgent(name string) *ChatAgent {
	return &ChatAgent{
		name:      name,
		agentType: TypeChat,
	}
}

// Name returns the name of the ChatAgent.
func (ca *ChatAgent) Name() string {
	return ca.name
}

// Type returns the type of the agent.
func (ca *ChatAgent) Type() AgentType {
	return ca.agentType
}

// Tools returns the list of tools available to this agent.
// ChatAgent doesn't support tools - use ToolAgent for tool-calling capabilities.
func (ca *ChatAgent) Tools() []tools.Tool {
	return nil
}

// Configure configures the ChatAgent with the provided settings.
func (ca *ChatAgent) Configure(config map[string]interface{}) error {
	if model, ok := config["model"].(string); ok {
		ca.model = model
	}
	if prompt, ok := config["prompt"].(string); ok {
		ca.prompt = prompt
	}
	if responseType, ok := config["response_type"].(string); ok {
		ca.responseType = llm.ResponseType(responseType)
	}
	// Note: LLM client must be provided via WithClient() - no default implementation
	return nil
}

// WithModel sets the model for the ChatAgent.
func (ca *ChatAgent) WithModel(model string) *ChatAgent {
	ca.model = model
	return ca
}

// WithPrompt sets the prompt for the ChatAgent.
func (ca *ChatAgent) WithPrompt(prompt string) *ChatAgent {
	ca.prompt = prompt
	return ca
}

// WithTools is not supported by ChatAgent.
// Use ToolAgent for tool-calling capabilities.
func (ca *ChatAgent) WithTools(tools ...tools.Tool) *ChatAgent {
	// No-op: ChatAgent doesn't support tools
	return ca
}

// WithClient sets the LLM client for the ChatAgent.
func (ca *ChatAgent) WithClient(client llm.Client) *ChatAgent {
	ca.client = client
	return ca
}

// WithJSONSchema sets the JSON schema for structured responses.
func (ca *ChatAgent) WithJSONSchema(schema *llm.JSONSchema) *ChatAgent {
	ca.jsonSchema = schema
	ca.responseType = llm.ResponseTypeJSONSchema
	return ca
}

// WithJSONResponse enables JSON object responses (no specific schema).
func (ca *ChatAgent) WithJSONResponse() *ChatAgent {
	ca.responseType = llm.ResponseTypeJSONObject
	ca.jsonSchema = nil
	return ca
}

// WithResponseType sets the response type for the agent.
func (ca *ChatAgent) WithResponseType(responseType llm.ResponseType) *ChatAgent {
	ca.responseType = responseType
	return ca
}


// Run executes the ChatAgent by performing a single LLM completion.
func (ca *ChatAgent) Run(wctx workflow.WorkContext) workflow.WorkReport {
	startTime := time.Now()
	logger := wctx.Logger().With("agent", "ChatAgent", "name", ca.name)
	
	if ca.client == nil {
		logger.Error("no LLM client configured")
		return workflow.NewFailedWorkReport(fmt.Errorf("no LLM client configured for agent %s", ca.name))
	}
	
	// ChatAgent doesn't support tools - use ToolAgent for tool-calling
	var toolDefs []llm.ToolDefinition
	
	// Build messages for the completion request
	var messages []llm.Message
	
	// Check for runtime message history in context
	var messageHistory []llm.Message
	if runtimeHistory, ok := wctx.Get("message_history"); ok {
		if historySlice, ok := runtimeHistory.([]llm.Message); ok {
			messageHistory = historySlice
		}
	}
	
	// Add message history first
	if len(messageHistory) > 0 {
		messages = append(messages, messageHistory...)
	}
	
	// Add system prompt if provided (only if not already in history)
	if ca.prompt != "" {
		// Check if system prompt already exists in history
		systemPromptExists := false
		for _, msg := range messageHistory {
			if msg.Role == "system" && msg.Content == ca.prompt {
				systemPromptExists = true
				break
			}
		}
		if !systemPromptExists {
			messages = append(messages, llm.Message{
				Role:    "system",
				Content: ca.prompt,
			})
		}
	}
	
	// Check for user input in context
	if userInput, ok := wctx.Get("user_input"); ok {
		if userInputStr, ok := userInput.(string); ok && userInputStr != "" {
			messages = append(messages, llm.Message{
				Role:    "user", 
				Content: userInputStr,
			})
		}
	}
	
	// If no messages were built, fall back to prompt-only mode
	var prompt string
	if len(messages) == 0 && ca.prompt != "" {
		prompt = ca.prompt
	}
	
	// Prepare the completion request
	req := llm.CompletionRequest{
		Model:        ca.model,
		Prompt:       prompt,
		Messages:     messages,
		Tools:        toolDefs,
		JSONSchema:   ca.jsonSchema,
		ResponseType: ca.responseType,
		Metadata: map[string]interface{}{
			"agent_name": ca.name,
			"agent_type": ca.agentType,
		},
	}
	
	// Perform the LLM completion
	response, err := ca.client.Complete(wctx.Context(), req)
	if err != nil {
		elapsed := time.Since(startTime)
		logger.Error("LLM completion failed", "elapsed", elapsed, "error", err)
		return workflow.NewFailedWorkReport(fmt.Errorf("LLM completion failed: %w", err))
	}
	
	elapsed := time.Since(startTime)
	logger.Info("LLM completion successful", "elapsed", elapsed, "tokens", response.Usage.TotalTokens)
	
	// Create the work report
	report := workflow.NewCompletedWorkReport()
	report.Data = response
	
	// Emit event if WorkContext supports events
	// Check if this WorkContext has event capabilities by looking at the context value
	if ctxValue := wctx.Context().Value("work_context"); ctxValue != nil {
		if workCtx, ok := ctxValue.(workflow.WorkContext); ok {
			event := workflow.Event{
				Type:      workflow.EventAgentCompleted,
				Source:    ca.name,
				Timestamp: time.Now(),
				Payload:   response,
				Metadata: map[string]interface{}{
					"agent_type": ca.agentType,
					"elapsed":    elapsed,
				},
			}
			workCtx.EmitEvent(event)
		}
	}
	
	// Add metadata
	report.SetMetadata("agent_name", ca.name)
	report.SetMetadata("agent_type", ca.agentType)
	report.SetMetadata("elapsed", elapsed)
	report.SetMetadata("token_usage", response.Usage)
	
	// Wait for callbacks to complete if WorkContext supports waiting
	if ctxValue := wctx.Context().Value("work_context"); ctxValue != nil {
		if workCtx, ok := ctxValue.(workflow.WorkContext); ok {
			workCtx.Wait()
		}
	}
	
	return report
}

// convertSchemaToMap converts tools.Schema to map[string]interface{} for LLM usage.
func convertSchemaToMap(schema tools.Schema) map[string]interface{} {
	result := map[string]interface{}{
		"type":        schema.Type,
		"description": schema.Description,
	}
	
	if schema.Properties != nil {
		result["properties"] = schema.Properties
	}
	
	if len(schema.Required) > 0 {
		result["required"] = schema.Required
	}
	
	return result
}