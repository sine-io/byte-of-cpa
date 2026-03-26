package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	const yamlContent = `host: 127.0.0.1
port: 8317
api_keys:
  - dev-key
providers:
  - id: provider-1
    provider: claude
    api_key: test-provider-key
    base_url: https://api.example.com
    models:
      - claude-3-7-sonnet
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Port != 8317 {
		t.Fatalf("expected port 8317, got %d", cfg.Port)
	}

	if len(cfg.APIKeys) != 1 || cfg.APIKeys[0] != "dev-key" {
		t.Fatalf("expected one api key dev-key, got %#v", cfg.APIKeys)
	}

	if len(cfg.Providers) != 1 {
		t.Fatalf("expected one provider, got %d", len(cfg.Providers))
	}

	if cfg.Providers[0].Provider != "claude" {
		t.Fatalf("expected provider claude, got %q", cfg.Providers[0].Provider)
	}
}

func TestLoad_MultipleProviders(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	const yamlContent = `host: 127.0.0.1
port: 8317
api_keys:
  - dev-key
providers:
  - id: provider-1
    provider: claude
    api_key: claude-key
    base_url: https://api.anthropic.com
    models:
      - claude-3-7-sonnet
  - id: provider-2
    provider: claude
    api_key: claude-key-2
    base_url: https://api.anthropic.com
    models:
      - claude-3-5-haiku
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if len(cfg.Providers) != 2 {
		t.Fatalf("expected two providers, got %d", len(cfg.Providers))
	}

	if cfg.Providers[1].Provider != "claude" {
		t.Fatalf("expected second provider claude, got %q", cfg.Providers[1].Provider)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	const yamlContent = `host: 127.0.0.1
port: 8317
api_keys:
  - dev-key
providers:
  - id: provider-1
    provider: claude
    api_key: test-key
    base_url: [unterminated
    models:
      - claude-3-7-sonnet
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected parse error for malformed yaml, got nil")
	}
}

func TestLoad_ProviderValidation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	const yamlContent = `host: 127.0.0.1
port: 8317
api_keys:
  - dev-key
providers:
  - id: provider-1
    provider: claude
    models:
      - claude-3-7-sonnet
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected validation error for incomplete provider config, got nil")
	}

	if !strings.Contains(err.Error(), "provider") {
		t.Fatalf("expected provider validation error, got %v", err)
	}
}

func TestLoad_HostRequired(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	const yamlContent = `host: ""
port: 8317
api_keys:
  - dev-key
providers:
  - id: provider-1
    provider: claude
    api_key: test-key
    base_url: https://api.example.com
    models:
      - claude-3-7-sonnet
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected validation error for missing host, got nil")
	}

	if !strings.Contains(err.Error(), "host") {
		t.Fatalf("expected host validation error, got %v", err)
	}
}

func TestLoad_PortValidation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	const yamlContent = `host: 127.0.0.1
port: 0
api_keys:
  - dev-key
providers:
  - id: provider-1
    provider: claude
    api_key: test-key
    base_url: https://api.example.com
    models:
      - claude-3-7-sonnet
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected validation error for invalid port, got nil")
	}

	if !strings.Contains(err.Error(), "port") {
		t.Fatalf("expected port validation error, got %v", err)
	}
}

func TestLoad_RequiresAtLeastOneProvider(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	const yamlContent = `host: 127.0.0.1
port: 8317
api_keys:
  - dev-key
providers: []
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected validation error for zero providers, got nil")
	}

	if !strings.Contains(err.Error(), "providers") {
		t.Fatalf("expected providers validation error, got %v", err)
	}
}

func TestLoad_DuplicateProviderIDs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	const yamlContent = `host: 127.0.0.1
port: 8317
api_keys:
  - dev-key
providers:
  - id: duplicate-id
    provider: claude
    api_key: claude-key
    base_url: https://api.anthropic.com
    models:
      - claude-3-7-sonnet
  - id: duplicate-id
    provider: claude
    api_key: claude-key-2
    base_url: https://api.anthropic.com
    models:
      - claude-3-5-haiku
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected validation error for duplicate provider ids, got nil")
	}

	if !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate provider id error, got %v", err)
	}
}

func TestLoad_UnsupportedProviderRejected(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	const yamlContent = `host: 127.0.0.1
port: 8317
api_keys:
  - dev-key
providers:
  - id: provider-1
    provider: openai
    api_key: openai-key
    base_url: https://api.openai.com
    models:
      - gpt-4.1
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected validation error for unsupported provider, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported provider") {
		t.Fatalf("expected unsupported provider validation error, got %v", err)
	}
}

func TestLoad_ProviderBaseURLRejectsTrailingBareQuestionMark(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	const yamlContent = `host: 127.0.0.1
port: 8317
api_keys:
  - dev-key
providers:
  - id: provider-1
    provider: claude
    api_key: claude-key
    base_url: https://api.anthropic.com?
    models:
      - claude-3-7-sonnet
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected validation error for base_url ending in bare ?, got nil")
	}
	if !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url validation error, got %v", err)
	}
}

func TestLoad_ProviderBaseURLRejectsTrailingBareHash(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	const yamlContent = `host: 127.0.0.1
port: 8317
api_keys:
  - dev-key
providers:
  - id: provider-1
    provider: claude
    api_key: claude-key
    base_url: https://api.anthropic.com#
    models:
      - claude-3-7-sonnet
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected validation error for base_url ending in bare #, got nil")
	}
	if !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url validation error, got %v", err)
	}
}

func TestLoad_BaseURLValidation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testCases := []struct {
		name    string
		baseURL string
	}{
		{
			name:    "not a url",
			baseURL: "not-a-url",
		},
		{
			name:    "relative path",
			baseURL: "/v1/messages",
		},
		{
			name:    "missing scheme",
			baseURL: "api.anthropic.com",
		},
		{
			name:    "unsupported scheme",
			baseURL: "ws://api.anthropic.com",
		},
		{
			name:    "has query",
			baseURL: "https://api.anthropic.com?foo=bar",
		},
		{
			name:    "has fragment",
			baseURL: "https://api.anthropic.com#frag",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			configPath := filepath.Join(tmpDir, strings.ReplaceAll(tc.name, " ", "-")+".yaml")
			yamlContent := `host: 127.0.0.1
port: 8317
api_keys:
  - dev-key
providers:
  - id: provider-1
    provider: claude
    api_key: claude-key
    base_url: ` + tc.baseURL + `
    models:
      - claude-3-7-sonnet
`

			if err := os.WriteFile(configPath, []byte(yamlContent), 0o600); err != nil {
				t.Fatalf("write temp config: %v", err)
			}

			_, err := Load(configPath)
			if err == nil {
				t.Fatal("expected validation error for invalid base_url, got nil")
			}
			if !strings.Contains(err.Error(), "base_url") {
				t.Fatalf("expected base_url validation error, got %v", err)
			}
		})
	}
}

func TestLoad_ProviderWhitespaceFieldsRejected(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testCases := []struct {
		name          string
		providerBlock string
		wantErrorPart string
	}{
		{
			name: "id",
			providerBlock: `  - id: "   "
    provider: claude
    api_key: test-key
    base_url: https://api.example.com
    models:
      - claude-3-7-sonnet
`,
			wantErrorPart: "id",
		},
		{
			name: "provider",
			providerBlock: `  - id: provider-1
    provider: "   "
    api_key: test-key
    base_url: https://api.example.com
    models:
      - claude-3-7-sonnet
`,
			wantErrorPart: "provider",
		},
		{
			name: "api_key",
			providerBlock: `  - id: provider-1
    provider: claude
    api_key: "   "
    base_url: https://api.example.com
    models:
      - claude-3-7-sonnet
`,
			wantErrorPart: "api_key",
		},
		{
			name: "base_url",
			providerBlock: `  - id: provider-1
    provider: claude
    api_key: test-key
    base_url: "   "
    models:
      - claude-3-7-sonnet
`,
			wantErrorPart: "base_url",
		},
		{
			name: "model entry",
			providerBlock: `  - id: provider-1
    provider: claude
    api_key: test-key
    base_url: https://api.example.com
    models:
      - "   "
`,
			wantErrorPart: "models",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			configPath := filepath.Join(tmpDir, tc.name+"-config.yaml")
			yamlContent := `host: 127.0.0.1
port: 8317
api_keys:
  - dev-key
providers:
` + tc.providerBlock

			if err := os.WriteFile(configPath, []byte(yamlContent), 0o600); err != nil {
				t.Fatalf("write temp config: %v", err)
			}

			_, err := Load(configPath)
			if err == nil {
				t.Fatal("expected validation error for whitespace-only value, got nil")
			}

			if !strings.Contains(err.Error(), tc.wantErrorPart) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErrorPart, err)
			}
		})
	}
}

func TestLoad_TrimsReturnedValues(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	const yamlContent = `host: " 127.0.0.1 "
port: 8317
api_keys:
  - " dev-key "
providers:
  - id: " provider-1 "
    provider: " claude "
    api_key: " test-key "
    base_url: " https://api.example.com "
    models:
      - " claude-3-7-sonnet "
      - " claude-3-5-haiku "
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Host != "127.0.0.1" {
		t.Fatalf("expected trimmed host, got %q", cfg.Host)
	}

	provider := cfg.Providers[0]
	if provider.ID != "provider-1" {
		t.Fatalf("expected trimmed provider id, got %q", provider.ID)
	}
	if provider.Provider != "claude" {
		t.Fatalf("expected trimmed provider type, got %q", provider.Provider)
	}
	if provider.APIKey != "test-key" {
		t.Fatalf("expected trimmed api_key, got %q", provider.APIKey)
	}
	if provider.BaseURL != "https://api.example.com" {
		t.Fatalf("expected trimmed base_url, got %q", provider.BaseURL)
	}
	if provider.Models[0] != "claude-3-7-sonnet" {
		t.Fatalf("expected trimmed model, got %q", provider.Models[0])
	}
}

func TestLoad_RejectsMultiDocumentYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	const yamlContent = `host: 127.0.0.1
port: 8317
providers:
  - id: provider-1
    provider: claude
    api_key: test-key
    base_url: https://api.example.com
    models:
      - claude-3-7-sonnet
---
host: 0.0.0.0
port: 9000
providers:
  - id: provider-2
    provider: openai
    api_key: second-key
    base_url: https://api.openai.com
    models:
      - gpt-4.1
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected multi-document yaml error, got nil")
	}

	if !strings.Contains(err.Error(), "multiple YAML documents") {
		t.Fatalf("expected multi-document validation error, got %v", err)
	}
}

func TestLoad_RejectsWhitespaceOnlyAPIKeyEntries(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	const yamlContent = `host: 127.0.0.1
port: 8317
api_keys: [" dev-key ", "   "]
providers:
  - id: provider-1
    provider: claude
    api_key: provider-key
    base_url: https://api.example.com
    models:
      - claude-3-7-sonnet
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected api_keys validation error, got nil")
	}

	if !strings.Contains(err.Error(), "api_keys") {
		t.Fatalf("expected api_keys validation error, got %v", err)
	}
}

func TestLoad_TrimsAPIKeyEntries(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	const yamlContent = `host: 127.0.0.1
port: 8317
api_keys: [" dev-key ", " qa-key "]
providers:
  - id: provider-1
    provider: claude
    api_key: provider-key
    base_url: https://api.example.com
    models:
      - claude-3-7-sonnet
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if len(cfg.APIKeys) != 2 || cfg.APIKeys[0] != "dev-key" || cfg.APIKeys[1] != "qa-key" {
		t.Fatalf("expected trimmed api_keys, got %#v", cfg.APIKeys)
	}
}

func TestLoad_RequiresAtLeastOneAPIKey(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	const yamlContent = `host: 127.0.0.1
port: 8317
api_keys: []
providers:
  - id: provider-1
    provider: claude
    api_key: provider-key
    base_url: https://api.example.com
    models:
      - claude-3-7-sonnet
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("expected validation error for empty api_keys, got nil")
	}
	if !strings.Contains(err.Error(), "api_keys") {
		t.Fatalf("expected api_keys validation error, got %v", err)
	}
}
