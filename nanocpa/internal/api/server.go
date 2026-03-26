package api

import (
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/config"
)

type Server struct {
	config  *config.Config
	handler http.Handler
	initErr error
}

func NewServer(cfg *config.Config) *Server {
	s := &Server{
		config:  cfg,
		handler: http.NewServeMux(),
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
