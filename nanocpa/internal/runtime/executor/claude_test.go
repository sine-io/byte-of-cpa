package executor

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/auth"
)

func TestClaudeExecutor_Execute_IncludesAPIKeyAndReturnsResponse(t *testing.T) {
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
		if got := r.Header.Get("content-type"); got != "application/json" {
			t.Fatalf("expected content-type application/json, got %q", got)
		}
		if got := r.Header.Get("anthropic-version"); got == "" {
			t.Fatalf("expected anthropic-version header to be set")
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read upstream request body: %v", err)
		}
		if !strings.Contains(string(body), `"model":"claude-3-7-sonnet"`) {
			t.Fatalf("expected translated body to include model, got %s", string(body))
		}

		w.Header().Set("x-request-id", "req-1")
		w.Header().Set("content-type", "text/plain")
		w.Header().Set("x-upstream-debug", "raw")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id":"msg_1",
			"model":"claude-3-7-sonnet",
			"content":[{"type":"text","text":"hello from claude"}]
		}`))
	}))
	defer upstream.Close()

	providerAuth := &auth.Auth{
		ID:       "p1",
		Provider: "claude",
		Attributes: map[string]string{
			"api_key":  "provider-secret",
			"base_url": upstream.URL,
		},
	}
	exec := NewClaude(upstream.Client())

	openAIRequest := []byte(`{
		"model":"claude-3-7-sonnet",
		"messages":[{"role":"user","content":"hello"}]
	}`)

	result, err := exec.Execute(context.Background(), openAIRequest, providerAuth)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}

	if result.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 status, got %d", result.StatusCode)
	}
	if got := result.Headers.Get("x-request-id"); got != "req-1" {
		t.Fatalf("expected x-request-id header passthrough, got %q", got)
	}
	if got := result.Headers.Get("content-type"); got != "application/json" {
		t.Fatalf("expected translated content-type application/json, got %q", got)
	}
	if got := result.Headers.Get("x-upstream-debug"); got != "" {
		t.Fatalf("expected non-safe upstream headers to be dropped, got %q", got)
	}
	if !strings.Contains(string(result.Body), `"object":"chat.completion"`) {
		t.Fatalf("unexpected response body: %s", string(result.Body))
	}
}

func TestClaudeExecutor_Execute_Non2xxReturnsUpstreamError(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"type":"error","error":{"message":"bad request"}}`))
	}))
	defer upstream.Close()

	providerAuth := &auth.Auth{
		ID:       "p1",
		Provider: "claude",
		Attributes: map[string]string{
			"api_key":  "provider-secret",
			"base_url": upstream.URL,
		},
	}
	exec := NewClaude(upstream.Client())

	openAIRequest := []byte(`{
		"model":"claude-3-7-sonnet",
		"messages":[{"role":"user","content":"hello"}]
	}`)

	result, err := exec.Execute(context.Background(), openAIRequest, providerAuth)
	if err == nil {
		t.Fatal("expected upstream error for non-2xx response")
	}
	if result != nil {
		t.Fatalf("expected nil result on non-2xx upstream response, got %#v", result)
	}

	var upstreamErr *UpstreamError
	if !AsUpstreamError(err, &upstreamErr) {
		t.Fatalf("expected UpstreamError, got %T", err)
	}
	if upstreamErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 upstream status, got %d", upstreamErr.StatusCode)
	}
	if !strings.Contains(string(upstreamErr.Body), "bad request") {
		t.Fatalf("expected upstream error body to be preserved, got %s", string(upstreamErr.Body))
	}
}

func TestNewClaude_DefaultClientHasTimeout(t *testing.T) {
	t.Parallel()

	exec := NewClaude(nil)
	if exec.client == nil {
		t.Fatal("expected default client to be initialized")
	}
	if exec.client.Timeout <= 0 {
		t.Fatalf("expected positive default timeout, got %s", exec.client.Timeout)
	}
}

func TestNewClaude_ProvidedClientWithoutTimeoutGetsSafeTimeout(t *testing.T) {
	t.Parallel()

	client := &http.Client{}
	exec := NewClaude(client)
	if exec.client == nil {
		t.Fatal("expected client to be initialized")
	}
	if exec.client.Timeout <= 0 {
		t.Fatalf("expected positive timeout, got %s", exec.client.Timeout)
	}
	if exec.client.Timeout > 120*time.Second {
		t.Fatalf("unexpectedly large timeout: %s", exec.client.Timeout)
	}
}
