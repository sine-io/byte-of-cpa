package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/api/handlers"
	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/auth"
	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/registry"
	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/runtime/executor"
)

type Server struct {
	config  *config.Config
	handler http.Handler
	initErr error
}

func NewServer(cfg *config.Config) *Server {
	s := &Server{config: cfg}

	if cfg == nil {
		s.initErr = errors.New("config is required")
		return s
	}
	if len(cfg.Providers) == 0 {
		s.initErr = errors.New("at least one provider is required")
		return s
	}

	auths := configToAuths(cfg)
	modelInfoByAuthID := configToModelInfosByAuthID(cfg)
	modelRegistry := registry.NewModelRegistry()
	for _, runtimeAuth := range auths {
		modelRegistry.RegisterClient(runtimeAuth.ID, runtimeAuth.Provider, modelInfoByAuthID[runtimeAuth.ID])
	}

	manager := auth.NewManager(modelRegistry, nil)
	for _, runtimeAuth := range auths {
		manager.RegisterAuth(runtimeAuth)
	}

	for _, provider := range cfg.Providers {
		switch provider.Provider {
		case "claude":
			manager.RegisterExecutor(provider.Provider, &executorAdapter{
				executor: executor.NewClaude(nil),
			})
		default:
			s.initErr = fmt.Errorf("unsupported provider: %s", provider.Provider)
			return s
		}
	}

	openAI := handlers.NewOpenAI(manager, modelRegistry)
	mux := http.NewServeMux()
	openAI.RegisterRoutes(mux)
	s.handler = APIKeyMiddleware(cfg.APIKeys, mux)
	return s
}

func (s *Server) Run() error {
	if s.initErr != nil {
		return s.initErr
	}
	return s.buildHTTPServer().ListenAndServe()
}

func configToAuths(cfg *config.Config) []*auth.Auth {
	auths := make([]*auth.Auth, 0, len(cfg.Providers))
	now := time.Now().UTC()
	for _, provider := range cfg.Providers {
		auths = append(auths, &auth.Auth{
			ID:       provider.ID,
			Provider: provider.Provider,
			Label:    provider.ID,
			Status:   auth.StatusActive,
			Attributes: map[string]string{
				"api_key":  provider.APIKey,
				"base_url": provider.BaseURL,
			},
			CreatedAt: now,
			UpdatedAt: now,
		})
	}
	return auths
}

func configToModelInfosByAuthID(cfg *config.Config) map[string][]registry.ModelInfo {
	result := make(map[string][]registry.ModelInfo, len(cfg.Providers))
	for _, provider := range cfg.Providers {
		models := make([]registry.ModelInfo, 0, len(provider.Models))
		for _, modelID := range provider.Models {
			models = append(models, registry.ModelInfo{ID: modelID})
		}
		result[provider.ID] = models
	}
	return result
}

type executorAdapter struct {
	executor executor.ChatCompletionsExecutor
}

func (e *executorAdapter) Execute(ctx context.Context, openAIRequest []byte, runtimeAuth *auth.Auth) (*auth.Result, error) {
	result, err := e.executor.Execute(ctx, openAIRequest, runtimeAuth)
	if err != nil {
		return nil, err
	}
	return &auth.Result{
		StatusCode: result.StatusCode,
		Body:       result.Body,
		Headers:    result.Headers,
	}, nil
}

func (s *Server) buildHTTPServer() *http.Server {
	return &http.Server{
		Addr:              net.JoinHostPort(s.config.Host, strconv.Itoa(s.config.Port)),
		Handler:           s.handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
