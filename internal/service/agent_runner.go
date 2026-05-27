package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jincurry/go-clawith/internal/llm"
	"github.com/jincurry/go-clawith/internal/model"
	"github.com/jincurry/go-clawith/internal/repository"
	"github.com/jincurry/go-clawith/internal/tool"
)

const maxToolIterations = 10

type AgentRunner struct {
	llmReg   *llm.Registry
	toolRepo *repository.ToolRepository
	executor *tool.Executor
}

func NewAgentRunner(
	llmReg *llm.Registry,
	toolRepo *repository.ToolRepository,
	executor *tool.Executor,
) *AgentRunner {
	return &AgentRunner{
		llmReg:   llmReg,
		toolRepo: toolRepo,
		executor: executor,
	}
}

type RunResult struct {
	Content    string
	Messages   []llm.Message // full message trace including tool calls
	TokensUsed int
}

// Run executes a full agent loop: LLM call → tool execution → repeat until text response.
func (r *AgentRunner) Run(ctx context.Context, agent *model.Agent, history []llm.Message) (*RunResult, error) {
	provider, err := r.llmReg.Get(agent.Provider)
	if err != nil {
		return nil, err
	}

	tools, toolDefs := r.loadAgentTools(agent)

	messages := make([]llm.Message, len(history))
	copy(messages, history)

	var totalTokens int
	var traceMessages []llm.Message

	for i := 0; i < maxToolIterations; i++ {
		req := &llm.CompletionRequest{
			Model:       agent.Model,
			Messages:    messages,
			Temperature: agent.Temperature,
			MaxTokens:   agent.MaxTokens,
		}
		if len(toolDefs) > 0 {
			req.Tools = toolDefs
		}

		resp, err := provider.Complete(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("llm call failed: %w", err)
		}
		totalTokens += resp.TokensUsed

		if len(resp.ToolCalls) == 0 {
			return &RunResult{
				Content:    resp.Content,
				Messages:   traceMessages,
				TokensUsed: totalTokens,
			}, nil
		}

		// LLM wants to call tools
		assistantMsg := llm.Message{
			Role:      "assistant",
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		}
		messages = append(messages, assistantMsg)
		traceMessages = append(traceMessages, assistantMsg)

		for _, tc := range resp.ToolCalls {
			result := r.executeTool(ctx, tools, tc)

			toolMsg := llm.Message{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			}
			messages = append(messages, toolMsg)
			traceMessages = append(traceMessages, toolMsg)
		}
	}

	return nil, fmt.Errorf("agent exceeded maximum tool iterations (%d)", maxToolIterations)
}

// StreamWithTools executes an agent loop with streaming for the final text response.
// Tool calls are executed synchronously; only the final text response is streamed.
func (r *AgentRunner) StreamWithTools(ctx context.Context, agent *model.Agent, history []llm.Message) (<-chan llm.StreamChunk, []llm.Message, error) {
	provider, err := r.llmReg.Get(agent.Provider)
	if err != nil {
		return nil, nil, err
	}

	tools, toolDefs := r.loadAgentTools(agent)
	messages := make([]llm.Message, len(history))
	copy(messages, history)

	var traceMessages []llm.Message

	// Resolve tool calls synchronously first
	for i := 0; i < maxToolIterations; i++ {
		resp, err := provider.Complete(ctx, &llm.CompletionRequest{
			Model:       agent.Model,
			Messages:    messages,
			Tools:       toolDefs,
			Temperature: agent.Temperature,
			MaxTokens:   agent.MaxTokens,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("llm call failed: %w", err)
		}

		if len(resp.ToolCalls) == 0 {
			// No more tool calls — now stream the final response
			break
		}

		assistantMsg := llm.Message{
			Role:      "assistant",
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		}
		messages = append(messages, assistantMsg)
		traceMessages = append(traceMessages, assistantMsg)

		for _, tc := range resp.ToolCalls {
			result := r.executeTool(ctx, tools, tc)
			toolMsg := llm.Message{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			}
			messages = append(messages, toolMsg)
			traceMessages = append(traceMessages, toolMsg)
		}
	}

	// Stream the final response
	stream, err := provider.Stream(ctx, &llm.CompletionRequest{
		Model:       agent.Model,
		Messages:    messages,
		Temperature: agent.Temperature,
		MaxTokens:   agent.MaxTokens,
		Stream:      true,
	})
	if err != nil {
		return nil, nil, err
	}

	return stream, traceMessages, nil
}

func (r *AgentRunner) loadAgentTools(agent *model.Agent) (map[string]*model.Tool, []llm.ToolDef) {
	toolMap := make(map[string]*model.Tool)
	var toolDefs []llm.ToolDef

	if agent.Tools == nil {
		return toolMap, toolDefs
	}

	for _, at := range agent.Tools {
		t, err := r.toolRepo.GetByID(at.ToolID)
		if err != nil {
			log.Printf("failed to load tool %s: %v", at.ToolID, err)
			continue
		}
		toolMap[t.Name] = t

		var params any
		if t.Schema != "" {
			json.Unmarshal([]byte(t.Schema), &params)
		}
		if params == nil {
			params = map[string]any{"type": "object", "properties": map[string]any{}}
		}

		toolDefs = append(toolDefs, llm.ToolDef{
			Name:        t.Name,
			Description: t.Description,
			Parameters:  params,
		})
	}

	return toolMap, toolDefs
}

func (r *AgentRunner) executeTool(ctx context.Context, tools map[string]*model.Tool, tc llm.ToolCall) string {
	t, ok := tools[tc.Name]
	if !ok {
		return fmt.Sprintf("error: tool %q not found", tc.Name)
	}

	result, err := r.executor.Execute(ctx, t, tc.Arguments)
	if err != nil {
		return fmt.Sprintf("error executing tool %q: %v", tc.Name, err)
	}

	if result.Error != "" {
		return fmt.Sprintf("tool error: %s", result.Error)
	}

	return result.Output
}
