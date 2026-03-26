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
)

func TestChatCompletionsRouteExists(t *testing.T) {
	t.Parallel()

	rec := performAuthorizedRequest(t, http.MethodPost, "/v1/chat/completions", `{"model":"gpt-4o-mini","messages":[]}`)

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

func TestOpenAI_ServerRegistersRoutesBehindAccessMiddleware(t *testing.T) {
	t.Parallel()

	server := api.NewServer(&config.Config{
		Host:    "127.0.0.1",
		Port:    18080,
		APIKeys: []string{"dev-key"},
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
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Object != "list" {
		t.Fatalf("expected list object, got %q", response.Object)
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

	rec := performAuthorizedRequest(t, http.MethodPost, "/v1/chat/completions", `{"model":`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d body=%s", rec.Code, rec.Body.String())
	}

	assertOpenAIError(t, rec, "invalid_request_error", "invalid JSON request body")
}

func TestChatCompletionsMissingModelReturnsValidationError(t *testing.T) {
	t.Parallel()

	rec := performAuthorizedRequest(t, http.MethodPost, "/v1/chat/completions", `{"messages":[]}`)

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
		http.MethodPost,
		"/v1/chat/completions",
		`{"model":"gpt-4o-mini","messages":[{"role":"user","content":"`+oversizedPrompt+`"}]}`,
	)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 for oversized body, got %d body=%s", rec.Code, rec.Body.String())
	}

	assertOpenAIError(t, rec, "invalid_request_error", "request body too large")
}

func TestModelsReturnsOpenAIListShape(t *testing.T) {
	t.Parallel()

	rec := performAuthorizedRequest(t, http.MethodGet, "/v1/models", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for models list, got %d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Fatalf("expected JSON content type, got %q", got)
	}

	var response struct {
		Object string            `json:"object"`
		Data   []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Object != "list" {
		t.Fatalf("expected list object, got %q", response.Object)
	}
	if response.Data == nil {
		t.Fatal("expected data array")
	}
}

func performAuthorizedRequest(t *testing.T, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	openAI := handlers.NewOpenAI()
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
