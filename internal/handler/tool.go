package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jincurry/go-clawith/internal/model"
	"github.com/jincurry/go-clawith/internal/repository"
)

type ToolHandler struct {
	toolRepo *repository.ToolRepository
}

func NewToolHandler(toolRepo *repository.ToolRepository) *ToolHandler {
	return &ToolHandler{toolRepo: toolRepo}
}

type CreateToolInput struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Type        string `json:"type" binding:"required"`
	Schema      string `json:"schema"`
	Config      string `json:"config"`
	Endpoint    string `json:"endpoint"`
}

func (h *ToolHandler) Create(c *gin.Context) {
	var input CreateToolInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tool := &model.Tool{
		Name:        input.Name,
		DisplayName: input.DisplayName,
		Description: input.Description,
		Type:        input.Type,
		Schema:      input.Schema,
		Config:      input.Config,
		Endpoint:    input.Endpoint,
	}

	if err := h.toolRepo.Create(tool); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tool)
}

func (h *ToolHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	tool, err := h.toolRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tool not found"})
		return
	}

	c.JSON(http.StatusOK, tool)
}

func (h *ToolHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	tools, total, err := h.toolRepo.List(offset, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  tools,
		"total": total,
		"page":  page,
	})
}

func (h *ToolHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.toolRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
