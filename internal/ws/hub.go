package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Conn   *websocket.Conn
	Send   chan []byte
}

type Hub struct {
	mu         sync.RWMutex
	clients    map[uuid.UUID]*Client
	userConns  map[uuid.UUID]map[uuid.UUID]*Client // userID -> clientID -> client
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]*Client),
		userConns:  make(map[uuid.UUID]map[uuid.UUID]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			if _, ok := h.userConns[client.UserID]; !ok {
				h.userConns[client.UserID] = make(map[uuid.UUID]*Client)
			}
			h.userConns[client.UserID][client.ID] = client
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				if conns, ok := h.userConns[client.UserID]; ok {
					delete(conns, client.ID)
					if len(conns) == 0 {
						delete(h.userConns, client.UserID)
					}
				}
				close(client.Send)
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

type WSMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

func (h *Hub) SendToUser(userID uuid.UUID, msg WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("ws marshal error: %v", err)
		return
	}

	h.mu.RLock()
	conns, ok := h.userConns[userID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	for _, client := range conns {
		select {
		case client.Send <- data:
		default:
			go h.Unregister(client)
		}
	}
}
