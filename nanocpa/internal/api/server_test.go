package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/config"
)

func TestServer_NewServer_RejectsNilConfig(t *testing.T) {
	t.Parallel()

	server := NewServer(nil)
	if server == nil {
		t.Fatal("expected server")
	}
	if server.initErr == nil {
		t.Fatal("expected init error for nil config")
	}
}

func TestServer_BuildHTTPServer_UsesSafeTimeoutDefaults(t *testing.T) {
	t.Parallel()

	server := NewServer(&config.Config{
		Host: "127.0.0.1",
		Port: 18080,
	})
	if server.initErr != nil {
		t.Fatalf("unexpected server init error: %v", server.initErr)
	}

	httpServer := server.buildHTTPServer()
	if httpServer == nil {
		t.Fatal("expected http server")
	}
	if httpServer.Addr != "127.0.0.1:18080" {
		t.Fatalf("unexpected server addr: %q", httpServer.Addr)
	}
	if httpServer.Handler == nil {
		t.Fatal("expected handler")
	}
	if httpServer.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("unexpected read header timeout: %s", httpServer.ReadHeaderTimeout)
	}
	if httpServer.ReadTimeout != 30*time.Second {
		t.Fatalf("unexpected read timeout: %s", httpServer.ReadTimeout)
	}
	if httpServer.WriteTimeout != 30*time.Second {
		t.Fatalf("unexpected write timeout: %s", httpServer.WriteTimeout)
	}
	if httpServer.IdleTimeout != 60*time.Second {
		t.Fatalf("unexpected idle timeout: %s", httpServer.IdleTimeout)
	}
}

func TestServer_BuildHTTPServer_FormatsIPv6ListenAddress(t *testing.T) {
	t.Parallel()

	server := NewServer(&config.Config{
		Host: "::1",
		Port: 18080,
	})
	if server.initErr != nil {
		t.Fatalf("unexpected server init error: %v", server.initErr)
	}

	httpServer := server.buildHTTPServer()
	if httpServer == nil {
		t.Fatal("expected http server")
	}
	if httpServer.Addr != "[::1]:18080" {
		t.Fatalf("unexpected server addr: %q", httpServer.Addr)
	}
}

func TestServer_RoutesSameModelAcrossMultipleClaudeProvidersRoundRobin(t *testing.T) {
	t.Parallel()

	var upstream1Calls int32
	upstream1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&upstream1Calls, 1)
		w.Header().Set("x-request-id", "provider-1")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"msg_1","model":"claude-3-5-haiku","content":[{"type":"text","text":"from provider 1"}]}`))
	}))
	defer upstream1.Close()

	var upstream2Calls int32
	upstream2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&upstream2Calls, 1)
		w.Header().Set("x-request-id", "provider-2")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"msg_2","model":"claude-3-5-haiku","content":[{"type":"text","text":"from provider 2"}]}`))
	}))
	defer upstream2.Close()

	server := NewServer(&config.Config{
		Host:    "127.0.0.1",
		Port:    18080,
		APIKeys: []string{"dev-key"},
		Providers: []config.Provider{
			{ID: "claude-2", Provider: "claude", APIKey: "key-2", BaseURL: upstream2.URL, Models: []string{"claude-3-5-haiku"}},
			{ID: "claude-1", Provider: "claude", APIKey: "key-1", BaseURL: upstream1.URL, Models: []string{"claude-3-5-haiku"}},
		},
	})

	first := performServerChatRequest(t, server, `{"model":"claude-3-5-haiku","messages":[{"role":"user","content":"hello"}]}`)
	second := performServerChatRequest(t, server, `{"model":"claude-3-5-haiku","messages":[{"role":"user","content":"hello again"}]}`)

	if first.Code != http.StatusOK {
		t.Fatalf("expected first request to succeed, got %d body=%s", first.Code, first.Body.String())
	}
	if second.Code != http.StatusOK {
		t.Fatalf("expected second request to succeed, got %d body=%s", second.Code, second.Body.String())
	}
	assertServerChatMessage(t, first, "claude-3-5-haiku", "from provider 1")
	assertServerChatMessage(t, second, "claude-3-5-haiku", "from provider 2")
	if got := atomic.LoadInt32(&upstream1Calls); got != 1 {
		t.Fatalf("expected one request to provider 1, got %d", got)
	}
	if got := atomic.LoadInt32(&upstream2Calls); got != 1 {
		t.Fatalf("expected one request to provider 2, got %d", got)
	}
}

func TestServer_RoundRobinStateIsolatedPerModel(t *testing.T) {
	t.Parallel()

	upstream1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"msg_1","model":"placeholder","content":[{"type":"text","text":"provider 1"}]}`))
	}))
	defer upstream1.Close()

	upstream2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"msg_2","model":"placeholder","content":[{"type":"text","text":"provider 2"}]}`))
	}))
	defer upstream2.Close()

	server := NewServer(&config.Config{
		Host:    "127.0.0.1",
		Port:    18080,
		APIKeys: []string{"dev-key"},
		Providers: []config.Provider{
			{ID: "claude-2", Provider: "claude", APIKey: "key-2", BaseURL: upstream2.URL, Models: []string{"claude-3-5-haiku", "claude-3-7-sonnet"}},
			{ID: "claude-1", Provider: "claude", APIKey: "key-1", BaseURL: upstream1.URL, Models: []string{"claude-3-5-haiku", "claude-3-7-sonnet"}},
		},
	})

	firstHaiku := performServerChatRequest(t, server, `{"model":"claude-3-5-haiku","messages":[{"role":"user","content":"haiku 1"}]}`)
	secondHaiku := performServerChatRequest(t, server, `{"model":"claude-3-5-haiku","messages":[{"role":"user","content":"haiku 2"}]}`)
	firstSonnet := performServerChatRequest(t, server, `{"model":"claude-3-7-sonnet","messages":[{"role":"user","content":"sonnet 1"}]}`)

	assertServerChatMessage(t, firstHaiku, "placeholder", "provider 1")
	assertServerChatMessage(t, secondHaiku, "placeholder", "provider 2")
	assertServerChatMessage(t, firstSonnet, "placeholder", "provider 1")
}

func performServerChatRequest(t *testing.T, server *Server, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer dev-key")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	return rec
}

func assertServerChatMessage(t *testing.T, rec *httptest.ResponseRecorder, wantModel, wantContent string) {
	t.Helper()

	var response struct {
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Model != wantModel {
		t.Fatalf("expected model %q, got %q", wantModel, response.Model)
	}
	if len(response.Choices) != 1 {
		t.Fatalf("expected one choice, got %d", len(response.Choices))
	}
	if response.Choices[0].Message.Content != wantContent {
		t.Fatalf("expected message %q, got %q", wantContent, response.Choices[0].Message.Content)
	}
}
