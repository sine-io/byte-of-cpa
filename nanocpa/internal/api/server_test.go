package api

import (
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
