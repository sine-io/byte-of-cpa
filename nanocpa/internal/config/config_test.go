package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_MinimalBootstrapConfig(t *testing.T) {
	t.Parallel()

	cfg, err := loadConfig(t, "host: 127.0.0.1\nport: 8317\n")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Host != "127.0.0.1" {
		t.Fatalf("unexpected host: %q", cfg.Host)
	}
	if cfg.Port != 8317 {
		t.Fatalf("unexpected port: %d", cfg.Port)
	}
}

func TestLoad_TrimsHost(t *testing.T) {
	t.Parallel()

	cfg, err := loadConfig(t, "host: \" 127.0.0.1 \"\nport: 8317\n")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Host != "127.0.0.1" {
		t.Fatalf("unexpected host: %q", cfg.Host)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	t.Parallel()

	_, err := loadConfig(t, "host: 127.0.0.1\nport: [unterminated\n")
	if err == nil {
		t.Fatal("expected parse error for malformed yaml")
	}
}

func TestLoad_HostRequired(t *testing.T) {
	t.Parallel()

	_, err := loadConfig(t, "host: \"\"\nport: 8317\n")
	if err == nil {
		t.Fatal("expected validation error for missing host")
	}
	if !strings.Contains(err.Error(), "host") {
		t.Fatalf("expected host validation error, got %v", err)
	}
}

func TestLoad_PortValidation(t *testing.T) {
	t.Parallel()

	_, err := loadConfig(t, "host: 127.0.0.1\nport: 0\n")
	if err == nil {
		t.Fatal("expected validation error for invalid port")
	}
	if !strings.Contains(err.Error(), "port") {
		t.Fatalf("expected port validation error, got %v", err)
	}
}

func TestLoad_UnknownField(t *testing.T) {
	t.Parallel()

	_, err := loadConfig(t, "host: 127.0.0.1\nport: 8317\nbootstrap_mode: true\n")
	if err == nil {
		t.Fatal("expected parse error for unknown field")
	}
	if !strings.Contains(err.Error(), "field bootstrap_mode not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoad_MultipleDocuments(t *testing.T) {
	t.Parallel()

	_, err := loadConfig(t, "host: 127.0.0.1\nport: 8317\n---\nhost: 0.0.0.0\nport: 8318\n")
	if err == nil {
		t.Fatal("expected error for multiple yaml documents")
	}
	if !strings.Contains(err.Error(), "multiple YAML documents") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func loadConfig(t *testing.T, yamlContent string) (*Config, error) {
	t.Helper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	return Load(configPath)
}
