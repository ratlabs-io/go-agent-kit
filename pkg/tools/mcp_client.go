package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// MCPServer represents connection details for an MCP server.
type MCPServer struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	APIKey  string `json:"api_key,omitempty"`
	Timeout int    `json:"timeout,omitempty"` // timeout in seconds
}

// MCPClient handles communication with external MCP servers.
type MCPClient struct {
	server *MCPServer
	client *http.Client
}

// NewMCPClient creates a new MCP client for the given server.
func NewMCPClient(server *MCPServer) *MCPClient {
	return &MCPClient{
		server: server,
		client: &http.Client{},
	}
}

// MCPTool represents a tool available on an MCP server.
type MCPTool struct {
	client *MCPClient
	name   string
	desc   string
	schema Schema
}

// NewMCPTool creates a tool that proxies to an MCP server.
func NewMCPTool(client *MCPClient, name, description string, schema Schema) *MCPTool {
	return &MCPTool{
		client: client,
		name:   name,
		desc:   description,
		schema: schema,
	}
}

// Name returns the tool name.
func (mt *MCPTool) Name() string {
	return mt.name
}

// Description returns the tool description.
func (mt *MCPTool) Description() string {
	return mt.desc
}

// Parameters returns the tool's parameter schema.
func (mt *MCPTool) Parameters() Schema {
	return mt.schema
}

// Execute calls the MCP server to execute the tool.
func (mt *MCPTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Prepare MCP request
	mcpRequest := map[string]interface{}{
		"method": "tools/call",
		"params": map[string]interface{}{
			"name":      mt.name,
			"arguments": params,
		},
	}
	
	// Marshal request
	reqBody, err := json.Marshal(mcpRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal MCP request: %w", err)
	}
	
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", mt.client.server.URL, 
		bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if mt.client.server.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+mt.client.server.APIKey)
	}
	
	// Make request
	resp, err := mt.client.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("MCP request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read MCP response: %w", err)
	}
	
	// Check status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MCP server returned status %d: %s", resp.StatusCode, string(body))
	}
	
	// Parse response
	var mcpResponse map[string]interface{}
	if err := json.Unmarshal(body, &mcpResponse); err != nil {
		return nil, fmt.Errorf("failed to parse MCP response: %w", err)
	}
	
	// Extract result
	if result, ok := mcpResponse["result"]; ok {
		return result, nil
	}
	
	// Check for errors
	if mcpErr, ok := mcpResponse["error"]; ok {
		return nil, fmt.Errorf("MCP tool execution failed: %v", mcpErr)
	}
	
	return mcpResponse, nil
}

// ListTools fetches available tools from the MCP server.
func (mc *MCPClient) ListTools(ctx context.Context) ([]Tool, error) {
	// Prepare MCP list request
	mcpRequest := map[string]interface{}{
		"method": "tools/list",
		"params": map[string]interface{}{},
	}
	
	// Marshal request
	reqBody, err := json.Marshal(mcpRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal MCP list request: %w", err)
	}
	
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", mc.server.URL, 
		bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if mc.server.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+mc.server.APIKey)
	}
	
	// Make request
	resp, err := mc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("MCP list request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read MCP response: %w", err)
	}
	
	// Check status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MCP server returned status %d: %s", resp.StatusCode, string(body))
	}
	
	// Parse response
	var mcpResponse map[string]interface{}
	if err := json.Unmarshal(body, &mcpResponse); err != nil {
		return nil, fmt.Errorf("failed to parse MCP response: %w", err)
	}
	
	// Extract tools
	result, ok := mcpResponse["result"]
	if !ok {
		return nil, fmt.Errorf("MCP response missing result field")
	}
	
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("MCP result is not an object")
	}
	
	toolsData, ok := resultMap["tools"]
	if !ok {
		return nil, fmt.Errorf("MCP result missing tools field")
	}
	
	toolsList, ok := toolsData.([]interface{})
	if !ok {
		return nil, fmt.Errorf("MCP tools field is not an array")
	}
	
	// Convert to Tool instances
	var tools []Tool
	for _, toolData := range toolsList {
		toolMap, ok := toolData.(map[string]interface{})
		if !ok {
			continue
		}
		
		name, _ := toolMap["name"].(string)
		description, _ := toolMap["description"].(string)
		
		// Parse schema if available
		schema := Schema{
			Type:        "object",
			Description: "Parameters for " + name,
			Properties:  map[string]interface{}{},
			Required:    []string{},
		}
		
		if schemaData, ok := toolMap["schema"].(map[string]interface{}); ok {
			// Convert schema data
			if schemaType, ok := schemaData["type"].(string); ok {
				schema.Type = schemaType
			}
			if props, ok := schemaData["properties"].(map[string]interface{}); ok {
				schema.Properties = props
			}
			if req, ok := schemaData["required"].([]interface{}); ok {
				for _, r := range req {
					if reqStr, ok := r.(string); ok {
						schema.Required = append(schema.Required, reqStr)
					}
				}
			}
		}
		
		tools = append(tools, NewMCPTool(mc, name, description, schema))
	}
	
	return tools, nil
}