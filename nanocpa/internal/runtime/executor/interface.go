package executor

import (
	"context"
	"errors"
	"net/http"

	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/auth"
)

type Result struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

type ChatCompletionsExecutor interface {
	Execute(ctx context.Context, openAIRequest []byte, runtimeAuth *auth.Auth) (*Result, error)
}

type UpstreamError struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

func (e *UpstreamError) Error() string {
	return "upstream provider returned non-2xx response"
}

func AsUpstreamError(err error, target **UpstreamError) bool {
	return errors.As(err, target)
}
