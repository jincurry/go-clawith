package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jincurry/go-clawith/internal/llm"
	"github.com/jincurry/go-clawith/internal/model"
	"github.com/jincurry/go-clawith/internal/repository"
)

type ChatService struct {
	chatRepo  *repository.ChatRepository
	agentRepo *repository.AgentRepository
	llmReg    *llm.Registry
}

func NewChatService(
	chatRepo *repository.ChatRepository,
	agentRepo *repository.AgentRepository,
	llmReg *llm.Registry,
) *ChatService {
	return &ChatService{
		chatRepo:  chatRepo,
		agentRepo: agentRepo,
		llmReg:    llmReg,
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

	provider, err := s.llmReg.Get(agent.Provider)
	if err != nil {
		return nil, err
	}

	resp, err := provider.Complete(ctx, &llm.CompletionRequest{
		Model:       agent.Model,
		Messages:    messages,
		Temperature: agent.Temperature,
		MaxTokens:   agent.MaxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("llm error: %w", err)
	}

	assistantMsg := &model.Message{
		SessionID:  sessionID,
		Role:       "assistant",
		Content:    resp.Content,
		TokenCount: resp.TokensUsed,
	}
	if err := s.chatRepo.CreateMessage(assistantMsg); err != nil {
		return nil, err
	}

	return assistantMsg, nil
}

func (s *ChatService) StreamMessage(ctx context.Context, sessionID uuid.UUID, content string) (<-chan llm.StreamChunk, error) {
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

	provider, err := s.llmReg.Get(agent.Provider)
	if err != nil {
		return nil, err
	}

	return provider.Stream(ctx, &llm.CompletionRequest{
		Model:       agent.Model,
		Messages:    messages,
		Temperature: agent.Temperature,
		MaxTokens:   agent.MaxTokens,
		Stream:      true,
	})
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
		msgs = append(msgs, llm.Message{
			Role:       m.Role,
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		})
	}

	return msgs
}
