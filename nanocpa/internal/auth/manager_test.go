package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/auth"
	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/registry"
)

func TestAuth_HoldsProviderIdentityAndRuntimeAttributes(t *testing.T) {
	t.Parallel()

	now := time.Now()
	a := auth.Auth{
		ID:       "provider-1",
		Provider: "claude",
		Label:    "Primary Claude",
		Status:   auth.StatusActive,
		Attributes: map[string]string{
			"api_key":  "secret-key",
			"base_url": "https://api.anthropic.com",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if a.Provider != "claude" {
		t.Fatalf("expected provider claude, got %q", a.Provider)
	}
	if got := a.Attributes["api_key"]; got != "secret-key" {
		t.Fatalf("expected api_key to be preserved, got %q", got)
	}
	if got := a.Attributes["base_url"]; got != "https://api.anthropic.com" {
		t.Fatalf("expected base_url to be preserved, got %q", got)
	}
	if a.Status != auth.StatusActive {
		t.Fatalf("expected active status, got %q", a.Status)
	}
}

type fakeExecutor struct {
	execute func(ctx context.Context, openAIRequest []byte, runtimeAuth *auth.Auth) (*auth.Result, error)
}

func (f *fakeExecutor) Execute(ctx context.Context, openAIRequest []byte, runtimeAuth *auth.Auth) (*auth.Result, error) {
	return f.execute(ctx, openAIRequest, runtimeAuth)
}

func TestManager_Execute_RoundRobinAcrossAuthsForSameModel(t *testing.T) {
	t.Parallel()

	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient("a-1", "claude", []registry.ModelInfo{{ID: "shared-model"}})
	modelRegistry.RegisterClient("a-2", "claude", []registry.ModelInfo{{ID: "shared-model"}})

	mgr := auth.NewManager(modelRegistry, nil)
	mgr.RegisterAuth(&auth.Auth{ID: "a-1", Provider: "claude", Status: auth.StatusActive, Attributes: map[string]string{"api_key": "key-1"}})
	mgr.RegisterAuth(&auth.Auth{ID: "a-2", Provider: "claude", Status: auth.StatusActive, Attributes: map[string]string{"api_key": "key-2"}})

	var gotAuthIDs []string
	mgr.RegisterExecutor("claude", &fakeExecutor{
		execute: func(ctx context.Context, openAIRequest []byte, runtimeAuth *auth.Auth) (*auth.Result, error) {
			gotAuthIDs = append(gotAuthIDs, runtimeAuth.ID)
			return &auth.Result{StatusCode: 200, Body: []byte(`{"ok":true}`)}, nil
		},
	})

	for i := 0; i < 4; i++ {
		if _, err := mgr.Execute(context.Background(), "shared-model", []byte(`{}`)); err != nil {
			t.Fatalf("execute %d: %v", i+1, err)
		}
	}

	if len(gotAuthIDs) != 4 {
		t.Fatalf("expected 4 executions, got %d", len(gotAuthIDs))
	}

	want := []string{"a-1", "a-2", "a-1", "a-2"}
	for i := range want {
		if gotAuthIDs[i] != want[i] {
			t.Fatalf("round-robin mismatch at call %d: got %q want %q", i+1, gotAuthIDs[i], want[i])
		}
	}
}

func TestManager_Execute_RoundRobinIsIsolatedPerModel(t *testing.T) {
	t.Parallel()

	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient("a-1", "claude", []registry.ModelInfo{{ID: "model-a"}, {ID: "model-b"}})
	modelRegistry.RegisterClient("a-2", "claude", []registry.ModelInfo{{ID: "model-a"}, {ID: "model-b"}})

	mgr := auth.NewManager(modelRegistry, nil)
	mgr.RegisterAuth(&auth.Auth{ID: "a-1", Provider: "claude", Status: auth.StatusActive, Attributes: map[string]string{"api_key": "key-1"}})
	mgr.RegisterAuth(&auth.Auth{ID: "a-2", Provider: "claude", Status: auth.StatusActive, Attributes: map[string]string{"api_key": "key-2"}})

	gotByModel := map[string][]string{}
	mgr.RegisterExecutor("claude", &fakeExecutor{
		execute: func(ctx context.Context, openAIRequest []byte, runtimeAuth *auth.Auth) (*auth.Result, error) {
			model := string(openAIRequest)
			gotByModel[model] = append(gotByModel[model], runtimeAuth.ID)
			return &auth.Result{StatusCode: 200, Body: []byte(`{"ok":true}`)}, nil
		},
	})

	sequence := []string{"model-a", "model-b", "model-a", "model-b", "model-a", "model-b"}
	for i, model := range sequence {
		if _, err := mgr.Execute(context.Background(), model, []byte(model)); err != nil {
			t.Fatalf("execute %d for %s: %v", i+1, model, err)
		}
	}

	want := []string{"a-1", "a-2", "a-1"}
	for _, model := range []string{"model-a", "model-b"} {
		got := gotByModel[model]
		if len(got) != len(want) {
			t.Fatalf("unexpected call count for %s: got=%d want=%d", model, len(got), len(want))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("unexpected auth for %s call %d: got=%q want=%q", model, i+1, got[i], want[i])
			}
		}
	}
}

func TestManager_Execute_ReturnsErrorWhenNoProviderSupportsModel(t *testing.T) {
	t.Parallel()

	mgr := auth.NewManager(registry.NewModelRegistry(), nil)
	mgr.RegisterAuth(&auth.Auth{ID: "a-1", Provider: "claude", Status: auth.StatusActive})
	mgr.RegisterExecutor("claude", &fakeExecutor{
		execute: func(ctx context.Context, openAIRequest []byte, runtimeAuth *auth.Auth) (*auth.Result, error) {
			return nil, errors.New("should not be called")
		},
	})

	if _, err := mgr.Execute(context.Background(), "missing-model", []byte(`{}`)); err == nil {
		t.Fatal("expected execute to fail for unsupported model")
	}
}

func TestManager_Execute_SkipsDisabledAndCooldownAuths(t *testing.T) {
	t.Parallel()

	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient("a-disabled", "claude", []registry.ModelInfo{{ID: "shared-model"}})
	modelRegistry.RegisterClient("a-cooldown", "claude", []registry.ModelInfo{{ID: "shared-model"}})
	modelRegistry.RegisterClient("a-active", "claude", []registry.ModelInfo{{ID: "shared-model"}})

	mgr := auth.NewManager(modelRegistry, nil)
	mgr.RegisterAuth(&auth.Auth{ID: "a-disabled", Provider: "claude", Status: auth.StatusActive, Disabled: true})
	mgr.RegisterAuth(&auth.Auth{ID: "a-cooldown", Provider: "claude", Status: auth.StatusCooldown})
	mgr.RegisterAuth(&auth.Auth{ID: "a-active", Provider: "claude", Status: auth.StatusActive})

	var gotAuthID string
	mgr.RegisterExecutor("claude", &fakeExecutor{
		execute: func(ctx context.Context, openAIRequest []byte, runtimeAuth *auth.Auth) (*auth.Result, error) {
			gotAuthID = runtimeAuth.ID
			return &auth.Result{StatusCode: 200, Body: []byte(`{"ok":true}`)}, nil
		},
	})

	if _, err := mgr.Execute(context.Background(), "shared-model", []byte(`{}`)); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if gotAuthID != "a-active" {
		t.Fatalf("expected active auth to be selected, got %q", gotAuthID)
	}
}

func TestManager_SupportsModel_RequiresAtLeastOneActiveAuth(t *testing.T) {
	t.Parallel()

	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient("a-disabled", "claude", []registry.ModelInfo{{ID: "shared-model"}})
	modelRegistry.RegisterClient("a-cooldown", "claude", []registry.ModelInfo{{ID: "shared-model"}})

	mgr := auth.NewManager(modelRegistry, nil)
	mgr.RegisterAuth(&auth.Auth{ID: "a-disabled", Provider: "claude", Status: auth.StatusActive, Disabled: true})
	mgr.RegisterAuth(&auth.Auth{ID: "a-cooldown", Provider: "claude", Status: auth.StatusCooldown})
	mgr.RegisterExecutor("claude", &fakeExecutor{
		execute: func(ctx context.Context, openAIRequest []byte, runtimeAuth *auth.Auth) (*auth.Result, error) {
			return &auth.Result{StatusCode: 200, Body: []byte(`{"ok":true}`)}, nil
		},
	})

	if mgr.SupportsModel("shared-model") {
		t.Fatal("expected model to be unsupported when only disabled/cooldown auths exist")
	}
}
