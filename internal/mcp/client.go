package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// Client implements an MCP client using the Streamable HTTP transport.
type Client struct {
	endpoint  string
	client    *http.Client
	sessionID string
	idCounter atomic.Int64
	mu        sync.Mutex

	ServerInfo Info
	Tools      []ToolDefinition
}

func NewClient(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

// Connect performs the MCP initialize handshake and discovers tools.
func (c *Client) Connect(ctx context.Context) error {
	initResult, err := c.initialize(ctx)
	if err != nil {
		return fmt.Errorf("initialize failed: %w", err)
	}
	c.ServerInfo = initResult.ServerInfo

	// Send initialized notification
	if err := c.notify(ctx, "notifications/initialized", nil); err != nil {
		return fmt.Errorf("initialized notification failed: %w", err)
	}

	// Discover tools
	if initResult.Capabilities.Tools != nil {
		tools, err := c.ListTools(ctx)
		if err != nil {
			return fmt.Errorf("tools/list failed: %w", err)
		}
		c.Tools = tools
	}

	return nil
}

func (c *Client) initialize(ctx context.Context) (*InitializeResult, error) {
	params := InitializeParams{
		ProtocolVersion: "2025-03-26",
		Capabilities:    ClientCaps{},
		ClientInfo: Info{
			Name:    "go-clawith",
			Version: "1.0.0",
		},
	}

	resp, err := c.call(ctx, "initialize", params)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, err
	}

	var result InitializeResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ListTools calls tools/list on the MCP server.
func (c *Client) ListTools(ctx context.Context) ([]ToolDefinition, error) {
	resp, err := c.call(ctx, "tools/list", nil)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, err
	}

	var result ToolsListResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result.Tools, nil
}

// CallTool calls tools/call on the MCP server.
func (c *Client) CallTool(ctx context.Context, name string, arguments map[string]any) (*ToolCallResult, error) {
	params := ToolCallParams{
		Name:      name,
		Arguments: arguments,
	}

	resp, err := c.call(ctx, "tools/call", params)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, err
	}

	var result ToolCallResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) call(ctx context.Context, method string, params any) (*JSONRPCResponse, error) {
	id := int(c.idCounter.Add(1))
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	if c.sessionID != "" {
		httpReq.Header.Set("Mcp-Session-Id", c.sessionID)
	}

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Capture session ID from response
	if sid := httpResp.Header.Get("Mcp-Session-Id"); sid != "" {
		c.mu.Lock()
		c.sessionID = sid
		c.mu.Unlock()
	}

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("server returned %d: %s", httpResp.StatusCode, string(respBody))
	}

	var resp JSONRPCResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("rpc error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	return &resp, nil
}

func (c *Client) notify(ctx context.Context, method string, params any) error {
	req := struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  any    `json:"params,omitempty"`
	}{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	if c.sessionID != "" {
		httpReq.Header.Set("Mcp-Session-Id", c.sessionID)
	}

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()
	io.ReadAll(httpResp.Body)

	if sid := httpResp.Header.Get("Mcp-Session-Id"); sid != "" {
		c.mu.Lock()
		c.sessionID = sid
		c.mu.Unlock()
	}

	return nil
}
