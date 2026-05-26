package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Trigger struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	AgentID   uuid.UUID `gorm:"type:uuid;index;not null" json:"agent_id"`
	Name      string    `gorm:"not null" json:"name"`
	Type      string    `gorm:"not null" json:"type"` // cron, webhook
	Config    string    `gorm:"type:jsonb;not null" json:"config"`
	Action    string    `gorm:"type:text;not null" json:"action"` // prompt to execute
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	LastRunAt *time.Time `json:"last_run_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Agent      Agent              `gorm:"foreignKey:AgentID" json:"-"`
	Executions []TriggerExecution `gorm:"foreignKey:TriggerID" json:"executions,omitempty"`
}

func (t *Trigger) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

type TriggerExecution struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TriggerID uuid.UUID `gorm:"type:uuid;index;not null" json:"trigger_id"`
	Status    string    `gorm:"not null" json:"status"` // running, success, failed
	Output    string    `gorm:"type:text" json:"output,omitempty"`
	Error     string    `gorm:"type:text" json:"error,omitempty"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`

	Trigger Trigger `gorm:"foreignKey:TriggerID" json:"-"`
}

func (te *TriggerExecution) BeforeCreate(tx *gorm.DB) error {
	if te.ID == uuid.Nil {
		te.ID = uuid.New()
	}
	return nil
}
