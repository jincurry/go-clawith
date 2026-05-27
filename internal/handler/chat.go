package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jincurry/go-clawith/internal/config"
	"github.com/jincurry/go-clawith/internal/middleware"
	"github.com/jincurry/go-clawith/internal/service"
	"github.com/jincurry/go-clawith/internal/ws"
)

type ChatHandler struct {
	chatSvc *service.ChatService
	hub     *ws.Hub
	jwtCfg  config.JWTConfig
}

func NewChatHandler(chatSvc *service.ChatService, hub *ws.Hub, jwtCfg config.JWTConfig) *ChatHandler {
	return &ChatHandler{chatSvc: chatSvc, hub: hub, jwtCfg: jwtCfg}
}

func (h *ChatHandler) CreateSession(c *gin.Context) {
	var input service.CreateSessionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	session, err := h.chatSvc.CreateSession(userID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, session)
}

func (h *ChatHandler) ListSessions(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	sessions, total, err := h.chatSvc.ListSessions(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  sessions,
		"total": total,
		"page":  page,
	})
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	messages, err := h.chatSvc.GetMessages(sessionID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": messages})
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	var input service.SendMessageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg, err := h.chatSvc.SendMessage(c.Request.Context(), sessionID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, msg)
}

func (h *ChatHandler) DeleteSession(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	if err := h.chatSvc.DeleteSession(sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *ChatHandler) WebSocket(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	token, err := jwt.ParseWithClaims(tokenStr, &middleware.Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(h.jwtCfg.Secret), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	claims := token.Claims.(*middleware.Claims)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}

	client := &ws.Client{
		ID:     uuid.New(),
		UserID: claims.UserID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	h.hub.Register(client)

	go h.writePump(client)
	go h.readPump(client)
}

func (h *ChatHandler) writePump(client *ws.Client) {
	defer client.Conn.Close()
	for msg := range client.Send {
		if err := client.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}

func (h *ChatHandler) readPump(client *ws.Client) {
	defer func() {
		h.hub.Unregister(client)
		client.Conn.Close()
	}()

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}

		var wsMsg struct {
			Type      string `json:"type"`
			SessionID string `json:"session_id"`
			Content   string `json:"content"`
		}
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			continue
		}

		if wsMsg.Type == "chat" {
			sessionID, err := uuid.Parse(wsMsg.SessionID)
			if err != nil {
				continue
			}
			go h.handleStreamChat(client, sessionID, wsMsg.Content)
		}
	}
}

func (h *ChatHandler) handleStreamChat(client *ws.Client, sessionID uuid.UUID, content string) {
	ctx := context.Background()

	streamCh, err := h.chatSvc.StreamMessage(ctx, sessionID, content, h.hub, client.UserID)
	if err != nil {
		errMsg, _ := json.Marshal(ws.WSMessage{Type: "error", Payload: err.Error()})
		client.Send <- errMsg
		return
	}

	var fullContent string
	for chunk := range streamCh {
		if chunk.Error != nil {
			errMsg, _ := json.Marshal(ws.WSMessage{Type: "error", Payload: chunk.Error.Error()})
			client.Send <- errMsg
			return
		}
		if chunk.Done {
			doneMsg, _ := json.Marshal(ws.WSMessage{Type: "done", Payload: nil})
			client.Send <- doneMsg
			break
		}
		if chunk.Content != "" {
			fullContent += chunk.Content
			chunkMsg, _ := json.Marshal(ws.WSMessage{Type: "chunk", Payload: chunk.Content})
			client.Send <- chunkMsg
		}
	}

	h.chatSvc.SaveAssistantMessage(sessionID, fullContent, 0)
}
