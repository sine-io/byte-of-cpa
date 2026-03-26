package executor

import (
	"errors"
	"net/http"

	"github.com/router-for-me/CLIProxyAPI/v6/nanocpa/internal/auth"
)

type Result = auth.Result

type ChatCompletionsExecutor = auth.Executor

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
