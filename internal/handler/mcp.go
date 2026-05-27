package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jincurry/go-clawith/internal/mcp"
	"github.com/jincurry/go-clawith/internal/model"
	"github.com/jincurry/go-clawith/internal/repository"
)

type MCPHandler struct {
	manager  *mcp.Manager
	toolRepo *repository.ToolRepository
}

func NewMCPHandler(manager *mcp.Manager, toolRepo *repository.ToolRepository) *MCPHandler {
	return &MCPHandler{manager: manager, toolRepo: toolRepo}
}

type AddMCPServerInput struct {
	Name      string   `json:"name" binding:"required"`
	Transport string   `json:"transport" binding:"required"`
	Endpoint  string   `json:"endpoint,omitempty"`
	Command   string   `json:"command,omitempty"`
	Args      []string `json:"args,omitempty"`
	Env       []string `json:"env,omitempty"`
}

// Connect adds and connects to a new MCP server, then syncs its tools.
func (h *MCPHandler) Connect(c *gin.Context) {
	var input AddMCPServerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg := mcp.ServerConfig{
		ID:        uuid.New().String(),
		Name:      input.Name,
		Transport: input.Transport,
		Endpoint:  input.Endpoint,
		Command:   input.Command,
		Args:      input.Args,
		Env:       input.Env,
	}

	if err := h.manager.AddServer(c.Request.Context(), cfg); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	// Sync discovered tools into the tools table
	tools := h.manager.GetServerTools(cfg.ID)
	synced := h.syncTools(cfg, tools)

	c.JSON(http.StatusCreated, gin.H{
		"server_id":    cfg.ID,
		"name":         cfg.Name,
		"tools_synced": synced,
		"tools":        tools,
	})
}

// Disconnect removes an MCP server and its synced tools.
func (h *MCPHandler) Disconnect(c *gin.Context) {
	id := c.Param("id")

	tools := h.manager.GetServerTools(id)
	for _, t := range tools {
		existing, err := h.toolRepo.GetByName(t.Name)
		if err == nil {
			h.toolRepo.Delete(existing.ID)
		}
	}

	h.manager.RemoveServer(id)
	c.JSON(http.StatusOK, gin.H{"status": "disconnected"})
}

// ListServers lists all connected MCP servers.
func (h *MCPHandler) ListServers(c *gin.Context) {
	servers := h.manager.ListServers()
	result := make([]gin.H, 0, len(servers))

	for _, srv := range servers {
		tools := h.manager.GetServerTools(srv.ID)
		toolNames := make([]string, 0, len(tools))
		for _, t := range tools {
			toolNames = append(toolNames, t.Name)
		}
		result = append(result, gin.H{
			"id":        srv.ID,
			"name":      srv.Name,
			"transport": srv.Transport,
			"endpoint":  srv.Endpoint,
			"command":   srv.Command,
			"tools":     toolNames,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// RefreshTools re-discovers tools from a specific server.
func (h *MCPHandler) RefreshTools(c *gin.Context) {
	id := c.Param("id")

	servers := h.manager.ListServers()
	var found *mcp.ServerConfig
	for _, s := range servers {
		if s.ID == id {
			found = &s
			break
		}
	}
	if found == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}

	tools := h.manager.GetServerTools(id)
	synced := h.syncTools(*found, tools)

	c.JSON(http.StatusOK, gin.H{
		"tools_synced": synced,
		"tools":        tools,
	})
}

func (h *MCPHandler) syncTools(cfg mcp.ServerConfig, tools []mcp.ToolDefinition) int {
	synced := 0
	for _, t := range tools {
		existing, _ := h.toolRepo.GetByName(t.Name)
		if existing != nil {
			continue
		}

		schema, _ := json.Marshal(t.InputSchema)

		mcpConfig, _ := json.Marshal(map[string]string{
			"server_id": cfg.ID,
			"server":    cfg.Name,
		})

		tool := &model.Tool{
			Name:        t.Name,
			DisplayName: t.Name,
			Description: t.Description,
			Type:        "mcp",
			Schema:      string(schema),
			Config:      string(mcpConfig),
		}

		if err := h.toolRepo.Create(tool); err == nil {
			synced++
		}
	}
	return synced
}
