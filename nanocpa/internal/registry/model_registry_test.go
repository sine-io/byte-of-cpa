package registry_test

import (
	"reflect"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/registry"
)

func TestModelRegistry_RegisterClientStoresModelAvailability(t *testing.T) {
	t.Parallel()

	r := registry.NewModelRegistry()

	r.RegisterClient("auth-1", "claude", []registry.ModelInfo{
		{ID: "claude-3-7-sonnet"},
		{ID: "claude-3-5-haiku"},
		{ID: ""},
	})

	if !r.ClientSupportsModel("auth-1", "claude-3-7-sonnet") {
		t.Fatal("expected registered client to support claude-3-7-sonnet")
	}
	if !r.ClientSupportsModel("auth-1", "claude-3-5-haiku") {
		t.Fatal("expected registered client to support claude-3-5-haiku")
	}
	if r.ClientSupportsModel("auth-1", "gpt-4o-mini") {
		t.Fatal("did not expect registered client to support gpt-4o-mini")
	}
}

func TestModelRegistry_ListModelsReturnsSortedUniqueModels(t *testing.T) {
	t.Parallel()

	r := registry.NewModelRegistry()
	r.RegisterClient("auth-1", "claude", []registry.ModelInfo{
		{ID: "claude-3-7-sonnet"},
		{ID: "claude-3-5-haiku"},
	})
	r.RegisterClient("auth-2", "openai", []registry.ModelInfo{
		{ID: "claude-3-7-sonnet"},
		{ID: "gpt-4o-mini"},
	})

	got := r.ListModels()
	want := []registry.ModelInfo{
		{ID: "claude-3-5-haiku"},
		{ID: "claude-3-7-sonnet"},
		{ID: "gpt-4o-mini"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected models: got=%v want=%v", got, want)
	}
}

func TestModelRegistry_GetModelProvidersReturnsSortedProviders(t *testing.T) {
	t.Parallel()

	r := registry.NewModelRegistry()
	r.RegisterClient("auth-1", "openai", []registry.ModelInfo{{ID: "shared-model"}})
	r.RegisterClient("auth-2", "claude", []registry.ModelInfo{{ID: "shared-model"}})

	got := r.GetModelProviders("shared-model")
	want := []string{"claude", "openai"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected providers: got=%v want=%v", got, want)
	}
}

func TestModelRegistry_ClientSupportsModelReflectsRegistration(t *testing.T) {
	t.Parallel()

	r := registry.NewModelRegistry()
	r.RegisterClient("auth-1", "claude", []registry.ModelInfo{{ID: "shared-model"}})

	if !r.ClientSupportsModel("auth-1", "shared-model") {
		t.Fatal("expected auth-1 to support shared-model")
	}
	if r.ClientSupportsModel("auth-2", "shared-model") {
		t.Fatal("did not expect unknown client to support shared-model")
	}
	if r.ClientSupportsModel("auth-1", "missing-model") {
		t.Fatal("did not expect auth-1 to support missing-model")
	}
}

func TestModelRegistry_UnregisterClientRemovesOrPreservesAvailabilityAsNeeded(t *testing.T) {
	t.Parallel()

	r := registry.NewModelRegistry()
	r.RegisterClient("auth-1", "claude", []registry.ModelInfo{
		{ID: "shared-model"},
		{ID: "claude-only"},
	})
	r.RegisterClient("auth-2", "openai", []registry.ModelInfo{
		{ID: "shared-model"},
		{ID: "openai-only"},
	})

	r.UnregisterClient("auth-1")

	if r.ClientSupportsModel("auth-1", "shared-model") {
		t.Fatal("did not expect auth-1 to support shared-model after unregister")
	}
	if !r.ClientSupportsModel("auth-2", "shared-model") {
		t.Fatal("expected auth-2 to keep shared-model after unregistering auth-1")
	}
	if providers := r.GetModelProviders("claude-only"); len(providers) != 0 {
		t.Fatalf("expected claude-only to be unavailable, got %v", providers)
	}
	if providers := r.GetModelProviders("shared-model"); !reflect.DeepEqual(providers, []string{"openai"}) {
		t.Fatalf("unexpected providers for shared-model after unregister: %v", providers)
	}

	r.UnregisterClient("auth-2")

	if providers := r.GetModelProviders("shared-model"); len(providers) != 0 {
		t.Fatalf("expected shared-model to be removed after unregistering all clients, got %v", providers)
	}
	if models := r.ListModels(); len(models) != 0 {
		t.Fatalf("expected no models after unregistering all clients, got %v", models)
	}
}
