package auth_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/auth"
	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/registry"
)

func TestAuth_HoldsProviderIdentityAndRuntimeAttributes(t *testing.T) {
	t.Parallel()

	now := time.Now()
	runtimeAuth := auth.Auth{
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

	if runtimeAuth.Provider != "claude" {
		t.Fatalf("expected provider claude, got %q", runtimeAuth.Provider)
	}
	if got := runtimeAuth.Attributes["api_key"]; got != "secret-key" {
		t.Fatalf("expected api_key to be preserved, got %q", got)
	}
	if got := runtimeAuth.Attributes["base_url"]; got != "https://api.anthropic.com" {
		t.Fatalf("expected base_url to be preserved, got %q", got)
	}
	if runtimeAuth.Status != auth.StatusActive {
		t.Fatalf("expected active status, got %q", runtimeAuth.Status)
	}
}

type fakeExecutor struct {
	execute func(ctx context.Context, openAIRequest []byte, runtimeAuth *auth.Auth) (*auth.Result, error)
}

func (f *fakeExecutor) Execute(ctx context.Context, openAIRequest []byte, runtimeAuth *auth.Auth) (*auth.Result, error) {
	return f.execute(ctx, openAIRequest, runtimeAuth)
}

func TestManager_RegisterAuthStoresRuntimeAuth(t *testing.T) {
	t.Parallel()

	mgr := auth.NewManager(registry.NewModelRegistry(), nil)
	runtimeAuth := &auth.Auth{
		ID:       "provider-1",
		Provider: "openai",
		Status:   auth.StatusActive,
		Attributes: map[string]string{
			"api_key": "secret",
		},
	}

	mgr.RegisterAuth(runtimeAuth)

	got := mgr.Auth("provider-1")
	if got != runtimeAuth {
		t.Fatalf("expected registered auth pointer, got %#v", got)
	}
	if missing := mgr.Auth("missing"); missing != nil {
		t.Fatalf("expected nil for missing auth lookup, got %#v", missing)
	}
}

func TestManager_RegisterExecutorStoresProviderExecutor(t *testing.T) {
	t.Parallel()

	mgr := auth.NewManager(registry.NewModelRegistry(), nil)
	executor := &fakeExecutor{
		execute: func(ctx context.Context, openAIRequest []byte, runtimeAuth *auth.Auth) (*auth.Result, error) {
			return &auth.Result{StatusCode: 200}, nil
		},
	}

	mgr.RegisterExecutor("openai", executor)

	if got := mgr.Executor("openai"); got != executor {
		t.Fatalf("expected registered executor, got %#v", got)
	}
	if missing := mgr.Executor("claude"); missing != nil {
		t.Fatalf("expected nil for missing executor lookup, got %#v", missing)
	}
}

func TestManager_CandidatesReturnsActiveAuthsForRequestedModel(t *testing.T) {
	t.Parallel()

	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient("openai-1", "openai", []registry.ModelInfo{{ID: "gpt-4o-mini"}})
	modelRegistry.RegisterClient("openai-2", "openai", []registry.ModelInfo{{ID: "gpt-4o-mini"}})
	modelRegistry.RegisterClient("claude-1", "claude", []registry.ModelInfo{{ID: "claude-3-5-haiku"}})

	mgr := auth.NewManager(modelRegistry, nil)
	mgr.RegisterAuth(&auth.Auth{ID: "openai-2", Provider: "openai", Status: auth.StatusActive})
	mgr.RegisterAuth(&auth.Auth{ID: "claude-1", Provider: "claude", Status: auth.StatusActive})
	mgr.RegisterAuth(&auth.Auth{ID: "openai-1", Provider: "openai", Status: auth.StatusActive})

	candidates, err := mgr.Candidates("gpt-4o-mini")
	if err != nil {
		t.Fatalf("candidates: %v", err)
	}
	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(candidates))
	}
	if candidates[0].ID != "openai-1" || candidates[1].ID != "openai-2" {
		t.Fatalf("expected model candidates sorted by auth id, got [%s %s]", candidates[0].ID, candidates[1].ID)
	}
}

func TestManager_SelectUsesDefaultRoundRobinSelector(t *testing.T) {
	t.Parallel()

	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient("openai-1", "openai", []registry.ModelInfo{{ID: "gpt-4o-mini"}})
	modelRegistry.RegisterClient("openai-2", "openai", []registry.ModelInfo{{ID: "gpt-4o-mini"}})

	mgr := auth.NewManager(modelRegistry, nil)
	mgr.RegisterAuth(&auth.Auth{ID: "openai-2", Provider: "openai", Status: auth.StatusActive})
	mgr.RegisterAuth(&auth.Auth{ID: "openai-1", Provider: "openai", Status: auth.StatusActive})

	first, err := mgr.Select("gpt-4o-mini")
	if err != nil {
		t.Fatalf("first select: %v", err)
	}
	second, err := mgr.Select("gpt-4o-mini")
	if err != nil {
		t.Fatalf("second select: %v", err)
	}
	third, err := mgr.Select("gpt-4o-mini")
	if err != nil {
		t.Fatalf("third select: %v", err)
	}

	if first.ID != "openai-1" {
		t.Fatalf("expected first selection to use sorted candidate order, got %q", first.ID)
	}
	if second.ID != "openai-2" {
		t.Fatalf("expected second selection to rotate to next candidate, got %q", second.ID)
	}
	if third.ID != "openai-1" {
		t.Fatalf("expected third selection to wrap to first candidate, got %q", third.ID)
	}
}

func TestManager_ExecuteReturnsExecutorBoundaryErrorForSupportedModel(t *testing.T) {
	t.Parallel()

	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient("openai-1", "openai", []registry.ModelInfo{{ID: "gpt-4o-mini"}})

	mgr := auth.NewManager(modelRegistry, nil)
	mgr.RegisterAuth(&auth.Auth{ID: "openai-1", Provider: "openai", Status: auth.StatusActive})

	if !mgr.SupportsModel("gpt-4o-mini") {
		t.Fatal("expected supported model when an active runtime auth exists")
	}

	_, err := mgr.Execute(context.Background(), "gpt-4o-mini", []byte(`{"model":"gpt-4o-mini"}`))
	if err == nil {
		t.Fatal("expected execute to fail without a provider executor")
	}
	if !strings.Contains(err.Error(), `no executor registered for provider "openai"`) {
		t.Fatalf("expected missing executor error, got %v", err)
	}
}

func TestManager_ExecuteForwardsSelectedAuthAndRequestBodyToExecutor(t *testing.T) {
	t.Parallel()

	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient("openai-1", "openai", []registry.ModelInfo{{ID: "gpt-4o-mini"}})
	modelRegistry.RegisterClient("openai-2", "openai", []registry.ModelInfo{{ID: "gpt-4o-mini"}})

	mgr := auth.NewManager(modelRegistry, nil)
	mgr.RegisterAuth(&auth.Auth{ID: "openai-2", Provider: "openai", Status: auth.StatusActive})
	mgr.RegisterAuth(&auth.Auth{ID: "openai-1", Provider: "openai", Status: auth.StatusActive})

	requestBody := []byte(`{"model":"gpt-4o-mini","messages":[]}`)
	executor := &fakeExecutor{
		execute: func(ctx context.Context, openAIRequest []byte, runtimeAuth *auth.Auth) (*auth.Result, error) {
			if runtimeAuth == nil {
				t.Fatal("expected selected runtime auth")
			}
			if runtimeAuth.ID != "openai-1" {
				t.Fatalf("expected first round-robin auth to be forwarded, got %q", runtimeAuth.ID)
			}
			if string(openAIRequest) != string(requestBody) {
				t.Fatalf("expected request body to be forwarded unchanged, got %s", string(openAIRequest))
			}

			return &auth.Result{
				StatusCode: 201,
				Body:       []byte(`{"ok":true}`),
			}, nil
		},
	}
	mgr.RegisterExecutor("openai", executor)

	result, err := mgr.Execute(context.Background(), "gpt-4o-mini", requestBody)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if result == nil {
		t.Fatal("expected result from executor")
	}
	if result.StatusCode != 201 {
		t.Fatalf("expected status 201, got %d", result.StatusCode)
	}
	if string(result.Body) != `{"ok":true}` {
		t.Fatalf("expected executor body to be returned, got %s", string(result.Body))
	}
}

func TestManager_CandidatesSkipDisabledAndCooldownAuths(t *testing.T) {
	t.Parallel()

	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient("auth-disabled", "openai", []registry.ModelInfo{{ID: "gpt-4o-mini"}})
	modelRegistry.RegisterClient("auth-cooldown", "openai", []registry.ModelInfo{{ID: "gpt-4o-mini"}})
	modelRegistry.RegisterClient("auth-active", "openai", []registry.ModelInfo{{ID: "gpt-4o-mini"}})

	mgr := auth.NewManager(modelRegistry, nil)
	mgr.RegisterAuth(&auth.Auth{ID: "auth-disabled", Provider: "openai", Status: auth.StatusActive, Disabled: true})
	mgr.RegisterAuth(&auth.Auth{ID: "auth-cooldown", Provider: "openai", Status: auth.StatusCooldown})
	mgr.RegisterAuth(&auth.Auth{ID: "auth-active", Provider: "openai", Status: auth.StatusActive})

	candidates, err := mgr.Candidates("gpt-4o-mini")
	if err != nil {
		t.Fatalf("candidates: %v", err)
	}
	if len(candidates) != 1 {
		t.Fatalf("expected 1 active candidate, got %d", len(candidates))
	}
	if candidates[0].ID != "auth-active" {
		t.Fatalf("expected active auth to be selected, got %q", candidates[0].ID)
	}
}
