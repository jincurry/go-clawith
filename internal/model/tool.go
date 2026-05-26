package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Tool struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	Type        string    `gorm:"not null" json:"type"` // builtin, mcp, custom
	Schema      string    `gorm:"type:jsonb" json:"schema"`
	Config      string    `gorm:"type:jsonb" json:"config,omitempty"`
	Endpoint    string    `json:"endpoint,omitempty"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (t *Tool) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
