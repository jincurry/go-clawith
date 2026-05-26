package service

import (
	"github.com/google/uuid"
	"github.com/jincurry/go-clawith/internal/model"
	"github.com/jincurry/go-clawith/internal/repository"
)

type AgentService struct {
	agentRepo *repository.AgentRepository
}

func NewAgentService(agentRepo *repository.AgentRepository) *AgentService {
	return &AgentService{agentRepo: agentRepo}
}

type CreateAgentInput struct {
	Name         string   `json:"name" binding:"required"`
	Description  string   `json:"description"`
	SystemPrompt string   `json:"system_prompt"`
	Model        string   `json:"model" binding:"required"`
	Provider     string   `json:"provider" binding:"required"`
	Temperature  float64  `json:"temperature"`
	MaxTokens    int      `json:"max_tokens"`
	Persona      string   `json:"persona"`
	ToolIDs      []string `json:"tool_ids"`
}

type UpdateAgentInput struct {
	Name         *string  `json:"name"`
	Description  *string  `json:"description"`
	SystemPrompt *string  `json:"system_prompt"`
	Model        *string  `json:"model"`
	Provider     *string  `json:"provider"`
	Temperature  *float64 `json:"temperature"`
	MaxTokens    *int     `json:"max_tokens"`
	Persona      *string  `json:"persona"`
	Memory       *string  `json:"memory"`
	IsActive     *bool    `json:"is_active"`
	ToolIDs      []string `json:"tool_ids"`
}

func (s *AgentService) Create(userID uuid.UUID, input CreateAgentInput) (*model.Agent, error) {
	agent := &model.Agent{
		UserID:       userID,
		Name:         input.Name,
		Description:  input.Description,
		SystemPrompt: input.SystemPrompt,
		Model:        input.Model,
		Provider:     input.Provider,
		Temperature:  input.Temperature,
		MaxTokens:    input.MaxTokens,
		Persona:      input.Persona,
	}

	if agent.Temperature == 0 {
		agent.Temperature = 0.7
	}
	if agent.MaxTokens == 0 {
		agent.MaxTokens = 4096
	}

	if err := s.agentRepo.Create(agent); err != nil {
		return nil, err
	}

	if len(input.ToolIDs) > 0 {
		toolUUIDs := make([]uuid.UUID, 0, len(input.ToolIDs))
		for _, id := range input.ToolIDs {
			uid, err := uuid.Parse(id)
			if err == nil {
				toolUUIDs = append(toolUUIDs, uid)
			}
		}
		if err := s.agentRepo.SetTools(agent.ID, toolUUIDs); err != nil {
			return nil, err
		}
	}

	return s.agentRepo.GetByID(agent.ID)
}

func (s *AgentService) Get(id uuid.UUID) (*model.Agent, error) {
	return s.agentRepo.GetByID(id)
}

func (s *AgentService) List(userID uuid.UUID, page, pageSize int) ([]model.Agent, int64, error) {
	offset := (page - 1) * pageSize
	return s.agentRepo.ListByUser(userID, offset, pageSize)
}

func (s *AgentService) Update(id uuid.UUID, input UpdateAgentInput) (*model.Agent, error) {
	agent, err := s.agentRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		agent.Name = *input.Name
	}
	if input.Description != nil {
		agent.Description = *input.Description
	}
	if input.SystemPrompt != nil {
		agent.SystemPrompt = *input.SystemPrompt
	}
	if input.Model != nil {
		agent.Model = *input.Model
	}
	if input.Provider != nil {
		agent.Provider = *input.Provider
	}
	if input.Temperature != nil {
		agent.Temperature = *input.Temperature
	}
	if input.MaxTokens != nil {
		agent.MaxTokens = *input.MaxTokens
	}
	if input.Persona != nil {
		agent.Persona = *input.Persona
	}
	if input.Memory != nil {
		agent.Memory = *input.Memory
	}
	if input.IsActive != nil {
		agent.IsActive = *input.IsActive
	}

	if err := s.agentRepo.Update(agent); err != nil {
		return nil, err
	}

	if input.ToolIDs != nil {
		toolUUIDs := make([]uuid.UUID, 0, len(input.ToolIDs))
		for _, tid := range input.ToolIDs {
			uid, err := uuid.Parse(tid)
			if err == nil {
				toolUUIDs = append(toolUUIDs, uid)
			}
		}
		if err := s.agentRepo.SetTools(id, toolUUIDs); err != nil {
			return nil, err
		}
	}

	return s.agentRepo.GetByID(id)
}

func (s *AgentService) Delete(id uuid.UUID) error {
	return s.agentRepo.Delete(id)
}
