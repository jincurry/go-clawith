package handler

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jincurry/go-clawith/internal/config"
	"github.com/jincurry/go-clawith/internal/middleware"
)

type Handlers struct {
	Auth    *AuthHandler
	Agent   *AgentHandler
	Chat    *ChatHandler
	Tool    *ToolHandler
	Trigger *TriggerHandler
}

func SetupRouter(cfg *config.Config, h *Handlers) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public routes
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
	}

	// WebSocket (authenticated via query param)
	r.GET("/ws", h.Chat.WebSocket)

	// Webhook triggers (public, authenticated by trigger ID)
	r.POST("/api/webhooks/:id", h.Trigger.Webhook)

	// Protected routes
	api := r.Group("/api")
	api.Use(middleware.AuthRequired(cfg.JWT))
	{
		api.GET("/auth/me", h.Auth.Me)

		// Agents
		agents := api.Group("/agents")
		{
			agents.POST("", h.Agent.Create)
			agents.GET("", h.Agent.List)
			agents.GET("/:id", h.Agent.Get)
			agents.PUT("/:id", h.Agent.Update)
			agents.DELETE("/:id", h.Agent.Delete)
		}

		// Chat
		chat := api.Group("/chat")
		{
			chat.POST("/sessions", h.Chat.CreateSession)
			chat.GET("/sessions", h.Chat.ListSessions)
			chat.GET("/sessions/:id/messages", h.Chat.GetMessages)
			chat.POST("/sessions/:id/messages", h.Chat.SendMessage)
			chat.DELETE("/sessions/:id", h.Chat.DeleteSession)
		}

		// Tools
		tools := api.Group("/tools")
		{
			tools.POST("", h.Tool.Create)
			tools.GET("", h.Tool.List)
			tools.GET("/:id", h.Tool.Get)
			tools.DELETE("/:id", h.Tool.Delete)
		}

		// Triggers
		triggers := api.Group("/triggers")
		{
			triggers.POST("", h.Trigger.Create)
			triggers.GET("/agent/:agent_id", h.Trigger.ListByAgent)
			triggers.DELETE("/:id", h.Trigger.Delete)
		}
	}

	return r
}
