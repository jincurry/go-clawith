package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jincurry/go-clawith/internal/model"
	"github.com/jincurry/go-clawith/internal/repository"
	"github.com/jincurry/go-clawith/internal/trigger"
)

type TriggerHandler struct {
	triggerRepo *repository.TriggerRepository
	scheduler   *trigger.Scheduler
}

func NewTriggerHandler(triggerRepo *repository.TriggerRepository, scheduler *trigger.Scheduler) *TriggerHandler {
	return &TriggerHandler{triggerRepo: triggerRepo, scheduler: scheduler}
}

type CreateTriggerInput struct {
	AgentID  string `json:"agent_id" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Type     string `json:"type" binding:"required"`
	Config   string `json:"config" binding:"required"`
	Action   string `json:"action" binding:"required"`
	IsActive bool   `json:"is_active"`
}

func (h *TriggerHandler) Create(c *gin.Context) {
	var input CreateTriggerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agentID, err := uuid.Parse(input.AgentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agent_id"})
		return
	}

	t := &model.Trigger{
		AgentID:  agentID,
		Name:     input.Name,
		Type:     input.Type,
		Config:   input.Config,
		Action:   input.Action,
		IsActive: input.IsActive,
	}

	if err := h.triggerRepo.Create(t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if t.IsActive {
		h.scheduler.AddTrigger(t)
	}

	c.JSON(http.StatusCreated, t)
}

func (h *TriggerHandler) ListByAgent(c *gin.Context) {
	agentID, err := uuid.Parse(c.Param("agent_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agent_id"})
		return
	}

	triggers, err := h.triggerRepo.ListByAgent(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": triggers})
}

func (h *TriggerHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	h.scheduler.RemoveTrigger(id)
	if err := h.triggerRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *TriggerHandler) Webhook(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid trigger id"})
		return
	}

	t, err := h.triggerRepo.GetByID(id)
	if err != nil || t.Type != "webhook" || !t.IsActive {
		c.JSON(http.StatusNotFound, gin.H{"error": "trigger not found"})
		return
	}

	go h.scheduler.ExecuteTrigger(t.ID)

	c.JSON(http.StatusOK, gin.H{"status": "triggered"})
}
