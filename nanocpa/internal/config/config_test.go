package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_MinimalBootstrapConfig(t *testing.T) {
	t.Parallel()

	cfg, err := loadConfig(t, validChapter2ConfigYAML())
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

	cfg, err := loadConfig(t, "host: \" 127.0.0.1 \"\nport: 8317\napi_keys:\n  - down-key-1\nproviders:\n  - id: claude-primary\n    provider: claude\n    api_key: up-key-1\n    base_url: https://api.anthropic.com\n    models:\n      - claude-3-7-sonnet\n")
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

	_, err := loadConfig(t, "host: \"\"\nport: 8317\napi_keys:\n  - down-key-1\nproviders:\n  - id: claude-primary\n    provider: claude\n    api_key: up-key-1\n    base_url: https://api.anthropic.com\n    models:\n      - claude-3-7-sonnet\n")
	if err == nil {
		t.Fatal("expected validation error for missing host")
	}
	if !strings.Contains(err.Error(), "host") {
		t.Fatalf("expected host validation error, got %v", err)
	}
}

func TestLoad_PortValidation(t *testing.T) {
	t.Parallel()

	_, err := loadConfig(t, "host: 127.0.0.1\nport: 0\napi_keys:\n  - down-key-1\nproviders:\n  - id: claude-primary\n    provider: claude\n    api_key: up-key-1\n    base_url: https://api.anthropic.com\n    models:\n      - claude-3-7-sonnet\n")
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

func TestValidate_EmptyAPIKeys(t *testing.T) {
	t.Parallel()

	_, err := loadConfig(t, "host: 127.0.0.1\nport: 8317\napi_keys: []\nproviders:\n  - id: claude-primary\n    provider: claude\n    api_key: up-key-1\n    base_url: https://api.anthropic.com\n    models:\n      - claude-3-7-sonnet\n")
	if err == nil {
		t.Fatal("expected validation error for empty api_keys")
	}
	if !strings.Contains(err.Error(), "api_keys must contain at least one key") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_MissingProviderFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		yamlContent string
		wantErr     string
	}{
		{
			name: "missing id",
			yamlContent: "host: 127.0.0.1\nport: 8317\napi_keys:\n  - down-key-1\nproviders:\n  - id: \"\"\n    provider: claude\n    api_key: up-key-1\n    base_url: https://api.anthropic.com\n    models:\n      - claude-3-7-sonnet\n",
			wantErr: "providers[0].id is required",
		},
		{
			name: "missing provider",
			yamlContent: "host: 127.0.0.1\nport: 8317\napi_keys:\n  - down-key-1\nproviders:\n  - id: claude-primary\n    provider: \"\"\n    api_key: up-key-1\n    base_url: https://api.anthropic.com\n    models:\n      - claude-3-7-sonnet\n",
			wantErr: "providers[0].provider is required",
		},
		{
			name: "missing api_key",
			yamlContent: "host: 127.0.0.1\nport: 8317\napi_keys:\n  - down-key-1\nproviders:\n  - id: claude-primary\n    provider: claude\n    api_key: \"\"\n    base_url: https://api.anthropic.com\n    models:\n      - claude-3-7-sonnet\n",
			wantErr: "providers[0].api_key is required",
		},
		{
			name: "missing base_url",
			yamlContent: "host: 127.0.0.1\nport: 8317\napi_keys:\n  - down-key-1\nproviders:\n  - id: claude-primary\n    provider: claude\n    api_key: up-key-1\n    base_url: \"\"\n    models:\n      - claude-3-7-sonnet\n",
			wantErr: "providers[0].base_url is required",
		},
		{
			name: "missing models",
			yamlContent: "host: 127.0.0.1\nport: 8317\napi_keys:\n  - down-key-1\nproviders:\n  - id: claude-primary\n    provider: claude\n    api_key: up-key-1\n    base_url: https://api.anthropic.com\n    models: []\n",
			wantErr: "providers[0].models must contain at least one model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := loadConfig(t, tt.yamlContent)
			if err == nil {
				t.Fatal("expected validation error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error to contain %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestValidate_UnsupportedProviderValue(t *testing.T) {
	t.Parallel()

	_, err := loadConfig(t, "host: 127.0.0.1\nport: 8317\napi_keys:\n  - down-key-1\nproviders:\n  - id: provider-1\n    provider: unknown\n    api_key: up-key-1\n    base_url: https://example.com\n    models:\n      - demo-model\n")
	if err == nil {
		t.Fatal("expected validation error for unsupported provider")
	}
	if !strings.Contains(err.Error(), "providers[0].provider must be one of [claude openai]") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func validChapter2ConfigYAML() string {
	return "host: 127.0.0.1\nport: 8317\napi_keys:\n  - down-key-1\n  - down-key-2\nproviders:\n  - id: claude-primary\n    provider: claude\n    api_key: up-key-1\n    base_url: https://api.anthropic.com\n    models:\n      - claude-3-7-sonnet\n"
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
