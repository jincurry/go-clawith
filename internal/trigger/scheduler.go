package trigger

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jincurry/go-clawith/internal/llm"
	"github.com/jincurry/go-clawith/internal/model"
	"github.com/jincurry/go-clawith/internal/repository"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron        *cron.Cron
	triggerRepo *repository.TriggerRepository
	agentRepo   *repository.AgentRepository
	llmReg      *llm.Registry
	entries     map[uuid.UUID]cron.EntryID
}

func NewScheduler(
	triggerRepo *repository.TriggerRepository,
	agentRepo *repository.AgentRepository,
	llmReg *llm.Registry,
) *Scheduler {
	return &Scheduler{
		cron:        cron.New(cron.WithSeconds()),
		triggerRepo: triggerRepo,
		agentRepo:   agentRepo,
		llmReg:      llmReg,
		entries:     make(map[uuid.UUID]cron.EntryID),
	}
}

func (s *Scheduler) Start() error {
	triggers, err := s.triggerRepo.ListActive()
	if err != nil {
		return err
	}

	for i := range triggers {
		if err := s.addTrigger(&triggers[i]); err != nil {
			log.Printf("failed to add trigger %s: %v", triggers[i].ID, err)
		}
	}

	s.cron.Start()
	log.Printf("trigger scheduler started with %d triggers", len(triggers))
	return nil
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}

func (s *Scheduler) AddTrigger(trigger *model.Trigger) error {
	return s.addTrigger(trigger)
}

func (s *Scheduler) RemoveTrigger(id uuid.UUID) {
	if entryID, ok := s.entries[id]; ok {
		s.cron.Remove(entryID)
		delete(s.entries, id)
	}
}

func (s *Scheduler) addTrigger(trigger *model.Trigger) error {
	if trigger.Type != "cron" {
		return nil
	}

	var cfg struct {
		Expression string `json:"expression"`
	}
	if err := json.Unmarshal([]byte(trigger.Config), &cfg); err != nil {
		return err
	}

	triggerID := trigger.ID
	entryID, err := s.cron.AddFunc(cfg.Expression, func() {
		s.executeTrigger(triggerID)
	})
	if err != nil {
		return err
	}

	s.entries[trigger.ID] = entryID
	return nil
}

func (s *Scheduler) executeTrigger(triggerID uuid.UUID) {
	trigger, err := s.triggerRepo.GetByID(triggerID)
	if err != nil || !trigger.IsActive {
		return
	}

	exec := &model.TriggerExecution{
		TriggerID: triggerID,
		Status:    "running",
		StartedAt: time.Now(),
	}
	s.triggerRepo.RecordExecution(exec)

	agent, err := s.agentRepo.GetByID(trigger.AgentID)
	if err != nil {
		s.failExecution(exec, err.Error())
		return
	}

	provider, err := s.llmReg.Get(agent.Provider)
	if err != nil {
		s.failExecution(exec, err.Error())
		return
	}

	messages := []llm.Message{
		{Role: "system", Content: agent.SystemPrompt},
		{Role: "user", Content: trigger.Action},
	}

	resp, err := provider.Complete(context.Background(), &llm.CompletionRequest{
		Model:       agent.Model,
		Messages:    messages,
		Temperature: agent.Temperature,
		MaxTokens:   agent.MaxTokens,
	})
	if err != nil {
		s.failExecution(exec, err.Error())
		return
	}

	now := time.Now()
	exec.Status = "success"
	exec.Output = resp.Content
	exec.EndedAt = &now
	s.triggerRepo.RecordExecution(exec)
	s.triggerRepo.UpdateLastRun(triggerID, now)
}

func (s *Scheduler) failExecution(exec *model.TriggerExecution, errMsg string) {
	now := time.Now()
	exec.Status = "failed"
	exec.Error = errMsg
	exec.EndedAt = &now
	s.triggerRepo.RecordExecution(exec)
}
