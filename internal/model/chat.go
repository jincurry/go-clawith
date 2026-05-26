package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatSession struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	AgentID   uuid.UUID `gorm:"type:uuid;index;not null" json:"agent_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User     User      `gorm:"foreignKey:UserID" json:"-"`
	Agent    Agent     `gorm:"foreignKey:AgentID" json:"-"`
	Messages []Message `gorm:"foreignKey:SessionID" json:"messages,omitempty"`
}

func (cs *ChatSession) BeforeCreate(tx *gorm.DB) error {
	if cs.ID == uuid.Nil {
		cs.ID = uuid.New()
	}
	return nil
}

type Message struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SessionID  uuid.UUID `gorm:"type:uuid;index;not null" json:"session_id"`
	Role       string    `gorm:"not null" json:"role"` // user, assistant, system, tool
	Content    string    `gorm:"type:text;not null" json:"content"`
	ToolCallID string    `json:"tool_call_id,omitempty"`
	ToolName   string    `json:"tool_name,omitempty"`
	TokenCount int       `json:"token_count,omitempty"`
	CreatedAt  time.Time `json:"created_at"`

	Session ChatSession `gorm:"foreignKey:SessionID" json:"-"`
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
