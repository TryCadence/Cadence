package patterns

import (
	"github.com/codemeapixel/cadence/internal/git"
	"github.com/codemeapixel/cadence/internal/metrics"
)

// Strategy defines the interface for different detection methods
type Strategy interface {
	Name() string
	Detect(pair *git.CommitPair, repoStats *metrics.RepositoryStats) (bool, string)
}

// Registry manages available detection strategies
type Registry struct {
	strategies map[string]Strategy
}

// NewRegistry creates a new strategy registry
func NewRegistry() *Registry {
	return &Registry{
		strategies: make(map[string]Strategy),
	}
}

// Register adds a strategy to the registry
func (r *Registry) Register(strategy Strategy) {
	r.strategies[strategy.Name()] = strategy
}

// Get retrieves a strategy by name
func (r *Registry) Get(name string) Strategy {
	return r.strategies[name]
}

// All returns all registered strategies
func (r *Registry) All() []Strategy {
	strategies := make([]Strategy, 0, len(r.strategies))
	for _, s := range r.strategies {
		strategies = append(strategies, s)
	}
	return strategies
}

// Detect runs all registered strategies
func (r *Registry) Detect(pair *git.CommitPair, repoStats *metrics.RepositoryStats) []string {
	results := make([]string, 0)
	for _, strategy := range r.strategies {
		if detected, reason := strategy.Detect(pair, repoStats); detected {
			results = append(results, reason)
		}
	}
	return results
}
