package glytos

import (
	"context"
	"net/url"
)

// SessionsService lists sessions across your agents.
type SessionsService struct{ client *Client }

// List returns sessions across your agents. Pass nil for no filters.
func (s *SessionsService) List(ctx context.Context, query url.Values) ([]Session, error) {
	var out []Session
	err := s.client.do(ctx, "GET", "/sessions", nil, query, &out)
	return out, err
}
