package auth

import "sync"

type Selector interface {
	Select(model string, candidates []*Auth) *Auth
}

type RoundRobinSelector struct {
	mu     sync.Mutex
	nextBy map[string]uint64
}

func NewRoundRobinSelector() *RoundRobinSelector {
	return &RoundRobinSelector{
		nextBy: map[string]uint64{},
	}
}

func (s *RoundRobinSelector) Select(model string, candidates []*Auth) *Auth {
	if len(candidates) == 0 {
		return nil
	}

	s.mu.Lock()
	next := s.nextBy[model]
	idx := int(next % uint64(len(candidates)))
	s.nextBy[model] = next + 1
	s.mu.Unlock()

	return candidates[idx]
}
