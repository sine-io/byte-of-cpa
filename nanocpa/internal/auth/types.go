package auth

import (
	"context"
	"net/http"
	"time"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusCooldown Status = "cooldown"
	StatusDisabled Status = "disabled"
)

type Auth struct {
	ID         string
	Provider   string
	Label      string
	Status     Status
	Disabled   bool
	Attributes map[string]string
	Metadata   map[string]any
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Result struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

type Executor interface {
	Execute(ctx context.Context, openAIRequest []byte, runtimeAuth *Auth) (*Result, error)
}
