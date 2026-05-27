package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jincurry/go-clawith/internal/mcp"
	"github.com/jincurry/go-clawith/internal/model"
)

type ExecutionResult struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

type Executor struct {
	client     *http.Client
	mcpManager *mcp.Manager
}

func NewExecutor(mcpManager *mcp.Manager) *Executor {
	return &Executor{
		client:     &http.Client{Timeout: 30 * time.Second},
		mcpManager: mcpManager,
	}
}

func (e *Executor) Execute(ctx context.Context, t *model.Tool, arguments string) (*ExecutionResult, error) {
	switch t.Type {
	case "builtin":
		return e.executeBuiltin(ctx, t, arguments)
	case "custom":
		return e.executeHTTP(ctx, t, arguments)
	case "mcp":
		return e.executeMCP(ctx, t, arguments)
	default:
		return nil, fmt.Errorf("unsupported tool type: %s", t.Type)
	}
}

func (e *Executor) executeBuiltin(_ context.Context, t *model.Tool, arguments string) (*ExecutionResult, error) {
	switch t.Name {
	case "current_time":
		return &ExecutionResult{Output: time.Now().Format(time.RFC3339)}, nil
	case "echo":
		var args struct {
			Message string `json:"message"`
		}
		json.Unmarshal([]byte(arguments), &args)
		return &ExecutionResult{Output: args.Message}, nil
	default:
		return &ExecutionResult{
			Output: fmt.Sprintf("builtin tool %q executed with args: %s", t.Name, arguments),
		}, nil
	}
}

func (e *Executor) executeHTTP(ctx context.Context, t *model.Tool, arguments string) (*ExecutionResult, error) {
	if t.Endpoint == "" {
		return nil, fmt.Errorf("tool %q has no endpoint configured", t.Name)
	}

	body := map[string]any{
		"tool":      t.Name,
		"arguments": json.RawMessage(arguments),
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.Endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	if t.Config != "" {
		var cfg struct {
			Headers map[string]string `json:"headers"`
		}
		if json.Unmarshal([]byte(t.Config), &cfg) == nil && cfg.Headers != nil {
			for k, v := range cfg.Headers {
				req.Header.Set(k, v)
			}
		}
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return &ExecutionResult{Error: fmt.Sprintf("request failed: %v", err)}, nil
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ExecutionResult{Error: fmt.Sprintf("failed to read response: %v", err)}, nil
	}

	if resp.StatusCode >= 400 {
		return &ExecutionResult{Error: fmt.Sprintf("tool returned %d: %s", resp.StatusCode, string(respBody))}, nil
	}

	return &ExecutionResult{Output: string(respBody)}, nil
}

func (e *Executor) executeMCP(ctx context.Context, t *model.Tool, arguments string) (*ExecutionResult, error) {
	if e.mcpManager == nil {
		return &ExecutionResult{Error: "MCP manager not configured"}, nil
	}

	output, err := e.mcpManager.CallTool(ctx, t.Name, arguments)
	if err != nil {
		return &ExecutionResult{Error: fmt.Sprintf("MCP call failed: %v", err)}, nil
	}

	return &ExecutionResult{Output: output}, nil
}

// MCPManager returns the MCP manager for external access (e.g., tool sync).
func (e *Executor) MCPManager() *mcp.Manager {
	return e.mcpManager
}
