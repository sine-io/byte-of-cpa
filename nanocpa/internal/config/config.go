package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Host      string     `yaml:"host"`
	Port      int        `yaml:"port"`
	APIKeys   []string   `yaml:"api_keys"`
	Providers []Provider `yaml:"providers"`
}

type Provider struct {
	ID       string   `yaml:"id"`
	Provider string   `yaml:"provider"`
	APIKey   string   `yaml:"api_key"`
	BaseURL  string   `yaml:"base_url"`
	Models   []string `yaml:"models"`
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

	return &cfg, nil
}

func normalizeConfig(cfg *Config) {
	cfg.Host = strings.TrimSpace(cfg.Host)
	for i := range cfg.APIKeys {
		cfg.APIKeys[i] = strings.TrimSpace(cfg.APIKeys[i])
	}
	for i := range cfg.Providers {
		p := &cfg.Providers[i]
		p.ID = strings.TrimSpace(p.ID)
		p.Provider = strings.ToLower(strings.TrimSpace(p.Provider))
		p.APIKey = strings.TrimSpace(p.APIKey)
		p.BaseURL = strings.TrimSpace(p.BaseURL)
		for j := range p.Models {
			p.Models[j] = strings.TrimSpace(p.Models[j])
		}
	}
}

func validateConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is required")
	}
	if cfg.Host == "" {
		return fmt.Errorf("host is required")
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if len(cfg.APIKeys) == 0 {
		return fmt.Errorf("api_keys must contain at least one key")
	}
	for i, key := range cfg.APIKeys {
		if key == "" {
			return fmt.Errorf("api_keys[%d] is required", i)
		}
	}
	if len(cfg.Providers) == 0 {
		return fmt.Errorf("providers must contain at least one provider")
	}
	seenProviderIDs := make(map[string]int, len(cfg.Providers))
	for i, provider := range cfg.Providers {
		prefix := fmt.Sprintf("providers[%d]", i)
		if provider.ID == "" {
			return fmt.Errorf("%s.id is required", prefix)
		}
		if previousIndex, exists := seenProviderIDs[provider.ID]; exists {
			return fmt.Errorf("%s.id %q duplicates providers[%d].id", prefix, provider.ID, previousIndex)
		}
		seenProviderIDs[provider.ID] = i
		if provider.Provider == "" {
			return fmt.Errorf("%s.provider is required", prefix)
		}
		if provider.Provider != "claude" && provider.Provider != "openai" {
			return fmt.Errorf("%s.provider must be one of [claude openai]", prefix)
		}
		if provider.APIKey == "" {
			return fmt.Errorf("%s.api_key is required", prefix)
		}
		if provider.BaseURL == "" {
			return fmt.Errorf("%s.base_url is required", prefix)
		}
		if len(provider.Models) == 0 {
			return fmt.Errorf("%s.models must contain at least one model", prefix)
		}
		for j, model := range provider.Models {
			if model == "" {
				return fmt.Errorf("%s.models[%d] is required", prefix, j)
			}
		}
	}
	return nil
}
