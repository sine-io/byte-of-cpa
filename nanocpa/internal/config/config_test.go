package config

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestLoad_Chapter2Snapshot(t *testing.T) {
	t.Parallel()

	cfg, err := Load(chapter2ExampleConfigPath(t))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	want := &Config{
		Host:    "127.0.0.1",
		Port:    8317,
		APIKeys: []string{"dev-key"},
		Providers: []Provider{
			{
				ID:       "claude-primary",
				Provider: "claude",
				APIKey:   "your-claude-api-key",
				BaseURL:  "https://api.anthropic.com",
				Models:   []string{"claude-3-7-sonnet"},
			},
			{
				ID:       "openai-primary",
				Provider: "openai",
				APIKey:   "your-openai-api-key",
				BaseURL:  "https://api.openai.com",
				Models:   []string{"gpt-4o-mini"},
			},
		},
	}
	if !reflect.DeepEqual(cfg, want) {
		t.Fatalf("unexpected chapter 2 snapshot config: got=%+v want=%+v", cfg, want)
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

func TestLoad_NormalizesChapter2Fields(t *testing.T) {
	t.Parallel()

	cfg, err := loadConfig(t, "host: \" 127.0.0.1 \"\nport: 8317\napi_keys:\n  - \" down-key-1 \"\nproviders:\n  - id: \" provider-1 \"\n    provider: \" CLAUDE \"\n    api_key: \" up-key-1 \"\n    base_url: \" https://api.anthropic.com \"\n    models:\n      - \" claude-3-7-sonnet \"\n")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if !reflect.DeepEqual(cfg.APIKeys, []string{"down-key-1"}) {
		t.Fatalf("unexpected normalized api_keys: %v", cfg.APIKeys)
	}
	if len(cfg.Providers) != 1 {
		t.Fatalf("unexpected providers length: %d", len(cfg.Providers))
	}
	p := cfg.Providers[0]
	if p.ID != "provider-1" {
		t.Fatalf("unexpected normalized provider id: %q", p.ID)
	}
	if p.Provider != "claude" {
		t.Fatalf("unexpected normalized provider type: %q", p.Provider)
	}
	if p.APIKey != "up-key-1" {
		t.Fatalf("unexpected normalized provider api_key: %q", p.APIKey)
	}
	if p.BaseURL != "https://api.anthropic.com" {
		t.Fatalf("unexpected normalized provider base_url: %q", p.BaseURL)
	}
	if !reflect.DeepEqual(p.Models, []string{"claude-3-7-sonnet"}) {
		t.Fatalf("unexpected normalized provider models: %v", p.Models)
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

	_, err := loadConfig(t, "port: 8317\napi_keys:\n  - down-key-1\nproviders:\n  - id: claude-primary\n    provider: claude\n    api_key: up-key-1\n    base_url: https://api.anthropic.com\n    models:\n      - claude-3-7-sonnet\n")
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

func TestValidate_EmptyProviders(t *testing.T) {
	t.Parallel()

	_, err := loadConfig(t, "host: 127.0.0.1\nport: 8317\napi_keys:\n  - down-key-1\nproviders: []\n")
	if err == nil {
		t.Fatal("expected validation error for empty providers")
	}
	if !strings.Contains(err.Error(), "providers must contain at least one provider") {
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
			name:        "missing id",
			yamlContent: "host: 127.0.0.1\nport: 8317\napi_keys:\n  - down-key-1\nproviders:\n  - provider: claude\n    api_key: up-key-1\n    base_url: https://api.anthropic.com\n    models:\n      - claude-3-7-sonnet\n",
			wantErr:     "providers[0].id is required",
		},
		{
			name:        "missing provider",
			yamlContent: "host: 127.0.0.1\nport: 8317\napi_keys:\n  - down-key-1\nproviders:\n  - id: claude-primary\n    api_key: up-key-1\n    base_url: https://api.anthropic.com\n    models:\n      - claude-3-7-sonnet\n",
			wantErr:     "providers[0].provider is required",
		},
		{
			name:        "missing api_key",
			yamlContent: "host: 127.0.0.1\nport: 8317\napi_keys:\n  - down-key-1\nproviders:\n  - id: claude-primary\n    provider: claude\n    base_url: https://api.anthropic.com\n    models:\n      - claude-3-7-sonnet\n",
			wantErr:     "providers[0].api_key is required",
		},
		{
			name:        "missing base_url",
			yamlContent: "host: 127.0.0.1\nport: 8317\napi_keys:\n  - down-key-1\nproviders:\n  - id: claude-primary\n    provider: claude\n    api_key: up-key-1\n    models:\n      - claude-3-7-sonnet\n",
			wantErr:     "providers[0].base_url is required",
		},
		{
			name:        "missing models",
			yamlContent: "host: 127.0.0.1\nport: 8317\napi_keys:\n  - down-key-1\nproviders:\n  - id: claude-primary\n    provider: claude\n    api_key: up-key-1\n    base_url: https://api.anthropic.com\n",
			wantErr:     "providers[0].models must contain at least one model",
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

func TestValidate_DuplicateProviderIDs(t *testing.T) {
	t.Parallel()

	_, err := loadConfig(t, "host: 127.0.0.1\nport: 8317\napi_keys:\n  - down-key-1\nproviders:\n  - id: provider-1\n    provider: claude\n    api_key: up-key-1\n    base_url: https://api.anthropic.com\n    models:\n      - claude-3-7-sonnet\n  - id: \" provider-1 \"\n    provider: openai\n    api_key: up-key-2\n    base_url: https://api.openai.com\n    models:\n      - gpt-4o-mini\n")
	if err == nil {
		t.Fatal("expected validation error for duplicate provider ids")
	}
	if !strings.Contains(err.Error(), `providers[1].id "provider-1" duplicates providers[0].id`) {
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

func chapter2ExampleConfigPath(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "config.example.yaml"))
}
