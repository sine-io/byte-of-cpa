package registry

import (
	"sort"
	"sync"
)

type ModelInfo struct {
	ID string
}

type ModelRegistry struct {
	mu sync.RWMutex

	clients       map[string]registeredClient
	modelToClient map[string]map[string]struct{}
	modelProvider map[string]map[string]struct{}
}

type registeredClient struct {
	provider string
	models   map[string]struct{}
}

func NewModelRegistry() *ModelRegistry {
	return &ModelRegistry{
		clients:       map[string]registeredClient{},
		modelToClient: map[string]map[string]struct{}{},
		modelProvider: map[string]map[string]struct{}{},
	}
}

func (r *ModelRegistry) RegisterClient(authID, provider string, models []ModelInfo) {
	if authID == "" {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.unregisterClientLocked(authID)

	clientModels := make(map[string]struct{}, len(models))
	for _, model := range models {
		if model.ID == "" {
			continue
		}
		clientModels[model.ID] = struct{}{}

		if _, ok := r.modelToClient[model.ID]; !ok {
			r.modelToClient[model.ID] = map[string]struct{}{}
		}
		r.modelToClient[model.ID][authID] = struct{}{}

		if _, ok := r.modelProvider[model.ID]; !ok {
			r.modelProvider[model.ID] = map[string]struct{}{}
		}
		r.modelProvider[model.ID][provider] = struct{}{}
	}

	r.clients[authID] = registeredClient{
		provider: provider,
		models:   clientModels,
	}
}

func (r *ModelRegistry) UnregisterClient(authID string) {
	if authID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.unregisterClientLocked(authID)
}

func (r *ModelRegistry) GetModelProviders(modelID string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := r.modelProvider[modelID]
	if len(providers) == 0 {
		return nil
	}

	result := make([]string, 0, len(providers))
	for provider := range providers {
		result = append(result, provider)
	}
	sort.Strings(result)
	return result
}

func (r *ModelRegistry) ClientSupportsModel(authID, modelID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	client, ok := r.clients[authID]
	if !ok {
		return false
	}
	_, exists := client.models[modelID]
	return exists
}

func (r *ModelRegistry) ListModels() []ModelInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.modelToClient))
	for modelID := range r.modelToClient {
		if len(r.modelToClient[modelID]) > 0 {
			ids = append(ids, modelID)
		}
	}
	sort.Strings(ids)

	result := make([]ModelInfo, 0, len(ids))
	for _, id := range ids {
		result = append(result, ModelInfo{ID: id})
	}
	return result
}

func (r *ModelRegistry) unregisterClientLocked(authID string) {
	client, ok := r.clients[authID]
	if !ok {
		return
	}

	for modelID := range client.models {
		if authIDs, exists := r.modelToClient[modelID]; exists {
			delete(authIDs, authID)
			if len(authIDs) == 0 {
				delete(r.modelToClient, modelID)
			}
		}

		providers := map[string]struct{}{}
		if authIDs, exists := r.modelToClient[modelID]; exists {
			for otherAuthID := range authIDs {
				otherClient, ok := r.clients[otherAuthID]
				if !ok {
					continue
				}
				providers[otherClient.provider] = struct{}{}
			}
		}
		if len(providers) == 0 {
			delete(r.modelProvider, modelID)
		} else {
			r.modelProvider[modelID] = providers
		}
	}

	delete(r.clients, authID)
}
