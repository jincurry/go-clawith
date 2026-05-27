package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jincurry/go-clawith/internal/middleware"
	"github.com/jincurry/go-clawith/internal/model"
	"gorm.io/gorm"
)

const workspaceRoot = "./data/workspaces"

type WorkspaceHandler struct {
	db *gorm.DB
}

func NewWorkspaceHandler(db *gorm.DB) *WorkspaceHandler {
	os.MkdirAll(workspaceRoot, 0755)
	return &WorkspaceHandler{db: db}
}

func (h *WorkspaceHandler) Create(c *gin.Context) {
	var input struct {
		AgentID string `json:"agent_id" binding:"required"`
		Name    string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agentID, err := uuid.Parse(input.AgentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agent_id"})
		return
	}

	ws := &model.Workspace{
		AgentID: agentID,
		Name:    input.Name,
	}
	ws.ID = uuid.New()
	ws.Path = filepath.Join(workspaceRoot, ws.ID.String())

	if err := os.MkdirAll(ws.Path, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create workspace directory"})
		return
	}

	if err := h.db.Create(ws).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ws)
}

func (h *WorkspaceHandler) List(c *gin.Context) {
	agentID := c.Query("agent_id")
	var workspaces []model.Workspace

	query := h.db
	if agentID != "" {
		query = query.Where("agent_id = ?", agentID)
	}

	if err := query.Order("created_at DESC").Find(&workspaces).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": workspaces})
}

func (h *WorkspaceHandler) Upload(c *gin.Context) {
	wsID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace id"})
		return
	}

	var ws model.Workspace
	if err := h.db.Where("id = ?", wsID).First(&ws).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
		return
	}
	defer file.Close()

	safeName := filepath.Base(header.Filename)
	if strings.Contains(safeName, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
		return
	}

	destPath := filepath.Join(ws.Path, safeName)
	dest, err := os.Create(destPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create file"})
		return
	}
	defer dest.Close()

	written, err := io.Copy(dest, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	wf := &model.WorkspaceFile{
		WorkspaceID: wsID,
		Name:        safeName,
		Path:        destPath,
		Size:        written,
		MimeType:    header.Header.Get("Content-Type"),
	}
	h.db.Create(wf)

	c.JSON(http.StatusCreated, wf)
}

func (h *WorkspaceHandler) ListFiles(c *gin.Context) {
	wsID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace id"})
		return
	}

	var files []model.WorkspaceFile
	h.db.Where("workspace_id = ?", wsID).Order("created_at DESC").Find(&files)

	c.JSON(http.StatusOK, gin.H{"data": files})
}

func (h *WorkspaceHandler) DeleteFile(c *gin.Context) {
	fileID, err := uuid.Parse(c.Param("file_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}

	var wf model.WorkspaceFile
	if err := h.db.Where("id = ?", fileID).First(&wf).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	os.Remove(wf.Path)
	h.db.Delete(&wf)

	c.JSON(http.StatusNoContent, nil)
}

func (h *WorkspaceHandler) Download(c *gin.Context) {
	fileID, err := uuid.Parse(c.Param("file_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}

	var wf model.WorkspaceFile
	if err := h.db.Where("id = ?", fileID).First(&wf).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", wf.Name))
	c.File(wf.Path)
}

// Ensure user owns the workspace's agent (used as middleware check if needed)
func (h *WorkspaceHandler) checkOwnership(c *gin.Context, wsID uuid.UUID) bool {
	userID := middleware.GetUserID(c)
	var ws model.Workspace
	if err := h.db.Preload("Agent").Where("id = ?", wsID).First(&ws).Error; err != nil {
		return false
	}
	return ws.Agent.UserID == userID
}
