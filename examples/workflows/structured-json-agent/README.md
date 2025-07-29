# Structured JSON Agent Example

This example demonstrates how to use agents with structured JSON responses, including JSON schemas for predictable, parseable outputs.

## Features Demonstrated

1. **JSON Schema Responses**: Define strict schemas for structured output
2. **Generic JSON Responses**: Request JSON without specific schemas  
3. **Multi-Step JSON Workflows**: Chain agents with structured inputs/outputs
4. **Type-Safe Parsing**: Parse JSON responses into Go structs

## What This Example Shows

### 1. Task Analysis with JSON Schema
- Creates a structured schema for task categorization
- Parses responses into a Go struct
- Handles multiple input types consistently

### 2. Generic JSON Responses  
- Requests JSON output without a specific schema
- Useful for exploratory or flexible responses
- Pretty-prints the JSON output

### 3. Multi-Step Analysis Workflow
- Chains two agents with different JSON schemas
- First agent breaks down the task
- Second agent creates detailed execution plans
- Shows how structured data flows between workflow steps

## Key Code Patterns

### Defining a JSON Schema
```go
schema := &llm.JSONSchema{
    Name:        "task_analysis",
    Description: "Analysis of a user task",
    Schema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "category": map[string]interface{}{
                "type": "string",
                "enum": []string{"question", "request", "task", "other"},
            },
            // ... more properties
        },
        "required": []string{"category", "complexity"},
        "additionalProperties": false, // Required by OpenAI for all objects
    },
    Strict: true, // Enforce strict adherence to schema
}
```

**Important OpenAI Requirements for Strict Mode**:
1. `"additionalProperties": false` must be set on all object schemas (root and nested)
2. ALL properties defined in the schema MUST be listed in the `required` array
3. If a field is optional, don't include it in the schema at all
4. The `required` array must include every key in `properties`

### Creating an Agent with JSON Schema
```go
agent := agent.NewChatAgent("analyzer").
    WithModel("gpt-4").
    WithPrompt("Your analysis prompt here").
    WithJSONSchema(schema).
    WithClient(llmClient)
```

### Creating an Agent with Generic JSON
```go
agent := agent.NewChatAgent("json-responder").
    WithModel("gpt-3.5-turbo").
    WithPrompt("Respond with JSON insights").
    WithJSONResponse().
    WithClient(llmClient)
```

### Parsing Structured Responses
```go
type TaskAnalysis struct {
    Category      string   `json:"category"`
    Complexity    string   `json:"complexity"`
    EstimatedTime int      `json:"estimated_time"`
    Requirements  []string `json:"requirements"`
}

var analysis TaskAnalysis
json.Unmarshal([]byte(response.Content), &analysis)
```

## Running the Example

```bash
# Set your OpenAI API key
export OPENAI_API_KEY=your-actual-api-key-here

# Run the example
go run examples/workflows/structured-json-agent/main.go
```

## Expected Output

The example will show:

1. **Task Analysis**: Structured categorization of different types of user inputs
2. **Business Insights**: Generic JSON response with flexible structure
3. **Project Planning**: Multi-step workflow creating detailed execution plans

Each section demonstrates different aspects of structured JSON responses and how they can make agent outputs more predictable and easier to integrate into applications.

## Use Cases

- **Data Processing**: Extract structured information from unstructured text
- **Classification**: Categorize inputs with consistent schema
- **Planning**: Generate structured plans and breakdowns
- **API Integration**: Create predictable outputs for downstream systems
- **Form Generation**: Structure data for UI components
- **Workflow Steps**: Pass structured data between workflow stages

## LLM Provider Support

JSON schema responses have specific model requirements:

### OpenAI Models
- **JSON Schema Support** (`response_format: {type: "json_schema"}`):
  - ✅ gpt-4o, gpt-4o-mini (recommended)
  - ✅ gpt-4-turbo-preview, gpt-4-turbo
  - ❌ gpt-3.5-turbo (use json_object instead)
  
- **JSON Object Support** (`response_format: {type: "json_object"}`):
  - ✅ All GPT models including gpt-3.5-turbo

### Other Providers
- **Anthropic Claude**: Supports JSON through prompt engineering
- **Google Gemini**: Supports structured outputs (varies by model)
- **Other providers**: Check documentation for JSON support

**Important**: If you get plain text instead of JSON, verify your model supports the requested format. The example uses `gpt-4o-mini` which fully supports JSON schemas.