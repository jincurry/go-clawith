package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Agent struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	Name         string    `gorm:"not null" json:"name"`
	Avatar       string    `json:"avatar,omitempty"`
	Description  string    `json:"description,omitempty"`
	SystemPrompt string    `gorm:"type:text" json:"system_prompt"`
	Model        string    `gorm:"not null" json:"model"`
	Provider     string    `gorm:"not null" json:"provider"` // openai, anthropic
	Temperature  float64   `gorm:"default:0.7" json:"temperature"`
	MaxTokens    int       `gorm:"default:4096" json:"max_tokens"`
	Memory       string    `gorm:"type:text" json:"memory,omitempty"`   // long-term memory (memory.md content)
	Persona      string    `gorm:"type:text" json:"persona,omitempty"` // personality (soul.md content)
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	User  User        `gorm:"foreignKey:UserID" json:"-"`
	Tools []AgentTool `gorm:"foreignKey:AgentID" json:"tools,omitempty"`
}

func (a *Agent) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

type AgentTool struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	AgentID uuid.UUID `gorm:"type:uuid;index;not null" json:"agent_id"`
	ToolID  uuid.UUID `gorm:"type:uuid;index;not null" json:"tool_id"`

	Agent Agent `gorm:"foreignKey:AgentID" json:"-"`
	Tool  Tool  `gorm:"foreignKey:ToolID" json:"-"`
}

func (at *AgentTool) BeforeCreate(tx *gorm.DB) error {
	if at.ID == uuid.Nil {
		at.ID = uuid.New()
	}
	return nil
}
