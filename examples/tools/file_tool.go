package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ratlabs-io/go-agent-kit/pkg/tools"
)

// FileTool provides file system operations for agents.
type FileTool struct {
	allowedPaths []string // If set, restricts operations to these paths
	readOnly     bool     // If true, only allows read operations
}

// NewFileTool creates a new file tool with default settings (full access).
func NewFileTool() *FileTool {
	return &FileTool{
		allowedPaths: nil,
		readOnly:     false,
	}
}

// NewReadOnlyFileTool creates a new file tool that only allows read operations.
func NewReadOnlyFileTool() *FileTool {
	return &FileTool{
		allowedPaths: nil,
		readOnly:     true,
	}
}

// NewRestrictedFileTool creates a new file tool restricted to specific paths.
func NewRestrictedFileTool(allowedPaths []string, readOnly bool) *FileTool {
	return &FileTool{
		allowedPaths: allowedPaths,
		readOnly:     readOnly,
	}
}

// Name returns the name of the tool.
func (ft *FileTool) Name() string {
	return "file_operations"
}

// Description returns a description of what the tool does.
func (ft *FileTool) Description() string {
	if ft.readOnly {
		return "Read files and directories from the file system. Supports reading file contents and listing directories."
	}
	return "Perform file system operations including reading, writing, and listing files and directories."
}

// Parameters returns the JSON Schema for the tool's parameters.
func (ft *FileTool) Parameters() tools.Schema {
	operations := []string{"read", "list"}
	if !ft.readOnly {
		operations = append(operations, "write", "delete", "mkdir")
	}

	return tools.Schema{
		Type:        "object",
		Description: "Parameters for file system operations",
		Properties: map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "The file operation to perform",
				"enum":        operations,
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "The file or directory path",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Content to write (for write operation)",
			},
			"recursive": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to perform recursive operations (for list, mkdir)",
				"default":     false,
			},
		},
		Required: []string{"operation", "path"},
	}
}

// Execute performs the file operation.
func (ft *FileTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract operation
	operation, ok := params["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation parameter is required and must be a string")
	}

	// Extract path
	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path parameter is required and must be a string")
	}

	// Validate path access
	if err := ft.validatePath(path); err != nil {
		return nil, err
	}

	// Check read-only restrictions
	if ft.readOnly && !isReadOperation(operation) {
		return nil, fmt.Errorf("write operations not allowed in read-only mode")
	}

	switch operation {
	case "read":
		return ft.readFile(path)
	case "write":
		content, _ := params["content"].(string)
		return ft.writeFile(path, content)
	case "list":
		recursive, _ := params["recursive"].(bool)
		return ft.listDirectory(path, recursive)
	case "delete":
		return ft.deleteFile(path)
	case "mkdir":
		recursive, _ := params["recursive"].(bool)
		return ft.makeDirectory(path, recursive)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

// validatePath checks if the path is allowed based on restrictions.
func (ft *FileTool) validatePath(path string) error {
	if len(ft.allowedPaths) == 0 {
		return nil // No restrictions
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	for _, allowedPath := range ft.allowedPaths {
		absAllowedPath, err := filepath.Abs(allowedPath)
		if err != nil {
			continue
		}

		if strings.HasPrefix(absPath, absAllowedPath) {
			return nil
		}
	}

	return fmt.Errorf("path %s is not allowed", path)
}

// isReadOperation checks if an operation is read-only.
func isReadOperation(operation string) bool {
	readOps := map[string]bool{
		"read": true,
		"list": true,
	}
	return readOps[operation]
}

// readFile reads the contents of a file.
func (ft *FileTool) readFile(path string) (interface{}, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return map[string]interface{}{
		"path":     path,
		"content":  string(content),
		"size":     info.Size(),
		"modified": info.ModTime(),
		"mode":     info.Mode().String(),
	}, nil
}

// writeFile writes content to a file.
func (ft *FileTool) writeFile(path, content string) (interface{}, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file %s: %w", path, err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info after write: %w", err)
	}

	return map[string]interface{}{
		"path":     path,
		"size":     info.Size(),
		"modified": info.ModTime(),
		"message":  "File written successfully",
	}, nil
}

// listDirectory lists the contents of a directory.
func (ft *FileTool) listDirectory(path string, recursive bool) (interface{}, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path %s: %w", path, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path %s is not a directory", path)
	}

	var files []map[string]interface{}

	if recursive {
		err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			files = append(files, map[string]interface{}{
				"path":     filePath,
				"name":     info.Name(),
				"size":     info.Size(),
				"modified": info.ModTime(),
				"is_dir":   info.IsDir(),
				"mode":     info.Mode().String(),
			})

			return nil
		})
	} else {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory %s: %w", path, err)
		}

		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			files = append(files, map[string]interface{}{
				"path":     filepath.Join(path, entry.Name()),
				"name":     entry.Name(),
				"size":     info.Size(),
				"modified": info.ModTime(),
				"is_dir":   entry.IsDir(),
				"mode":     info.Mode().String(),
			})
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list directory %s: %w", path, err)
	}

	return map[string]interface{}{
		"path":      path,
		"files":     files,
		"count":     len(files),
		"recursive": recursive,
	}, nil
}

// deleteFile deletes a file or directory.
func (ft *FileTool) deleteFile(path string) (interface{}, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path %s: %w", path, err)
	}

	err = os.RemoveAll(path)
	if err != nil {
		return nil, fmt.Errorf("failed to delete %s: %w", path, err)
	}

	return map[string]interface{}{
		"path":    path,
		"was_dir": info.IsDir(),
		"message": "Deleted successfully",
	}, nil
}

// makeDirectory creates a directory.
func (ft *FileTool) makeDirectory(path string, recursive bool) (interface{}, error) {
	var err error
	if recursive {
		err = os.MkdirAll(path, 0755)
	} else {
		err = os.Mkdir(path, 0755)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", path, err)
	}

	return map[string]interface{}{
		"path":      path,
		"recursive": recursive,
		"message":   "Directory created successfully",
	}, nil
}
