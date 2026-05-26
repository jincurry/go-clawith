package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jincurry/go-clawith/internal/model"
	"gorm.io/gorm"
)

type TriggerRepository struct {
	db *gorm.DB
}

func NewTriggerRepository(db *gorm.DB) *TriggerRepository {
	return &TriggerRepository{db: db}
}

func (r *TriggerRepository) Create(trigger *model.Trigger) error {
	return r.db.Create(trigger).Error
}

func (r *TriggerRepository) GetByID(id uuid.UUID) (*model.Trigger, error) {
	var trigger model.Trigger
	err := r.db.Where("id = ?", id).First(&trigger).Error
	return &trigger, err
}

func (r *TriggerRepository) ListByAgent(agentID uuid.UUID) ([]model.Trigger, error) {
	var triggers []model.Trigger
	err := r.db.Where("agent_id = ?", agentID).Order("created_at DESC").Find(&triggers).Error
	return triggers, err
}

func (r *TriggerRepository) ListActive() ([]model.Trigger, error) {
	var triggers []model.Trigger
	err := r.db.Where("is_active = ?", true).Find(&triggers).Error
	return triggers, err
}

func (r *TriggerRepository) Update(trigger *model.Trigger) error {
	return r.db.Save(trigger).Error
}

func (r *TriggerRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Trigger{}, "id = ?", id).Error
}

func (r *TriggerRepository) RecordExecution(exec *model.TriggerExecution) error {
	return r.db.Create(exec).Error
}

func (r *TriggerRepository) UpdateLastRun(id uuid.UUID, t time.Time) error {
	return r.db.Model(&model.Trigger{}).Where("id = ?", id).Update("last_run_at", t).Error
}
