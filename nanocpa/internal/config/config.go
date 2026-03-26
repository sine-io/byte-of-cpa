package config

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Provider struct {
	ID       string   `yaml:"id"`
	Provider string   `yaml:"provider"`
	APIKey   string   `yaml:"api_key"`
	BaseURL  string   `yaml:"base_url"`
	Models   []string `yaml:"models"`
}

type Config struct {
	Host      string     `yaml:"host"`
	Port      int        `yaml:"port"`
	APIKeys   []string   `yaml:"api_keys"`
	Providers []Provider `yaml:"providers"`
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	decoder := yaml.NewDecoder(bytes.NewReader(b))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parse config yaml: %w", err)
	}
	var extraDoc yaml.Node
	if err := decoder.Decode(&extraDoc); err != io.EOF {
		if err != nil {
			return nil, fmt.Errorf("parse config yaml: %w", err)
		}
		return nil, fmt.Errorf("multiple YAML documents are not supported")
	}

	normalizeConfig(&cfg)

	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	if err := validateProviders(cfg.Providers); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func normalizeConfig(cfg *Config) {
	cfg.Host = strings.TrimSpace(cfg.Host)
	for i := range cfg.APIKeys {
		cfg.APIKeys[i] = strings.TrimSpace(cfg.APIKeys[i])
	}
	for i := range cfg.Providers {
		cfg.Providers[i].ID = strings.TrimSpace(cfg.Providers[i].ID)
		cfg.Providers[i].Provider = strings.TrimSpace(cfg.Providers[i].Provider)
		cfg.Providers[i].APIKey = strings.TrimSpace(cfg.Providers[i].APIKey)
		cfg.Providers[i].BaseURL = strings.TrimSpace(cfg.Providers[i].BaseURL)
		for j := range cfg.Providers[i].Models {
			cfg.Providers[i].Models[j] = strings.TrimSpace(cfg.Providers[i].Models[j])
		}
	}
}

func validateConfig(cfg *Config) error {
	if strings.TrimSpace(cfg.Host) == "" {
		return fmt.Errorf("host is required")
	}
	if len(cfg.APIKeys) == 0 {
		return fmt.Errorf("api_keys must contain at least one key")
	}
	for i, apiKey := range cfg.APIKeys {
		if apiKey == "" {
			return fmt.Errorf("api_keys[%d]: value is required", i+1)
		}
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if len(cfg.Providers) == 0 {
		return fmt.Errorf("providers must contain at least one provider")
	}
	return nil
}

func validateProviders(providers []Provider) error {
	seenIDs := make(map[string]struct{}, len(providers))

	for i, provider := range providers {
		idx := i + 1
		providerID := provider.ID
		if providerID == "" {
			return fmt.Errorf("provider[%d]: id is required", idx)
		}
		if _, exists := seenIDs[providerID]; exists {
			return fmt.Errorf("provider[%d]: duplicate provider id %q", idx, providerID)
		}
		seenIDs[providerID] = struct{}{}
		if provider.Provider == "" {
			return fmt.Errorf("provider[%d]: provider is required", idx)
		}
		if provider.Provider != "claude" {
			return fmt.Errorf("provider[%d]: unsupported provider %q (only \"claude\" is supported)", idx, provider.Provider)
		}
		if provider.APIKey == "" {
			return fmt.Errorf("provider[%d]: api_key is required", idx)
		}
		if provider.BaseURL == "" {
			return fmt.Errorf("provider[%d]: base_url is required", idx)
		}
		if strings.HasSuffix(provider.BaseURL, "?") || strings.HasSuffix(provider.BaseURL, "#") {
			return fmt.Errorf("provider[%d]: base_url must not include query or fragment", idx)
		}
		baseURL, err := url.Parse(provider.BaseURL)
		if err != nil || !baseURL.IsAbs() || baseURL.Host == "" {
			return fmt.Errorf("provider[%d]: base_url must be an absolute http/https URL", idx)
		}
		if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
			return fmt.Errorf("provider[%d]: base_url must be an absolute http/https URL", idx)
		}
		if baseURL.RawQuery != "" || baseURL.Fragment != "" {
			return fmt.Errorf("provider[%d]: base_url must not include query or fragment", idx)
		}
		if len(provider.Models) == 0 {
			return fmt.Errorf("provider[%d]: models is required", idx)
		}
		for _, model := range provider.Models {
			if model == "" {
				return fmt.Errorf("provider[%d]: models must not contain empty values", idx)
			}
		}
	}
	return nil
}
