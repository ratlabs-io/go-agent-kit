package builtin

import (
	"context"
	"fmt"

	"github.com/ratlabs-io/go-agent-kit/pkg/tools"
)

// MCPExample demonstrates how to connect to an external MCP server
// and use its tools within the go-agent-kit framework.
func MCPExample() {
	// Define MCP server connection
	server := &tools.MCPServer{
		Name:   "example-mcp-server",
		URL:    "http://localhost:8080/mcp",
		APIKey: "your-api-key-here", // optional
	}
	
	// Create MCP client
	mcpClient := tools.NewMCPClient(server)
	
	// Create tool registry
	registry := tools.NewDefaultToolRegistry()
	
	// Fetch and register tools from MCP server
	ctx := context.Background()
	mcpTools, err := mcpClient.ListTools(ctx)
	if err != nil {
		fmt.Printf("Failed to list MCP tools: %v\n", err)
		return
	}
	
	// Register all MCP tools
	for _, tool := range mcpTools {
		if err := registry.Register(tool); err != nil {
			fmt.Printf("Failed to register tool %s: %v\n", tool.Name(), err)
			continue
		}
		fmt.Printf("Registered MCP tool: %s\n", tool.Name())
	}
	
	// Now the MCP tools can be used by agents just like native tools
	fmt.Printf("Successfully registered %d MCP tools\n", len(mcpTools))
}

// MCPRegistry creates a tool registry pre-configured with MCP tools
// from multiple servers. This demonstrates how to manage multiple MCP sources.
func MCPRegistry(servers []*tools.MCPServer) (*tools.DefaultToolRegistry, error) {
	registry := tools.NewDefaultToolRegistry()
	ctx := context.Background()
	
	for _, server := range servers {
		mcpClient := tools.NewMCPClient(server)
		
		tools, err := mcpClient.ListTools(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list tools from server %s: %w", server.Name, err)
		}
		
		for _, tool := range tools {
			// Prefix tool names with server name to avoid conflicts
			prefixedTool := &MCPToolProxy{
				tool:       tool,
				namePrefix: server.Name + ".",
			}
			
			if err := registry.Register(prefixedTool); err != nil {
				return nil, fmt.Errorf("failed to register tool %s from %s: %w", 
					tool.Name(), server.Name, err)
			}
		}
	}
	
	return registry, nil
}

// MCPToolProxy wraps an MCP tool to add name prefixing for multi-server setups.
type MCPToolProxy struct {
	tool       tools.Tool
	namePrefix string
}

func (mtp *MCPToolProxy) Name() string {
	return mtp.namePrefix + mtp.tool.Name()
}

func (mtp *MCPToolProxy) Description() string {
	return mtp.tool.Description()
}

func (mtp *MCPToolProxy) Parameters() tools.Schema {
	return mtp.tool.Parameters()
}

func (mtp *MCPToolProxy) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return mtp.tool.Execute(ctx, params)
}