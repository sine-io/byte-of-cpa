package registry_test

import (
	"reflect"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/registry"
)

func TestModelRegistry_RegisterClientAndResolveModelProviders(t *testing.T) {
	t.Parallel()

	r := registry.NewModelRegistry()
	r.RegisterClient("auth-1", "claude", []registry.ModelInfo{
		{ID: "claude-3-7-sonnet"},
		{ID: "claude-3-5-haiku"},
	})
	r.RegisterClient("auth-2", "openai", []registry.ModelInfo{
		{ID: "claude-3-7-sonnet"},
	})

	if !r.ClientSupportsModel("auth-1", "claude-3-7-sonnet") {
		t.Fatal("expected auth-1 to support claude-3-7-sonnet")
	}
	if r.ClientSupportsModel("auth-1", "gpt-4o") {
		t.Fatal("did not expect auth-1 to support gpt-4o")
	}

	providers := r.GetModelProviders("claude-3-7-sonnet")
	wantProviders := []string{"claude", "openai"}
	if !reflect.DeepEqual(providers, wantProviders) {
		t.Fatalf("unexpected providers: got=%v want=%v", providers, wantProviders)
	}

	models := r.ListModels()
	wantModels := []registry.ModelInfo{
		{ID: "claude-3-5-haiku"},
		{ID: "claude-3-7-sonnet"},
	}
	if !reflect.DeepEqual(models, wantModels) {
		t.Fatalf("unexpected models: got=%v want=%v", models, wantModels)
	}
}

func TestModelRegistry_UnregisterClientRemovesModelAvailability(t *testing.T) {
	t.Parallel()

	r := registry.NewModelRegistry()
	r.RegisterClient("auth-1", "claude", []registry.ModelInfo{
		{ID: "shared-model"},
		{ID: "auth-1-only"},
	})
	r.RegisterClient("auth-2", "openai", []registry.ModelInfo{
		{ID: "shared-model"},
		{ID: "auth-2-only"},
	})

	r.UnregisterClient("auth-1")

	if r.ClientSupportsModel("auth-1", "shared-model") {
		t.Fatal("did not expect auth-1 to support shared-model after unregister")
	}
	if !r.ClientSupportsModel("auth-2", "shared-model") {
		t.Fatal("expected auth-2 to keep shared-model availability")
	}
	if providers := r.GetModelProviders("auth-1-only"); len(providers) != 0 {
		t.Fatalf("expected auth-1-only to be unavailable, got providers=%v", providers)
	}
	if providers := r.GetModelProviders("shared-model"); !reflect.DeepEqual(providers, []string{"openai"}) {
		t.Fatalf("unexpected providers for shared-model: %v", providers)
	}

	r.UnregisterClient("auth-2")

	if providers := r.GetModelProviders("shared-model"); len(providers) != 0 {
		t.Fatalf("expected no providers for shared-model, got %v", providers)
	}
	if models := r.ListModels(); len(models) != 0 {
		t.Fatalf("expected no available models, got %v", models)
	}
}
