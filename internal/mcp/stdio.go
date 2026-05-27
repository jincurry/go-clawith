package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
)

// StdioClient implements an MCP client using the stdio transport.
// It spawns a subprocess and communicates via stdin/stdout.
type StdioClient struct {
	command   string
	args      []string
	env       []string
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    *bufio.Scanner
	idCounter atomic.Int64
	mu        sync.Mutex
	pending   map[int]chan *JSONRPCResponse

	ServerInfo Info
	Tools      []ToolDefinition
}

func NewStdioClient(command string, args []string, env []string) *StdioClient {
	return &StdioClient{
		command: command,
		args:    args,
		env:     env,
		pending: make(map[int]chan *JSONRPCResponse),
	}
}

// Connect starts the subprocess, performs MCP handshake, and discovers tools.
func (c *StdioClient) Connect(ctx context.Context) error {
	c.cmd = exec.CommandContext(ctx, c.command, c.args...)
	if len(c.env) > 0 {
		c.cmd.Env = c.env
	}

	var err error
	c.stdin, err = c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdoutPipe, err := c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	c.stdout = bufio.NewScanner(stdoutPipe)

	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	go c.readLoop()

	// Initialize
	initResult, err := c.initialize(ctx)
	if err != nil {
		c.Close()
		return fmt.Errorf("initialize failed: %w", err)
	}
	c.ServerInfo = initResult.ServerInfo

	// Notify initialized
	c.sendNotification("notifications/initialized", nil)

	// Discover tools
	if initResult.Capabilities.Tools != nil {
		tools, err := c.ListTools(ctx)
		if err != nil {
			c.Close()
			return fmt.Errorf("tools/list failed: %w", err)
		}
		c.Tools = tools
	}

	return nil
}

func (c *StdioClient) Close() error {
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}
	return nil
}

func (c *StdioClient) initialize(ctx context.Context) (*InitializeResult, error) {
	resp, err := c.call(ctx, "initialize", InitializeParams{
		ProtocolVersion: "2025-03-26",
		Capabilities:    ClientCaps{},
		ClientInfo:      Info{Name: "go-clawith", Version: "1.0.0"},
	})
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(resp.Result)
	var result InitializeResult
	json.Unmarshal(data, &result)
	return &result, nil
}

func (c *StdioClient) ListTools(ctx context.Context) ([]ToolDefinition, error) {
	resp, err := c.call(ctx, "tools/list", nil)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(resp.Result)
	var result ToolsListResult
	json.Unmarshal(data, &result)
	return result.Tools, nil
}

func (c *StdioClient) CallTool(ctx context.Context, name string, arguments map[string]any) (*ToolCallResult, error) {
	resp, err := c.call(ctx, "tools/call", ToolCallParams{Name: name, Arguments: arguments})
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(resp.Result)
	var result ToolCallResult
	json.Unmarshal(data, &result)
	return &result, nil
}

func (c *StdioClient) call(ctx context.Context, method string, params any) (*JSONRPCResponse, error) {
	id := int(c.idCounter.Add(1))

	ch := make(chan *JSONRPCResponse, 1)
	c.mu.Lock()
	c.pending[id] = ch
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
	}()

	req := JSONRPCRequest{JSONRPC: "2.0", ID: id, Method: method, Params: params}
	if err := c.send(req); err != nil {
		return nil, err
	}

	select {
	case resp := <-ch:
		if resp.Error != nil {
			return nil, fmt.Errorf("rpc error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *StdioClient) send(msg any) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	c.mu.Lock()
	defer c.mu.Unlock()
	_, err = c.stdin.Write(data)
	return err
}

func (c *StdioClient) sendNotification(method string, params any) {
	msg := struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  any    `json:"params,omitempty"`
	}{JSONRPC: "2.0", Method: method, Params: params}
	c.send(msg)
}

func (c *StdioClient) readLoop() {
	for c.stdout.Scan() {
		line := c.stdout.Bytes()
		var resp JSONRPCResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			continue
		}
		// Only process responses (have ID), skip notifications
		if resp.ID == 0 && resp.Result == nil && resp.Error == nil {
			continue
		}

		c.mu.Lock()
		ch, ok := c.pending[resp.ID]
		c.mu.Unlock()

		if ok {
			ch <- &resp
		}
	}
}
