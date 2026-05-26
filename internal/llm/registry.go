package llm

import (
	"fmt"
	"sync"

	"github.com/jincurry/go-clawith/internal/config"
)

type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
	default_  string
}

func NewRegistry(cfg config.LLMConfig) *Registry {
	r := &Registry{
		providers: make(map[string]Provider),
		default_:  cfg.DefaultProvider,
	}

	if cfg.OpenAI.APIKey != "" {
		r.Register(NewOpenAIProvider(cfg.OpenAI))
	}
	if cfg.Anthropic.APIKey != "" {
		r.Register(NewAnthropicProvider(cfg.Anthropic))
	}

	return r
}

func (r *Registry) Register(p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.Name()] = p
}

func (r *Registry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if name == "" {
		name = r.default_
	}
	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %q not registered", name)
	}
	return p, nil
}

func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}
