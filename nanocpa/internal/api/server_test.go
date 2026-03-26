package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/config"
)

func TestServer_ChatCompletions_AlternatesAcrossTwoConfiguredAuths(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	seenAPIKeys := make([]string, 0, 4)

	upstream1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seenAPIKeys = append(seenAPIKeys, r.Header.Get("x-api-key"))
		mu.Unlock()
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"u1","model":"claude-3-7-sonnet","content":[{"type":"text","text":"from-upstream-1"}]}`))
	}))
	defer upstream1.Close()

	upstream2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seenAPIKeys = append(seenAPIKeys, r.Header.Get("x-api-key"))
		mu.Unlock()
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"u2","model":"claude-3-7-sonnet","content":[{"type":"text","text":"from-upstream-2"}]}`))
	}))
	defer upstream2.Close()

	cfg := &config.Config{
		Host:    "127.0.0.1",
		Port:    18080,
		APIKeys: []string{"dev-key"},
		Providers: []config.Provider{
			{
				ID:       "p1",
				Provider: "claude",
				APIKey:   "provider-secret-1",
				BaseURL:  upstream1.URL,
				Models:   []string{"claude-3-7-sonnet"},
			},
			{
				ID:       "p2",
				Provider: "claude",
				APIKey:   "provider-secret-2",
				BaseURL:  upstream2.URL,
				Models:   []string{"claude-3-7-sonnet"},
			},
		},
	}

	server := NewServer(cfg)
	if server.initErr != nil {
		t.Fatalf("unexpected server init error: %v", server.initErr)
	}

	wantContentOrder := []string{"from-upstream-1", "from-upstream-2", "from-upstream-1", "from-upstream-2"}
	for i := 0; i < len(wantContentOrder); i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", io.NopCloser(strings.NewReader(`{"model":"claude-3-7-sonnet","messages":[{"role":"user","content":"hello"}]}`)))
		req.Header.Set("authorization", "Bearer dev-key")
		req.Header.Set("content-type", "application/json")
		rec := httptest.NewRecorder()

		server.handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d body=%s", i+1, rec.Code, rec.Body.String())
		}

		var payload struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("request %d: unmarshal response: %v body=%s", i+1, err, rec.Body.String())
		}
		if len(payload.Choices) != 1 {
			t.Fatalf("request %d: expected one choice, got %d", i+1, len(payload.Choices))
		}
		if payload.Choices[0].Message.Content != wantContentOrder[i] {
			t.Fatalf("request %d: unexpected content got=%q want=%q", i+1, payload.Choices[0].Message.Content, wantContentOrder[i])
		}
	}

	mu.Lock()
	defer mu.Unlock()
	wantAPIKeys := []string{"provider-secret-1", "provider-secret-2", "provider-secret-1", "provider-secret-2"}
	if len(seenAPIKeys) != len(wantAPIKeys) {
		t.Fatalf("unexpected upstream call count: got=%d want=%d", len(seenAPIKeys), len(wantAPIKeys))
	}
	for i := range wantAPIKeys {
		if seenAPIKeys[i] != wantAPIKeys[i] {
			t.Fatalf("unexpected key on call %d: got=%q want=%q", i+1, seenAPIKeys[i], wantAPIKeys[i])
		}
	}
}

func TestServer_BuildHTTPServer_UsesSafeTimeoutDefaults(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Host:    "127.0.0.1",
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
		},
	}

	server := NewServer(cfg)
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
	if httpServer.ReadHeaderTimeout < time.Second {
		t.Fatalf("expected read header timeout to be set, got %s", httpServer.ReadHeaderTimeout)
	}
	if httpServer.ReadTimeout < time.Second {
		t.Fatalf("expected read timeout to be set, got %s", httpServer.ReadTimeout)
	}
	if httpServer.WriteTimeout < time.Second {
		t.Fatalf("expected write timeout to be set, got %s", httpServer.WriteTimeout)
	}
	if httpServer.IdleTimeout < time.Second {
		t.Fatalf("expected idle timeout to be set, got %s", httpServer.IdleTimeout)
	}
}

func TestServer_BuildHTTPServer_FormatsIPv6ListenAddress(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Host:    "::1",
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
		},
	}

	server := NewServer(cfg)
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
