package vscode

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// FileOperation represents a requested change
type FileOperation struct {
	Type    string `json:"type"` // READ, WRITE, LIST, DELETE, PATCH
	Path    string `json:"path"`
	Content string `json:"content,omitempty"`
}

// Bridge handles communication between Talos Swarm and VS Code local FS
type Bridge struct {
	WorkspaceRoot string
	AllowedPaths  []string
}

// NewBridge creates a new FS bridge
func NewBridge(root string) *Bridge {
	return &Bridge{
		WorkspaceRoot: root,
		AllowedPaths:  []string{root}, // Default to workspace root
	}
}

// ExecuteOperation runs a filesystem operation safely
func (b *Bridge) ExecuteOperation(ctx context.Context, op FileOperation) (interface{}, error) {
	// Security check: Ensure path is within workspace
	cleanPath := filepath.Clean(filepath.Join(b.WorkspaceRoot, op.Path))
	if !strings.HasPrefix(cleanPath, b.WorkspaceRoot) {
		return nil, fmt.Errorf("access denied: path outside workspace")
	}

	switch op.Type {
	case "READ":
		data, err := ioutil.ReadFile(cleanPath)
		if err != nil {
			return nil, err
		}
		return string(data), nil

	case "WRITE":
		// Ensure dir exists
		if err := os.MkdirAll(filepath.Dir(cleanPath), 0755); err != nil {
			return nil, err
		}
		if err := ioutil.WriteFile(cleanPath, []byte(op.Content), 0644); err != nil {
			return nil, err
		}
		return "success", nil

	case "LIST":
		files, err := ioutil.ReadDir(cleanPath)
		if err != nil {
			return nil, err
		}
		var names []string
		for _, f := range files {
			names = append(names, f.Name())
		}
		return names, nil

	case "DELETE":
		if err := os.Remove(cleanPath); err != nil {
			return nil, err
		}
		return "success", nil

	default:
		return nil, fmt.Errorf("unknown operation: %s", op.Type)
	}
}

// GenerateCodeMap scans the workspace and creates a map of the codebase
func (b *Bridge) GenerateCodeMap(ctx context.Context) (map[string]interface{}, error) {
	codeMap := make(map[string]interface{})

	err := filepath.Walk(b.WorkspaceRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip git and hidden files
		if strings.Contains(path, ".git") || strings.Contains(path, "node_modules") {
			return filepath.SkipDir
		}

		if !info.IsDir() {
			relPath, _ := filepath.Rel(b.WorkspaceRoot, path)
			codeMap[relPath] = map[string]interface{}{
				"size": info.Size(),
				"mod":  info.ModTime(),
			}
		}
		return nil
	})

	return codeMap, err
}
