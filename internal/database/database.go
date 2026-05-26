package database

import (
	"fmt"
	"log"

	"github.com/jincurry/go-clawith/internal/config"
	"github.com/jincurry/go-clawith/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	logLevel := logger.Info
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	log.Println("Running auto migration...")
	return db.AutoMigrate(
		&model.User{},
		&model.Agent{},
		&model.AgentTool{},
		&model.ChatSession{},
		&model.Message{},
		&model.Tool{},
		&model.Trigger{},
		&model.TriggerExecution{},
		&model.Workspace{},
		&model.WorkspaceFile{},
	)
}
