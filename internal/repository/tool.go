package repository

import (
	"github.com/google/uuid"
	"github.com/jincurry/go-clawith/internal/model"
	"gorm.io/gorm"
)

type ToolRepository struct {
	db *gorm.DB
}

func NewToolRepository(db *gorm.DB) *ToolRepository {
	return &ToolRepository{db: db}
}

func (r *ToolRepository) Create(tool *model.Tool) error {
	return r.db.Create(tool).Error
}

func (r *ToolRepository) GetByID(id uuid.UUID) (*model.Tool, error) {
	var tool model.Tool
	err := r.db.Where("id = ?", id).First(&tool).Error
	return &tool, err
}

func (r *ToolRepository) GetByName(name string) (*model.Tool, error) {
	var tool model.Tool
	err := r.db.Where("name = ?", name).First(&tool).Error
	return &tool, err
}

func (r *ToolRepository) List(offset, limit int) ([]model.Tool, int64, error) {
	var tools []model.Tool
	var total int64

	r.db.Model(&model.Tool{}).Count(&total)
	err := r.db.Offset(offset).Limit(limit).Order("name ASC").Find(&tools).Error
	return tools, total, err
}

func (r *ToolRepository) Update(tool *model.Tool) error {
	return r.db.Save(tool).Error
}

func (r *ToolRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Tool{}, "id = ?", id).Error
}
