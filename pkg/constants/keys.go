package constants

// Context Keys
// These constants define the standard keys used throughout the go-agent-kit library
// for storing and retrieving data from the WorkContext.

const (
	// WorkContext Data Keys - used with ctx.Set() and ctx.Get()
	
	// KeyUserInput is the key for user input text in the WorkContext.
	// Used by agents to get the current user message or instruction.
	KeyUserInput = "user_input"
	
	// KeyMessageHistory is the key for conversation history in the WorkContext.
	// Contains a slice of llm.Message representing the conversation context.
	KeyMessageHistory = "message_history"
	
	// KeyPreviousOutput is the key for storing the output of the previous action.
	// Used by sequential workflows to pass data between chained actions.
	KeyPreviousOutput = "previous_output"
	
	// KeyOriginalInput is the key for storing the original user input.
	// Used by accumulating workflows to preserve the initial user message.
	KeyOriginalInput = "original_input"
	
	// Loop Context Keys - used by loop constructs
	
	// KeyCurrentItem is the key for the current item in an iterator loop.
	// Used by NewLoopOver to provide access to the current item being processed.
	KeyCurrentItem = "current_item"
	
	// KeyCurrentIndex is the key for the current index in an iterator loop.
	// Used by NewLoopOver to provide access to the current zero-based index.
	KeyCurrentIndex = "current_index"
	
	// KeyLoopIteration is the key for the current iteration number.
	// Used by all loop constructs to track the current iteration (1-based).
	KeyLoopIteration = "loop_iteration"
)

const (
	// Context Value Keys - used with context.WithValue() and context.Value()
	
	// KeyWorkContext is the key for storing the WorkContext in the Go context.
	// Used internally for accessing WorkContext from nested function calls.
	KeyWorkContext = "work_context"
	
	// KeyLogger is the key for storing a custom logger in the Go context.
	// Used by the logging system to retrieve context-specific loggers.
	KeyLogger = "logger"
)

const (
	// Message Roles - used in llm.Message.Role field
	
	// RoleSystem indicates a system message containing instructions or context.
	// System messages typically contain the agent's behavior instructions.
	RoleSystem = "system"
	
	// RoleUser indicates a message from the user/human.
	// User messages contain questions, requests, or input from the end user.
	RoleUser = "user"
	
	// RoleAssistant indicates a message from the AI assistant.
	// Assistant messages contain the AI's responses and generated content.
	RoleAssistant = "assistant"
	
	// RoleTool indicates a message from a tool execution.
	// Tool messages contain the results of function/tool calls.
	RoleTool = "tool"
)

const (
	// Agent Types - used to identify different types of agents
	
	// AgentTypeChat identifies a simple chat completion agent.
	// Chat agents perform single-hop LLM completions without tool support.
	AgentTypeChat = "chat"
	
	// AgentTypeTool identifies a tool-enabled agent.
	// Tool agents can call functions and use external tools.
	AgentTypeTool = "tool"
)

const (
	// Workflow Types - used for logging and identification
	
	// FlowTypeSequential identifies sequential workflow execution.
	FlowTypeSequential = "SequentialFlow"
	
	// FlowTypeParallel identifies parallel workflow execution.
	FlowTypeParallel = "ParallelFlow"
	
	// FlowTypeConditional identifies conditional workflow execution.
	FlowTypeConditional = "ConditionalFlow"
	
	// FlowTypeSwitch identifies switch-based workflow execution.
	FlowTypeSwitch = "SwitchFlow"
	
	// FlowTypeLoop identifies loop-based workflow execution.
	FlowTypeLoop = "LoopFlow"
	
	// FlowTypeRetry identifies retry-based workflow execution.
	FlowTypeRetry = "RetryFlow"
	
	// FlowTypeTryCatch identifies try-catch workflow execution.
	FlowTypeTryCatch = "TryCatchFlow"
)

const (
	// LLM Response Types - used to specify desired response format
	
	// ResponseTypeText requests a standard text response from the LLM.
	ResponseTypeText = "text"
	
	// ResponseTypeJSONObject requests a valid JSON object response.
	ResponseTypeJSONObject = "json_object"
	
	// ResponseTypeJSONSchema requests a response following a specific JSON schema.
	ResponseTypeJSONSchema = "json_schema"
)

const (
	// HTTP and MCP Constants
	
	// HTTPMethodPost is the HTTP method used for MCP tool calls.
	HTTPMethodPost = "POST"
	
	// MCPMethodToolsCall is the MCP method name for executing tools.
	MCPMethodToolsCall = "tools/call" 
	
	// MCPMethodToolsList is the MCP method name for listing available tools.
	MCPMethodToolsList = "tools/list"
	
	// HTTPHeaderAuthorization is the Authorization header name.
	HTTPHeaderAuthorization = "Authorization"
	
	// HTTPHeaderContentType is the Content-Type header name.
	HTTPHeaderContentType = "Content-Type"
	
	// ContentTypeJSON is the JSON content type.
	ContentTypeJSON = "application/json"
)

const (
	// JSON Field Names - used in structured data
	
	// FieldName is the JSON field name for names.
	FieldName = "name"
	
	// FieldDescription is the JSON field name for descriptions.
	FieldDescription = "description"
	
	// FieldType is the JSON field name for types.
	FieldType = "type"
	
	// FieldData is the JSON field name for data payloads.
	FieldData = "data"
	
	// FieldError is the JSON field name for error information.
	FieldError = "error"
	
	// FieldMetadata is the JSON field name for metadata.
	FieldMetadata = "metadata"
	
	// FieldResult is the JSON field name for results.
	FieldResult = "result"
	
	// FieldMethod is the JSON field name for method names.
	FieldMethod = "method"
	
	// FieldParams is the JSON field name for parameters.
	FieldParams = "params"
	
	// FieldArguments is the JSON field name for arguments.
	FieldArguments = "arguments"
	
	// FieldTools is the JSON field name for tools arrays.
	FieldTools = "tools"
	
	// FieldSchema is the JSON field name for schema definitions.
	FieldSchema = "schema"
	
	// FieldProperties is the JSON field name for schema properties.
	FieldProperties = "properties"
	
	// FieldRequired is the JSON field name for required fields.
	FieldRequired = "required"
	
	// FieldContent is the JSON field name for message content.
	FieldContent = "Content"
)

const (
	// Schema Types - used in JSON schema definitions
	
	// SchemaTypeObject indicates an object type in JSON schema.
	SchemaTypeObject = "object"
	
	// SchemaTypeString indicates a string type in JSON schema.
	SchemaTypeString = "string"
	
	// SchemaTypeNumber indicates a number type in JSON schema.
	SchemaTypeNumber = "number"
	
	// SchemaTypeBoolean indicates a boolean type in JSON schema.
	SchemaTypeBoolean = "boolean"
	
	// SchemaTypeArray indicates an array type in JSON schema.
	SchemaTypeArray = "array"
)