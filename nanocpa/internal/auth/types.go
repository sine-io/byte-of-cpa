package auth

import "time"

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
