package handlers_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/api"
	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/api/handlers"
	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/auth"
	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/registry"
	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/runtime/executor"
)

type testChatRunner struct {
	execute func(ctx context.Context, model string, requestBody []byte) (*auth.Result, error)
}

func (r *testChatRunner) Execute(ctx context.Context, model string, requestBody []byte) (*auth.Result, error) {
	return r.execute(ctx, model, requestBody)
}

type stateAwareRunner struct {
	testChatRunner
	supportsModel func(model string) bool
}

func (r *stateAwareRunner) SupportsModel(model string) bool {
	return r.supportsModel(model)
}

func TestOpenAIHandler_ChatCompletions_EndToEnd(t *testing.T) {
	t.Parallel()

	var hits atomic.Int64
	var upstreamAPIKey atomic.Value
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		upstreamAPIKey.Store(r.Header.Get("x-api-key"))
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id":"msg_1",
			"model":"claude-3-7-sonnet",
			"content":[{"type":"text","text":"hello from claude"}]
		}`))
	}))
	defer upstream.Close()

	runtimeAuth := &auth.Auth{
		ID:       "p1",
		Provider: "claude",
		Attributes: map[string]string{
			"api_key":  "provider-secret",
			"base_url": upstream.URL,
		},
	}
	claudeExec := executor.NewClaude(upstream.Client())
	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient(runtimeAuth.ID, runtimeAuth.Provider, []registry.ModelInfo{{ID: "claude-3-7-sonnet"}})
	runner := &testChatRunner{
		execute: func(ctx context.Context, model string, requestBody []byte) (*auth.Result, error) {
			result, err := claudeExec.Execute(ctx, requestBody, runtimeAuth)
			if err != nil {
				return nil, err
			}
			return &auth.Result{StatusCode: result.StatusCode, Body: result.Body, Headers: result.Headers}, nil
		},
	}

	h := handlers.NewOpenAI(runner, modelRegistry)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	securedMux := api.APIKeyMiddleware([]string{"dev-key"}, mux)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", http.NoBody)
	req.Header.Set("authorization", "Bearer dev-key")
	req.Body = io.NopCloser(strings.NewReader(`{"model":"claude-3-7-sonnet","messages":[{"role":"user","content":"hello"}]}`))
	req.Header.Set("content-type", "application/json")
	rec := httptest.NewRecorder()

	securedMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 status, got %d body=%s", rec.Code, rec.Body.String())
	}
	if hits.Load() != 1 {
		t.Fatalf("expected one upstream call, got %d", hits.Load())
	}
	gotAPIKey, _ := upstreamAPIKey.Load().(string)
	if gotAPIKey != "provider-secret" {
		t.Fatalf("expected provider api key header, got %q", gotAPIKey)
	}
	var decoded map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if decoded["object"] != "chat.completion" {
		t.Fatalf("expected chat.completion object, got %#v", decoded["object"])
	}
	choices, ok := decoded["choices"].([]any)
	if !ok || len(choices) != 1 {
		t.Fatalf("expected one choice, got %#v", decoded["choices"])
	}
	choice, ok := choices[0].(map[string]any)
	if !ok {
		t.Fatalf("expected choice object, got %#v", choices[0])
	}
	message, ok := choice["message"].(map[string]any)
	if !ok {
		t.Fatalf("expected message object, got %#v", choice["message"])
	}
	if message["content"] != "hello from claude" {
		t.Fatalf("expected assistant content, got %#v", message["content"])
	}
}

func TestOpenAIHandler_ChatCompletions_RejectsDisallowedModel(t *testing.T) {
	t.Parallel()

	var hits atomic.Int64
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	runtimeAuth := &auth.Auth{
		ID:       "p1",
		Provider: "claude",
		Attributes: map[string]string{
			"api_key":  "provider-secret",
			"base_url": upstream.URL,
		},
	}
	claudeExec := executor.NewClaude(upstream.Client())
	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient(runtimeAuth.ID, runtimeAuth.Provider, []registry.ModelInfo{{ID: "claude-3-7-sonnet"}})
	runner := &testChatRunner{
		execute: func(ctx context.Context, model string, requestBody []byte) (*auth.Result, error) {
			result, err := claudeExec.Execute(ctx, requestBody, runtimeAuth)
			if err != nil {
				return nil, err
			}
			return &auth.Result{StatusCode: result.StatusCode, Body: result.Body, Headers: result.Headers}, nil
		},
	}

	h := handlers.NewOpenAI(runner, modelRegistry)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	securedMux := api.APIKeyMiddleware([]string{"dev-key"}, mux)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", http.NoBody)
	req.Header.Set("authorization", "Bearer dev-key")
	req.Body = io.NopCloser(strings.NewReader(`{"model":"not-allowed","messages":[{"role":"user","content":"hello"}]}`))
	req.Header.Set("content-type", "application/json")
	rec := httptest.NewRecorder()

	securedMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 status, got %d body=%s", rec.Code, rec.Body.String())
	}
	if hits.Load() != 0 {
		t.Fatalf("expected zero upstream calls, got %d", hits.Load())
	}
	if !strings.Contains(rec.Body.String(), `"invalid_request_error"`) {
		t.Fatalf("expected structured invalid_request_error, got %s", rec.Body.String())
	}
}

func TestOpenAIHandler_ChatCompletions_TranslationValidationErrorIs4xx(t *testing.T) {
	t.Parallel()

	var hits atomic.Int64
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	runtimeAuth := &auth.Auth{
		ID:       "p1",
		Provider: "claude",
		Attributes: map[string]string{
			"api_key":  "provider-secret",
			"base_url": upstream.URL,
		},
	}
	claudeExec := executor.NewClaude(upstream.Client())
	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient(runtimeAuth.ID, runtimeAuth.Provider, []registry.ModelInfo{{ID: "claude-3-7-sonnet"}})
	runner := &testChatRunner{
		execute: func(ctx context.Context, model string, requestBody []byte) (*auth.Result, error) {
			result, err := claudeExec.Execute(ctx, requestBody, runtimeAuth)
			if err != nil {
				return nil, err
			}
			return &auth.Result{StatusCode: result.StatusCode, Body: result.Body, Headers: result.Headers}, nil
		},
	}

	h := handlers.NewOpenAI(runner, modelRegistry)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	securedMux := api.APIKeyMiddleware([]string{"dev-key"}, mux)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", http.NoBody)
	req.Header.Set("authorization", "Bearer dev-key")
	req.Body = io.NopCloser(strings.NewReader(`{
		"model":"claude-3-7-sonnet",
		"messages":[{"role":"user","content":[{"type":"text","text":"hello"}]}]
	}`))
	req.Header.Set("content-type", "application/json")
	rec := httptest.NewRecorder()

	securedMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 status, got %d body=%s", rec.Code, rec.Body.String())
	}
	if hits.Load() != 0 {
		t.Fatalf("expected zero upstream calls, got %d", hits.Load())
	}
	if !strings.Contains(rec.Body.String(), `"invalid_request_error"`) {
		t.Fatalf("expected structured invalid_request_error, got %s", rec.Body.String())
	}
}

func TestOpenAIHandler_ChatCompletions_RejectsInvalidRoleBeforeUpstream(t *testing.T) {
	t.Parallel()

	var hits atomic.Int64
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	runtimeAuth := &auth.Auth{
		ID:       "p1",
		Provider: "claude",
		Attributes: map[string]string{
			"api_key":  "provider-secret",
			"base_url": upstream.URL,
		},
	}
	claudeExec := executor.NewClaude(upstream.Client())
	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient(runtimeAuth.ID, runtimeAuth.Provider, []registry.ModelInfo{{ID: "claude-3-7-sonnet"}})
	runner := &testChatRunner{
		execute: func(ctx context.Context, model string, requestBody []byte) (*auth.Result, error) {
			result, err := claudeExec.Execute(ctx, requestBody, runtimeAuth)
			if err != nil {
				return nil, err
			}
			return &auth.Result{StatusCode: result.StatusCode, Body: result.Body, Headers: result.Headers}, nil
		},
	}

	h := handlers.NewOpenAI(runner, modelRegistry)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	securedMux := api.APIKeyMiddleware([]string{"dev-key"}, mux)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", http.NoBody)
	req.Header.Set("authorization", "Bearer dev-key")
	req.Body = io.NopCloser(strings.NewReader(`{
		"model":"claude-3-7-sonnet",
		"messages":[{"role":"tool","content":"bad role"}]
	}`))
	req.Header.Set("content-type", "application/json")
	rec := httptest.NewRecorder()

	securedMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 status, got %d body=%s", rec.Code, rec.Body.String())
	}
	if hits.Load() != 0 {
		t.Fatalf("expected zero upstream calls, got %d", hits.Load())
	}
}

func TestOpenAIHandler_ChatCompletions_RejectsEmptyEffectiveMessagesBeforeUpstream(t *testing.T) {
	t.Parallel()

	var hits atomic.Int64
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	runtimeAuth := &auth.Auth{
		ID:       "p1",
		Provider: "claude",
		Attributes: map[string]string{
			"api_key":  "provider-secret",
			"base_url": upstream.URL,
		},
	}
	claudeExec := executor.NewClaude(upstream.Client())
	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient(runtimeAuth.ID, runtimeAuth.Provider, []registry.ModelInfo{{ID: "claude-3-7-sonnet"}})
	runner := &testChatRunner{
		execute: func(ctx context.Context, model string, requestBody []byte) (*auth.Result, error) {
			result, err := claudeExec.Execute(ctx, requestBody, runtimeAuth)
			if err != nil {
				return nil, err
			}
			return &auth.Result{StatusCode: result.StatusCode, Body: result.Body, Headers: result.Headers}, nil
		},
	}

	h := handlers.NewOpenAI(runner, modelRegistry)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	securedMux := api.APIKeyMiddleware([]string{"dev-key"}, mux)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", http.NoBody)
	req.Header.Set("authorization", "Bearer dev-key")
	req.Body = io.NopCloser(strings.NewReader(`{
		"model":"claude-3-7-sonnet",
		"messages":[{"role":"system","content":"instructions only"}]
	}`))
	req.Header.Set("content-type", "application/json")
	rec := httptest.NewRecorder()

	securedMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 status, got %d body=%s", rec.Code, rec.Body.String())
	}
	if hits.Load() != 0 {
		t.Fatalf("expected zero upstream calls, got %d", hits.Load())
	}
}

func TestOpenAIHandler_ChatCompletions_UpstreamNon2xxUsesErrorPath(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"type":"error","error":{"message":"rate limited"}}`))
	}))
	defer upstream.Close()

	runtimeAuth := &auth.Auth{
		ID:       "p1",
		Provider: "claude",
		Attributes: map[string]string{
			"api_key":  "provider-secret",
			"base_url": upstream.URL,
		},
	}
	claudeExec := executor.NewClaude(upstream.Client())
	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient(runtimeAuth.ID, runtimeAuth.Provider, []registry.ModelInfo{{ID: "claude-3-7-sonnet"}})
	runner := &testChatRunner{
		execute: func(ctx context.Context, model string, requestBody []byte) (*auth.Result, error) {
			result, err := claudeExec.Execute(ctx, requestBody, runtimeAuth)
			if err != nil {
				return nil, err
			}
			return &auth.Result{StatusCode: result.StatusCode, Body: result.Body, Headers: result.Headers}, nil
		},
	}

	h := handlers.NewOpenAI(runner, modelRegistry)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	securedMux := api.APIKeyMiddleware([]string{"dev-key"}, mux)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", http.NoBody)
	req.Header.Set("authorization", "Bearer dev-key")
	req.Body = io.NopCloser(strings.NewReader(`{"model":"claude-3-7-sonnet","messages":[{"role":"user","content":"hello"}]}`))
	req.Header.Set("content-type", "application/json")
	rec := httptest.NewRecorder()

	securedMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 status from upstream error, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"api_error"`) {
		t.Fatalf("expected api_error response type, got %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"chat.completion"`) {
		t.Fatalf("expected error path, got success payload: %s", rec.Body.String())
	}
}

func TestOpenAIHandler_ChatCompletions_BodyTooLargeReturns4xx(t *testing.T) {
	t.Parallel()

	var executed atomic.Bool
	runner := &testChatRunner{
		execute: func(ctx context.Context, model string, requestBody []byte) (*auth.Result, error) {
			executed.Store(true)
			return &auth.Result{
				StatusCode: http.StatusOK,
				Body:       []byte(`{"object":"chat.completion","choices":[{"message":{"content":"ok"}}]}`),
				Headers:    http.Header{"content-type": []string{"application/json"}},
			}, nil
		},
	}
	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient("p1", "claude", []registry.ModelInfo{{ID: "claude-3-7-sonnet"}})

	h := handlers.NewOpenAI(runner, modelRegistry)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	securedMux := api.APIKeyMiddleware([]string{"dev-key"}, mux)

	oversizedPrompt := strings.Repeat("x", 5*1024*1024)
	requestBody := `{"model":"claude-3-7-sonnet","messages":[{"role":"user","content":"` + oversizedPrompt + `"}]}`

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", http.NoBody)
	req.Header.Set("authorization", "Bearer dev-key")
	req.Body = io.NopCloser(strings.NewReader(requestBody))
	req.Header.Set("content-type", "application/json")
	rec := httptest.NewRecorder()

	securedMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 status, got %d body=%s", rec.Code, rec.Body.String())
	}
	if executed.Load() {
		t.Fatal("expected runner execute not to be called for oversized body")
	}
	if !strings.Contains(rec.Body.String(), `"invalid_request_error"`) {
		t.Fatalf("expected structured invalid_request_error, got %s", rec.Body.String())
	}
}

func TestOpenAIHandler_Models_ReflectsRegistryUpdatesAfterHandlerCreation(t *testing.T) {
	t.Parallel()

	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient("p1", "claude", []registry.ModelInfo{{ID: "claude-3-7-sonnet"}})
	h := handlers.NewOpenAI(nil, modelRegistry)
	modelRegistry.RegisterClient("p2", "openai", []registry.ModelInfo{{ID: "gpt-4o"}})

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/v1/models", http.NoBody)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 status, got %d body=%s", rec.Code, rec.Body.String())
	}

	var decoded struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	gotIDs := make([]string, 0, len(decoded.Data))
	for _, m := range decoded.Data {
		gotIDs = append(gotIDs, m.ID)
	}
	sort.Strings(gotIDs)
	wantIDs := []string{"claude-3-7-sonnet", "gpt-4o"}
	if strings.Join(gotIDs, ",") != strings.Join(wantIDs, ",") {
		t.Fatalf("unexpected model ids: got=%v want=%v", gotIDs, wantIDs)
	}
}

func TestOpenAIHandler_Models_ReturnsEscapedJSON(t *testing.T) {
	t.Parallel()

	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient("p1", "claude", []registry.ModelInfo{{ID: `claude-"3-sonnet"`}})
	h := handlers.NewOpenAI(nil, modelRegistry)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/v1/models", http.NoBody)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 status, got %d body=%s", rec.Code, rec.Body.String())
	}

	var decoded struct {
		Object string `json:"object"`
		Data   []struct {
			ID     string `json:"id"`
			Object string `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("expected valid JSON models payload, got error: %v body=%s", err, rec.Body.String())
	}
	if decoded.Object != "list" {
		t.Fatalf("expected list object, got %q", decoded.Object)
	}
	if len(decoded.Data) != 1 {
		t.Fatalf("expected one model entry, got %d", len(decoded.Data))
	}
	if decoded.Data[0].ID != `claude-"3-sonnet"` {
		t.Fatalf("expected escaped model id roundtrip, got %q", decoded.Data[0].ID)
	}
}

func TestOpenAIHandler_Models_FiltersByRunnerSupportedModels(t *testing.T) {
	t.Parallel()

	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient("p1", "claude", []registry.ModelInfo{{ID: "model-active"}, {ID: "model-inactive"}})

	runner := &stateAwareRunner{
		testChatRunner: testChatRunner{
			execute: func(ctx context.Context, model string, requestBody []byte) (*auth.Result, error) {
				return &auth.Result{StatusCode: http.StatusOK, Body: []byte(`{}`), Headers: http.Header{}}, nil
			},
		},
		supportsModel: func(model string) bool {
			return model == "model-active"
		},
	}

	h := handlers.NewOpenAI(runner, modelRegistry)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/v1/models", http.NoBody)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 status, got %d body=%s", rec.Code, rec.Body.String())
	}

	var decoded struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(decoded.Data) != 1 || decoded.Data[0].ID != "model-active" {
		t.Fatalf("expected only active model, got %#v", decoded.Data)
	}
}

func TestOpenAIHandler_ChatCompletions_RejectsModelUnsupportedByRunnerState(t *testing.T) {
	t.Parallel()

	var executed atomic.Bool
	runner := &stateAwareRunner{
		testChatRunner: testChatRunner{
			execute: func(ctx context.Context, model string, requestBody []byte) (*auth.Result, error) {
				executed.Store(true)
				return &auth.Result{
					StatusCode: http.StatusOK,
					Body:       []byte(`{"object":"chat.completion","choices":[{"message":{"content":"ok"}}]}`),
					Headers:    http.Header{"content-type": []string{"application/json"}},
				}, nil
			},
		},
		supportsModel: func(model string) bool {
			return false
		},
	}
	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient("p1", "claude", []registry.ModelInfo{{ID: "claude-3-7-sonnet"}})
	h := handlers.NewOpenAI(runner, modelRegistry)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	securedMux := api.APIKeyMiddleware([]string{"dev-key"}, mux)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", http.NoBody)
	req.Header.Set("authorization", "Bearer dev-key")
	req.Body = io.NopCloser(strings.NewReader(`{"model":"claude-3-7-sonnet","messages":[{"role":"user","content":"hello"}]}`))
	req.Header.Set("content-type", "application/json")
	rec := httptest.NewRecorder()

	securedMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 status, got %d body=%s", rec.Code, rec.Body.String())
	}
	if executed.Load() {
		t.Fatal("expected runner execute not to be called when model unsupported by runner state")
	}
}

func TestOpenAIHandler_ChatCompletions_NilRunnerUsesExplicitUpstreamErrorPath(t *testing.T) {
	t.Parallel()

	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient("p1", "claude", []registry.ModelInfo{{ID: "claude-3-7-sonnet"}})
	h := handlers.NewOpenAI(nil, modelRegistry)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	securedMux := api.APIKeyMiddleware([]string{"dev-key"}, mux)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", http.NoBody)
	req.Header.Set("authorization", "Bearer dev-key")
	req.Body = io.NopCloser(strings.NewReader(`{"model":"claude-3-7-sonnet","messages":[{"role":"user","content":"hello"}]}`))
	req.Header.Set("content-type", "application/json")
	rec := httptest.NewRecorder()

	securedMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 status, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"api_error"`) {
		t.Fatalf("expected api_error response type, got %s", rec.Body.String())
	}
}

func TestOpenAIHandler_ChatCompletions_PassesParsedModelToRunner(t *testing.T) {
	t.Parallel()

	var gotModel atomic.Value
	runner := &testChatRunner{
		execute: func(ctx context.Context, model string, requestBody []byte) (*auth.Result, error) {
			gotModel.Store(model)
			return &auth.Result{
				StatusCode: http.StatusOK,
				Body:       []byte(`{"object":"chat.completion","choices":[{"message":{"content":"ok"}}]}`),
				Headers:    http.Header{"content-type": []string{"application/json"}},
			}, nil
		},
	}
	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient("p1", "claude", []registry.ModelInfo{{ID: "claude-3-7-sonnet"}})

	h := handlers.NewOpenAI(runner, modelRegistry)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	securedMux := api.APIKeyMiddleware([]string{"dev-key"}, mux)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", http.NoBody)
	req.Header.Set("authorization", "Bearer dev-key")
	req.Body = io.NopCloser(strings.NewReader(`{"model":"claude-3-7-sonnet","messages":[{"role":"user","content":"hello"}]}`))
	req.Header.Set("content-type", "application/json")
	rec := httptest.NewRecorder()

	securedMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 status, got %d body=%s", rec.Code, rec.Body.String())
	}
	if got, _ := gotModel.Load().(string); got != "claude-3-7-sonnet" {
		t.Fatalf("expected runner to receive parsed model, got %q", got)
	}
}

func TestServerBootstrap_Run_DoesNotFailInitializationForMultipleRuntimeAuths(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Host:    "bad:host",
		Port:    18080,
		APIKeys: []string{"dev-key"},
		Providers: []config.Provider{
			{
				ID:       "p1",
				Provider: "claude",
				APIKey:   "provider-secret-1",
				BaseURL:  "https://example.invalid",
				Models:   []string{"claude-3-7-sonnet"},
			},
			{
				ID:       "p2",
				Provider: "claude",
				APIKey:   "provider-secret-2",
				BaseURL:  "https://example.invalid",
				Models:   []string{"claude-3-7-sonnet"},
			},
		},
	}

	err := api.NewServer(cfg).Run()
	if err == nil {
		t.Fatal("expected run to fail to bind invalid listen address")
	}
	if strings.Contains(err.Error(), "exactly one runtime auth is required") {
		t.Fatalf("expected phase 1 single-auth bootstrap guard to be removed, got %v", err)
	}
}

func TestAPIKeyMiddleware_Unauthorized_UsesOpenAIErrorEnvelope(t *testing.T) {
	t.Parallel()

	modelRegistry := registry.NewModelRegistry()
	modelRegistry.RegisterClient("p1", "claude", []registry.ModelInfo{{ID: "claude-3-7-sonnet"}})
	h := handlers.NewOpenAI(nil, modelRegistry)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	securedMux := api.APIKeyMiddleware([]string{"dev-key"}, mux)

	req := httptest.NewRequest(http.MethodGet, "/v1/models", http.NoBody)
	req.Header.Set("authorization", "Bearer wrong-key")
	rec := httptest.NewRecorder()

	securedMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 status, got %d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("content-type"); got != "application/json" {
		t.Fatalf("expected application/json content-type, got %q", got)
	}
	if !strings.Contains(rec.Body.String(), `"error":{"message":"unauthorized","type":"invalid_request_error"}`) {
		t.Fatalf("expected openai-style error envelope, got %s", rec.Body.String())
	}
}
