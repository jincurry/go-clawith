package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Workspace struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	AgentID   uuid.UUID `gorm:"type:uuid;index;not null" json:"agent_id"`
	Name      string    `gorm:"not null" json:"name"`
	Path      string    `gorm:"not null" json:"path"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Agent Agent          `gorm:"foreignKey:AgentID" json:"-"`
	Files []WorkspaceFile `gorm:"foreignKey:WorkspaceID" json:"files,omitempty"`
}

func (w *Workspace) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

type WorkspaceFile struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	WorkspaceID uuid.UUID `gorm:"type:uuid;index;not null" json:"workspace_id"`
	Name        string    `gorm:"not null" json:"name"`
	Path        string    `gorm:"not null" json:"path"`
	Size        int64     `json:"size"`
	MimeType    string    `json:"mime_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (wf *WorkspaceFile) BeforeCreate(tx *gorm.DB) error {
	if wf.ID == uuid.Nil {
		wf.ID = uuid.New()
	}
	return nil
}
