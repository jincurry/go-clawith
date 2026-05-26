package repository

import (
	"github.com/google/uuid"
	"github.com/jincurry/go-clawith/internal/model"
	"gorm.io/gorm"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) CreateSession(session *model.ChatSession) error {
	return r.db.Create(session).Error
}

func (r *ChatRepository) GetSession(id uuid.UUID) (*model.ChatSession, error) {
	var session model.ChatSession
	err := r.db.Where("id = ?", id).First(&session).Error
	return &session, err
}

func (r *ChatRepository) ListSessions(userID uuid.UUID, offset, limit int) ([]model.ChatSession, int64, error) {
	var sessions []model.ChatSession
	var total int64

	query := r.db.Where("user_id = ?", userID)
	query.Model(&model.ChatSession{}).Count(&total)
	err := query.Offset(offset).Limit(limit).Order("updated_at DESC").Find(&sessions).Error
	return sessions, total, err
}

func (r *ChatRepository) DeleteSession(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("session_id = ?", id).Delete(&model.Message{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.ChatSession{}, "id = ?", id).Error
	})
}

func (r *ChatRepository) CreateMessage(msg *model.Message) error {
	return r.db.Create(msg).Error
}

func (r *ChatRepository) ListMessages(sessionID uuid.UUID, offset, limit int) ([]model.Message, error) {
	var messages []model.Message
	err := r.db.Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Offset(offset).Limit(limit).
		Find(&messages).Error
	return messages, err
}

func (r *ChatRepository) UpdateSessionTitle(id uuid.UUID, title string) error {
	return r.db.Model(&model.ChatSession{}).Where("id = ?", id).Update("title", title).Error
}
