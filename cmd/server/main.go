package main

import (
	"log"

	"github.com/jincurry/go-clawith/internal/config"
	"github.com/jincurry/go-clawith/internal/database"
	"github.com/jincurry/go-clawith/internal/handler"
	"github.com/jincurry/go-clawith/internal/llm"
	"github.com/jincurry/go-clawith/internal/repository"
	"github.com/jincurry/go-clawith/internal/service"
	"github.com/jincurry/go-clawith/internal/trigger"
	"github.com/jincurry/go-clawith/internal/ws"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// Repositories
	userRepo := repository.NewUserRepository(db)
	agentRepo := repository.NewAgentRepository(db)
	chatRepo := repository.NewChatRepository(db)
	toolRepo := repository.NewToolRepository(db)
	triggerRepo := repository.NewTriggerRepository(db)

	// LLM
	llmRegistry := llm.NewRegistry(cfg.LLM)

	// Services
	authSvc := service.NewAuthService(userRepo, cfg.JWT)
	agentSvc := service.NewAgentService(agentRepo)
	chatSvc := service.NewChatService(chatRepo, agentRepo, llmRegistry)

	// WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()

	// Trigger Scheduler
	scheduler := trigger.NewScheduler(triggerRepo, agentRepo, llmRegistry)
	if err := scheduler.Start(); err != nil {
		log.Printf("warning: trigger scheduler start error: %v", err)
	}
	defer scheduler.Stop()

	// Handlers
	handlers := &handler.Handlers{
		Auth:    handler.NewAuthHandler(authSvc),
		Agent:   handler.NewAgentHandler(agentSvc),
		Chat:    handler.NewChatHandler(chatSvc, hub, cfg.JWT),
		Tool:    handler.NewToolHandler(toolRepo),
		Trigger: handler.NewTriggerHandler(triggerRepo, scheduler),
	}

	router := handler.SetupRouter(cfg, handlers)

	log.Printf("server starting on :%s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
