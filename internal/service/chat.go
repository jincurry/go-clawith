package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jincurry/go-clawith/internal/llm"
	"github.com/jincurry/go-clawith/internal/model"
	"github.com/jincurry/go-clawith/internal/repository"
	"github.com/jincurry/go-clawith/internal/ws"
)

type ChatService struct {
	chatRepo    *repository.ChatRepository
	agentRepo   *repository.AgentRepository
	agentRunner *AgentRunner
	llmReg      *llm.Registry
}

func NewChatService(
	chatRepo *repository.ChatRepository,
	agentRepo *repository.AgentRepository,
	agentRunner *AgentRunner,
	llmReg *llm.Registry,
) *ChatService {
	return &ChatService{
		chatRepo:    chatRepo,
		agentRepo:   agentRepo,
		agentRunner: agentRunner,
		llmReg:      llmReg,
	}
}

type CreateSessionInput struct {
	AgentID string `json:"agent_id" binding:"required"`
	Title   string `json:"title"`
}

type SendMessageInput struct {
	Content string `json:"content" binding:"required"`
}

func (s *ChatService) CreateSession(userID uuid.UUID, input CreateSessionInput) (*model.ChatSession, error) {
	agentID, err := uuid.Parse(input.AgentID)
	if err != nil {
		return nil, fmt.Errorf("invalid agent_id")
	}

	session := &model.ChatSession{
		UserID:  userID,
		AgentID: agentID,
		Title:   input.Title,
	}
	if session.Title == "" {
		session.Title = "New Chat"
	}

	if err := s.chatRepo.CreateSession(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *ChatService) GetSession(id uuid.UUID) (*model.ChatSession, error) {
	return s.chatRepo.GetSession(id)
}

func (s *ChatService) ListSessions(userID uuid.UUID, page, pageSize int) ([]model.ChatSession, int64, error) {
	offset := (page - 1) * pageSize
	return s.chatRepo.ListSessions(userID, offset, pageSize)
}

func (s *ChatService) DeleteSession(id uuid.UUID) error {
	return s.chatRepo.DeleteSession(id)
}

func (s *ChatService) GetMessages(sessionID uuid.UUID, page, pageSize int) ([]model.Message, error) {
	offset := (page - 1) * pageSize
	return s.chatRepo.ListMessages(sessionID, offset, pageSize)
}

// SendMessage handles synchronous chat with full tool-calling loop.
func (s *ChatService) SendMessage(ctx context.Context, sessionID uuid.UUID, input SendMessageInput) (*model.Message, error) {
	session, err := s.chatRepo.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	userMsg := &model.Message{
		SessionID: sessionID,
		Role:      "user",
		Content:   input.Content,
	}
	if err := s.chatRepo.CreateMessage(userMsg); err != nil {
		return nil, err
	}

	agent, err := s.agentRepo.GetByID(session.AgentID)
	if err != nil {
		return nil, err
	}

	history, err := s.chatRepo.ListMessages(sessionID, 0, 100)
	if err != nil {
		return nil, err
	}

	messages := s.buildLLMMessages(agent, history)

	result, err := s.agentRunner.Run(ctx, agent, messages)
	if err != nil {
		return nil, fmt.Errorf("agent error: %w", err)
	}

	// Save tool call trace messages
	s.saveTraceMessages(sessionID, result.Messages)

	assistantMsg := &model.Message{
		SessionID:  sessionID,
		Role:       "assistant",
		Content:    result.Content,
		TokenCount: result.TokensUsed,
	}
	if err := s.chatRepo.CreateMessage(assistantMsg); err != nil {
		return nil, err
	}

	return assistantMsg, nil
}

// StreamMessage handles WebSocket streaming with tool-calling loop.
// Tool calls are resolved first (non-streaming), then the final response is streamed.
// Tool call events are pushed to the client via hub.
func (s *ChatService) StreamMessage(ctx context.Context, sessionID uuid.UUID, content string, hub *ws.Hub, userID uuid.UUID) (<-chan llm.StreamChunk, error) {
	session, err := s.chatRepo.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	userMsg := &model.Message{
		SessionID: sessionID,
		Role:      "user",
		Content:   content,
	}
	if err := s.chatRepo.CreateMessage(userMsg); err != nil {
		return nil, err
	}

	agent, err := s.agentRepo.GetByID(session.AgentID)
	if err != nil {
		return nil, err
	}

	history, err := s.chatRepo.ListMessages(sessionID, 0, 100)
	if err != nil {
		return nil, err
	}

	messages := s.buildLLMMessages(agent, history)

	stream, traceMessages, err := s.agentRunner.StreamWithTools(ctx, agent, messages)
	if err != nil {
		return nil, err
	}

	// Notify client about tool calls that happened
	for _, tm := range traceMessages {
		if tm.Role == "assistant" && len(tm.ToolCalls) > 0 {
			for _, tc := range tm.ToolCalls {
				hub.SendToUser(userID, ws.WSMessage{
					Type: "tool_call",
					Payload: map[string]string{
						"name":      tc.Name,
						"arguments": tc.Arguments,
					},
				})
			}
		}
		if tm.Role == "tool" {
			hub.SendToUser(userID, ws.WSMessage{
				Type: "tool_result",
				Payload: map[string]string{
					"tool_call_id": tm.ToolCallID,
					"content":      truncate(tm.Content, 500),
				},
			})
		}
	}

	// Save trace messages
	s.saveTraceMessages(sessionID, traceMessages)

	return stream, nil
}

func (s *ChatService) SaveAssistantMessage(sessionID uuid.UUID, content string, tokenCount int) error {
	msg := &model.Message{
		SessionID:  sessionID,
		Role:       "assistant",
		Content:    content,
		TokenCount: tokenCount,
	}
	return s.chatRepo.CreateMessage(msg)
}

func (s *ChatService) buildLLMMessages(agent *model.Agent, history []model.Message) []llm.Message {
	msgs := make([]llm.Message, 0, len(history)+1)

	systemPrompt := agent.SystemPrompt
	if agent.Persona != "" {
		systemPrompt = agent.Persona + "\n\n" + systemPrompt
	}
	if agent.Memory != "" {
		systemPrompt += "\n\n## Memory\n" + agent.Memory
	}

	if systemPrompt != "" {
		msgs = append(msgs, llm.Message{Role: "system", Content: systemPrompt})
	}

	for _, m := range history {
		msg := llm.Message{
			Role:       m.Role,
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		}
		if m.Role == "assistant" && m.ToolName != "" {
			// Reconstruct tool calls from saved messages
			var toolCalls []llm.ToolCall
			json.Unmarshal([]byte(m.ToolName), &toolCalls)
			msg.ToolCalls = toolCalls
		}
		msgs = append(msgs, msg)
	}

	return msgs
}

func (s *ChatService) saveTraceMessages(sessionID uuid.UUID, messages []llm.Message) {
	for _, m := range messages {
		msg := &model.Message{
			SessionID:  sessionID,
			Role:       m.Role,
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		}
		if len(m.ToolCalls) > 0 {
			data, _ := json.Marshal(m.ToolCalls)
			msg.ToolName = string(data)
		}
		s.chatRepo.CreateMessage(msg)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
