package ai

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

type Provider interface {
	Name() string
	Complete(ctx context.Context, req CompletionRequest) (string, error)
	IsAvailable() bool
	DefaultModel() string
}

type CompletionRequest struct {
	SystemPrompt string
	UserPrompt   string
	Model        string
	MaxTokens    int
	Temperature  float32
}

type ProviderFactory func(cfg *Config) (Provider, error)

var (
	providersMu sync.RWMutex
	providers   = make(map[string]ProviderFactory)
)

func RegisterProvider(name string, factory ProviderFactory) {
	providersMu.Lock()
	defer providersMu.Unlock()

	if factory == nil {
		panic("ai: RegisterProvider factory is nil")
	}
	if _, dup := providers[name]; dup {
		panic("ai: RegisterProvider called twice for provider " + name)
	}
	providers[name] = factory
}

func NewProvider(cfg *Config) (Provider, error) {
	name := cfg.Provider
	if name == "" {
		name = "openai"
	}

	providersMu.RLock()
	factory, ok := providers[name]
	providersMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("ai: unknown provider %q (registered: %v)", name, RegisteredProviders())
	}
	return factory(cfg)
}

func RegisteredProviders() []string {
	providersMu.RLock()
	defer providersMu.RUnlock()

	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func ResetProviders() {
	providersMu.Lock()
	defer providersMu.Unlock()
	providers = make(map[string]ProviderFactory)
}
