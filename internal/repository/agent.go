package repository

import (
	"github.com/google/uuid"
	"github.com/jincurry/go-clawith/internal/model"
	"gorm.io/gorm"
)

type AgentRepository struct {
	db *gorm.DB
}

func NewAgentRepository(db *gorm.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

func (r *AgentRepository) Create(agent *model.Agent) error {
	return r.db.Create(agent).Error
}

func (r *AgentRepository) GetByID(id uuid.UUID) (*model.Agent, error) {
	var agent model.Agent
	err := r.db.Preload("Tools.Tool").Where("id = ?", id).First(&agent).Error
	return &agent, err
}

func (r *AgentRepository) ListByUser(userID uuid.UUID, offset, limit int) ([]model.Agent, int64, error) {
	var agents []model.Agent
	var total int64

	query := r.db.Where("user_id = ?", userID)
	query.Model(&model.Agent{}).Count(&total)
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&agents).Error
	return agents, total, err
}

func (r *AgentRepository) Update(agent *model.Agent) error {
	return r.db.Save(agent).Error
}

func (r *AgentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Agent{}, "id = ?", id).Error
}

func (r *AgentRepository) SetTools(agentID uuid.UUID, toolIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("agent_id = ?", agentID).Delete(&model.AgentTool{}).Error; err != nil {
			return err
		}
		for _, toolID := range toolIDs {
			at := model.AgentTool{AgentID: agentID, ToolID: toolID}
			if err := tx.Create(&at).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
