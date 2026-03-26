package auth

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/registry"
)

type Manager struct {
	mu sync.RWMutex

	auths     map[string]*Auth
	executors map[string]Executor
	registry  *registry.ModelRegistry
	selector  Selector
}

func modelNotAvailableError(model string) error {
	return fmt.Errorf("model %q is not available", model)
}

func NewManager(modelRegistry *registry.ModelRegistry, selector Selector) *Manager {
	if selector == nil {
		selector = NewRoundRobinSelector()
	}
	return &Manager{
		auths:     map[string]*Auth{},
		executors: map[string]Executor{},
		registry:  modelRegistry,
		selector:  selector,
	}
}

func (m *Manager) Auth(id string) *Auth {
	if id == "" {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.auths[id]
}

func (m *Manager) RegisterAuth(runtimeAuth *Auth) {
	if runtimeAuth == nil || runtimeAuth.ID == "" {
		return
	}
	m.mu.Lock()
	m.auths[runtimeAuth.ID] = runtimeAuth
	m.mu.Unlock()
}

func (m *Manager) Executor(provider string) Executor {
	if provider == "" {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.executors[provider]
}

func (m *Manager) RegisterExecutor(provider string, chatExecutor Executor) {
	if provider == "" || chatExecutor == nil {
		return
	}
	m.mu.Lock()
	m.executors[provider] = chatExecutor
	m.mu.Unlock()
}

func (m *Manager) Candidates(model string) ([]*Auth, error) {
	if m.registry == nil {
		return nil, fmt.Errorf("model registry is required")
	}

	providers := m.registry.GetModelProviders(model)
	if len(providers) == 0 {
		return nil, modelNotAvailableError(model)
	}

	providerSet := make(map[string]struct{}, len(providers))
	for _, provider := range providers {
		providerSet[provider] = struct{}{}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	candidates := make([]*Auth, 0, len(m.auths))
	for _, runtimeAuth := range m.auths {
		if runtimeAuth == nil {
			continue
		}
		if runtimeAuth.Disabled || runtimeAuth.Status != StatusActive {
			continue
		}
		if _, ok := providerSet[runtimeAuth.Provider]; !ok {
			continue
		}
		if !m.registry.ClientSupportsModel(runtimeAuth.ID, model) {
			continue
		}
		candidates = append(candidates, runtimeAuth)
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].ID < candidates[j].ID
	})

	return candidates, nil
}

func (m *Manager) Select(model string) (*Auth, error) {
	candidates, err := m.Candidates(model)
	if err != nil {
		return nil, err
	}

	selected := m.selector.Select(model, candidates)
	if selected == nil {
		return nil, fmt.Errorf("no runtime auth available for model %q", model)
	}

	return selected, nil
}

func (m *Manager) Execute(ctx context.Context, model string, openAIRequest []byte) (*Result, error) {
	selected, err := m.Select(model)
	if err != nil {
		return nil, err
	}

	selectedExecutor := m.Executor(selected.Provider)
	if selectedExecutor == nil {
		return nil, fmt.Errorf("no executor registered for provider %q", selected.Provider)
	}

	return selectedExecutor.Execute(ctx, openAIRequest, selected)
}

func (m *Manager) SupportsModel(model string) bool {
	candidates, err := m.Candidates(model)
	return err == nil && len(candidates) > 0
}
