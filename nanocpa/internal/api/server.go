package api

import (
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/api/handlers"
	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/config"
)

type Server struct {
	config  *config.Config
	handler http.Handler
	initErr error
}

func NewServer(cfg *config.Config) *Server {
	handler := http.Handler(http.NewServeMux())
	if cfg != nil {
		openAI := handlers.NewOpenAI()
		mux := http.NewServeMux()
		openAI.RegisterRoutes(mux)
		handler = APIKeyMiddleware(cfg.APIKeys, mux)
	}

	s := &Server{
		config:  cfg,
		handler: handler,
	}

	if cfg == nil {
		s.initErr = errors.New("config is required")
	}

	return s
}

func (s *Server) Run() error {
	if s.initErr != nil {
		return s.initErr
	}
	return s.buildHTTPServer().ListenAndServe()
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

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.handler == nil {
		http.NotFound(w, r)
		return
	}
	s.handler.ServeHTTP(w, r)
}
