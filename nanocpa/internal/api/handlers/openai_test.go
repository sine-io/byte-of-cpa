package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/api"
	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/api/handlers"
	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/registry"
)

func TestChatCompletionsSupportedModelStillReturnsStableAPIError(t *testing.T) {
	t.Parallel()

	rec := performAuthorizedRequest(t, newRegistry(
		registryRegistration{authID: "provider-1", provider: "openai", models: []string{"gpt-4o-mini"}},
	), http.MethodPost, "/v1/chat/completions", `{"model":"gpt-4o-mini","messages":[]}`)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 for registered route without upstream, got %d body=%s", rec.Code, rec.Body.String())
	}

	var response struct {
		Error struct {
			Type string `json:"type"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Error.Type != "api_error" {
		t.Fatalf("expected api_error, got %q", response.Error.Type)
	}
}

func TestOpenAI_ServerUsesRuntimeManagerForConfiguredChatModels(t *testing.T) {
	t.Parallel()

	server := api.NewServer(&config.Config{
		Host:    "127.0.0.1",
		Port:    18080,
		APIKeys: []string{"dev-key"},
		Providers: []config.Provider{
			{
				ID:       "provider-1",
				Provider: "openai",
				APIKey:   "key-1",
				BaseURL:  "https://example.invalid/openai",
				Models:   []string{"gpt-4o-mini"},
			},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"gpt-4o-mini","messages":[]}`))
	req.Header.Set("Authorization", "Bearer dev-key")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 for configured model without executor, got %d body=%s", rec.Code, rec.Body.String())
	}
	assertOpenAIError(t, rec, "api_error", "upstream provider request failed")
}

func TestOpenAI_ServerExecutesConfiguredClaudeProvider(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/messages" {
			t.Fatalf("expected /v1/messages, got %s", r.URL.Path)
		}
		if got := r.Header.Get("x-api-key"); got != "provider-secret" {
			t.Fatalf("expected x-api-key header, got %q", got)
		}
		if got := r.Header.Get("anthropic-version"); got != "2023-06-01" {
			t.Fatalf("expected anthropic-version 2023-06-01, got %q", got)
		}

		w.Header().Set("x-request-id", "req-claude")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id":"msg_1",
			"model":"claude-3-5-haiku",
			"content":[{"type":"text","text":"hello from claude"}]
		}`))
	}))
	defer upstream.Close()

	server := api.NewServer(&config.Config{
		Host:    "127.0.0.1",
		Port:    18080,
		APIKeys: []string{"dev-key"},
		Providers: []config.Provider{
			{
				ID:       "provider-1",
				Provider: "claude",
				APIKey:   "provider-secret",
				BaseURL:  upstream.URL,
				Models:   []string{"claude-3-5-haiku"},
			},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{
		"model":"claude-3-5-haiku",
		"messages":[
			{"role":"system","content":"be concise"},
			{"role":"user","content":"hello"}
		]
	}`))
	req.Header.Set("Authorization", "Bearer dev-key")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for configured Claude model, got %d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("x-request-id"); got != "req-claude" {
		t.Fatalf("expected x-request-id passthrough, got %q", got)
	}

	var response struct {
		Object  string `json:"object"`
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Object != "chat.completion" {
		t.Fatalf("expected chat.completion object, got %q", response.Object)
	}
	if response.Model != "claude-3-5-haiku" {
		t.Fatalf("expected model claude-3-5-haiku, got %q", response.Model)
	}
	if len(response.Choices) != 1 {
		t.Fatalf("expected one choice, got %d", len(response.Choices))
	}
	if response.Choices[0].Message.Role != "assistant" || response.Choices[0].Message.Content != "hello from claude" {
		t.Fatalf("unexpected assistant message: %+v", response.Choices[0].Message)
	}
}

func TestOpenAI_ServerRegistersRoutesBehindAccessMiddleware(t *testing.T) {
	t.Parallel()

	server := api.NewServer(&config.Config{
		Host:    "127.0.0.1",
		Port:    18080,
		APIKeys: []string{"dev-key"},
		Providers: []config.Provider{
			{
				ID:       "provider-1",
				Provider: "claude",
				APIKey:   "key-1",
				BaseURL:  "https://example.invalid/claude",
				Models:   []string{"claude-3-5-haiku"},
			},
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	req.Header.Set("Authorization", "Bearer dev-key")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for registered models route, got %d body=%s", rec.Code, rec.Body.String())
	}

	var response struct {
		Object string `json:"object"`
		Data   []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Object != "list" {
		t.Fatalf("expected list object, got %q", response.Object)
	}
	if len(response.Data) != 1 || response.Data[0].ID != "claude-3-5-haiku" {
		t.Fatalf("expected configured model list, got %+v", response.Data)
	}
}

func TestOpenAI_ServerRejectsUnauthorizedRequestsBeforeRouteHandlers(t *testing.T) {
	t.Parallel()

	server := api.NewServer(&config.Config{
		Host:    "127.0.0.1",
		Port:    18080,
		APIKeys: []string{"dev-key"},
	})
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	req.Header.Set("Authorization", "Bearer wrong-key")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when bearer token is invalid, got %d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("WWW-Authenticate"); got != "Bearer" {
		t.Fatalf("expected WWW-Authenticate Bearer, got %q", got)
	}
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Fatalf("expected JSON content type, got %q", got)
	}
	const wantBody = `{"error":{"message":"unauthorized","type":"invalid_request_error"}}`
	if got := strings.TrimSpace(rec.Body.String()); got != wantBody {
		t.Fatalf("expected body %q, got %q", wantBody, got)
	}

	assertOpenAIError(t, rec, "invalid_request_error", "unauthorized")
}

func TestChatCompletionsInvalidJSONReturnsOpenAIStyleError(t *testing.T) {
	t.Parallel()

	rec := performAuthorizedRequest(t, newRegistry(), http.MethodPost, "/v1/chat/completions", `{"model":`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d body=%s", rec.Code, rec.Body.String())
	}

	assertOpenAIError(t, rec, "invalid_request_error", "invalid JSON request body")
}

func TestChatCompletionsMissingModelReturnsValidationError(t *testing.T) {
	t.Parallel()

	rec := performAuthorizedRequest(t, newRegistry(), http.MethodPost, "/v1/chat/completions", `{"messages":[]}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing model, got %d body=%s", rec.Code, rec.Body.String())
	}

	assertOpenAIError(t, rec, "invalid_request_error", "model is required")
}

func TestChatCompletionsBodyTooLargeReturnsValidationError(t *testing.T) {
	t.Parallel()

	oversizedPrompt := strings.Repeat("x", 5*1024*1024)
	rec := performAuthorizedRequest(
		t,
		newRegistry(registryRegistration{authID: "provider-1", provider: "openai", models: []string{"gpt-4o-mini"}}),
		http.MethodPost,
		"/v1/chat/completions",
		`{"model":"gpt-4o-mini","messages":[{"role":"user","content":"`+oversizedPrompt+`"}]}`,
	)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 for oversized body, got %d body=%s", rec.Code, rec.Body.String())
	}

	assertOpenAIError(t, rec, "invalid_request_error", "request body too large")
}

func TestModelsReturnsConfiguredRegistryModels(t *testing.T) {
	t.Parallel()

	rec := performAuthorizedRequest(t, newRegistry(
		registryRegistration{authID: "provider-1", provider: "openai", models: []string{"gpt-4o-mini", "gpt-4.1-mini"}},
		registryRegistration{authID: "provider-2", provider: "claude", models: []string{"claude-3-5-haiku", "gpt-4o-mini"}},
	), http.MethodGet, "/v1/models", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for models list, got %d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Fatalf("expected JSON content type, got %q", got)
	}

	var response struct {
		Object string `json:"object"`
		Data   []struct {
			ID     string `json:"id"`
			Object string `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Object != "list" {
		t.Fatalf("expected list object, got %q", response.Object)
	}
	wantIDs := []string{"claude-3-5-haiku", "gpt-4.1-mini", "gpt-4o-mini"}
	if len(response.Data) != len(wantIDs) {
		t.Fatalf("expected %d models, got %d", len(wantIDs), len(response.Data))
	}
	for i, wantID := range wantIDs {
		if response.Data[i].ID != wantID {
			t.Fatalf("unexpected model at index %d: got=%q want=%q", i, response.Data[i].ID, wantID)
		}
		if response.Data[i].Object != "model" {
			t.Fatalf("expected model object at index %d, got %q", i, response.Data[i].Object)
		}
	}
}

func TestChatCompletionsRejectsUnsupportedModelBeforeUpstreamHandling(t *testing.T) {
	t.Parallel()

	rec := performAuthorizedRequest(t, newRegistry(
		registryRegistration{authID: "provider-1", provider: "openai", models: []string{"gpt-4o-mini"}},
	), http.MethodPost, "/v1/chat/completions", `{"model":"claude-3-5-haiku","messages":[]}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unsupported model, got %d body=%s", rec.Code, rec.Body.String())
	}

	assertOpenAIError(t, rec, "invalid_request_error", `model "claude-3-5-haiku" is not available`)
}

func performAuthorizedRequest(t *testing.T, modelRegistry *registry.ModelRegistry, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	openAI := handlers.NewOpenAI(modelRegistry, nil)
	mux := http.NewServeMux()
	openAI.RegisterRoutes(mux)
	handler := api.APIKeyMiddleware([]string{"dev-key"}, mux)

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer dev-key")
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

type registryRegistration struct {
	authID   string
	provider string
	models   []string
}

func newRegistry(registrations ...registryRegistration) *registry.ModelRegistry {
	r := registry.NewModelRegistry()
	for _, registration := range registrations {
		models := make([]registry.ModelInfo, 0, len(registration.models))
		for _, modelID := range registration.models {
			models = append(models, registry.ModelInfo{ID: modelID})
		}
		r.RegisterClient(registration.authID, registration.provider, models)
	}
	return r
}

func assertOpenAIError(t *testing.T, rec *httptest.ResponseRecorder, wantType, wantMessageSubstring string) {
	t.Helper()

	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Fatalf("expected JSON content type, got %q", got)
	}

	var response struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Error.Type != wantType {
		t.Fatalf("expected error type %q, got %q", wantType, response.Error.Type)
	}
	if !strings.Contains(response.Error.Message, wantMessageSubstring) {
		t.Fatalf("expected message containing %q, got %q", wantMessageSubstring, response.Error.Message)
	}
}
