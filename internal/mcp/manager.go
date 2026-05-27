package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

// MCPServer represents a configured MCP server connection.
type MCPServer interface {
	Connect(ctx context.Context) error
	ListTools(ctx context.Context) ([]ToolDefinition, error)
	CallTool(ctx context.Context, name string, arguments map[string]any) (*ToolCallResult, error)
}

// ServerConfig describes how to connect to an MCP server.
type ServerConfig struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Transport string `json:"transport"` // "http" or "stdio"

	// HTTP transport
	Endpoint string `json:"endpoint,omitempty"`

	// Stdio transport
	Command string   `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
	Env     []string `json:"env,omitempty"`
}

type serverEntry struct {
	Config ServerConfig
	Server MCPServer
	Tools  []ToolDefinition
}

// Manager manages multiple MCP server connections and provides
// a unified interface for tool discovery and execution.
type Manager struct {
	mu      sync.RWMutex
	servers map[string]*serverEntry     // serverID -> entry
	toolMap map[string]*serverEntry     // toolName -> owning server entry
}

func NewManager() *Manager {
	return &Manager{
		servers: make(map[string]*serverEntry),
		toolMap: make(map[string]*serverEntry),
	}
}

// AddServer connects to an MCP server and discovers its tools.
func (m *Manager) AddServer(ctx context.Context, cfg ServerConfig) error {
	var server MCPServer

	switch cfg.Transport {
	case "http":
		if cfg.Endpoint == "" {
			return fmt.Errorf("http transport requires endpoint")
		}
		server = NewClient(cfg.Endpoint)
	case "stdio":
		if cfg.Command == "" {
			return fmt.Errorf("stdio transport requires command")
		}
		server = NewStdioClient(cfg.Command, cfg.Args, cfg.Env)
	default:
		return fmt.Errorf("unsupported transport: %s", cfg.Transport)
	}

	if err := server.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to %s: %w", cfg.Name, err)
	}

	tools, err := server.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tools from %s: %w", cfg.Name, err)
	}

	entry := &serverEntry{
		Config: cfg,
		Server: server,
		Tools:  tools,
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.servers[cfg.ID] = entry
	for _, tool := range tools {
		m.toolMap[tool.Name] = entry
	}

	log.Printf("MCP server %q connected: %d tools discovered", cfg.Name, len(tools))
	return nil
}

// RemoveServer disconnects from an MCP server.
func (m *Manager) RemoveServer(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.servers[id]
	if !ok {
		return
	}

	// Remove tools from map
	for _, tool := range entry.Tools {
		delete(m.toolMap, tool.Name)
	}

	// Close stdio clients
	if sc, ok := entry.Server.(*StdioClient); ok {
		sc.Close()
	}

	delete(m.servers, id)
	log.Printf("MCP server %q disconnected", entry.Config.Name)
}

// CallTool finds the server that owns the tool and executes it.
func (m *Manager) CallTool(ctx context.Context, toolName string, arguments string) (string, error) {
	m.mu.RLock()
	entry, ok := m.toolMap[toolName]
	m.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("MCP tool %q not found on any connected server", toolName)
	}

	var args map[string]any
	if arguments != "" {
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return "", fmt.Errorf("failed to parse tool arguments: %w", err)
		}
	}

	result, err := entry.Server.CallTool(ctx, toolName, args)
	if err != nil {
		return "", err
	}

	if result.IsError {
		var text string
		for _, block := range result.Content {
			if block.Type == "text" {
				text += block.Text
			}
		}
		return "", fmt.Errorf("tool error: %s", text)
	}

	var output string
	for _, block := range result.Content {
		if block.Type == "text" {
			output += block.Text
		}
	}

	return output, nil
}

// GetAllTools returns all tools from all connected servers.
func (m *Manager) GetAllTools() []ToolDefinition {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var all []ToolDefinition
	for _, entry := range m.servers {
		all = append(all, entry.Tools...)
	}
	return all
}

// GetServerTools returns tools from a specific server.
func (m *Manager) GetServerTools(id string) []ToolDefinition {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, ok := m.servers[id]
	if !ok {
		return nil
	}
	return entry.Tools
}

// HasTool checks if a tool is available on any connected server.
func (m *Manager) HasTool(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.toolMap[name]
	return ok
}

// ListServers returns all connected server configs.
func (m *Manager) ListServers() []ServerConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	configs := make([]ServerConfig, 0, len(m.servers))
	for _, entry := range m.servers {
		configs = append(configs, entry.Config)
	}
	return configs
}
