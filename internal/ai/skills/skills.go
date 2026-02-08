package skills

import (
	"fmt"
	"sort"
	"sync"
)

type Skill interface {
	Name() string
	Description() string
	Category() string
	SystemPrompt() string
	FormatInput(input interface{}) (string, error)
	ParseOutput(raw string) (interface{}, error)
	MaxTokens() int
}

var (
	registryMu sync.RWMutex
	registry   = make(map[string]Skill)
)

func Register(skill Skill) {
	registryMu.Lock()
	defer registryMu.Unlock()

	if skill == nil {
		panic("skills: Register skill is nil")
	}
	name := skill.Name()
	if _, dup := registry[name]; dup {
		panic("skills: Register called twice for skill " + name)
	}
	registry[name] = skill
}

func Get(name string) (Skill, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	skill, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("skills: unknown skill %q (registered: %v)", name, RegisteredSkills())
	}
	return skill, nil
}

func RegisteredSkills() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func All() []Skill {
	registryMu.RLock()
	defer registryMu.RUnlock()

	skills := make([]Skill, 0, len(registry))
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		skills = append(skills, registry[name])
	}
	return skills
}

func ResetRegistry() {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry = make(map[string]Skill)
}
