package glytos

import (
	"context"
	"encoding/json"
	"net/url"
)

// CallsService manages phone and web calls.
type CallsService struct{ client *Client }

// WebTokenParams are the fields for CallsService.WebToken. Provide either an
// existing WorkflowUUID or a transient Agent definition.
type WebTokenParams struct {
	// WorkflowUUID scopes the token to a saved agent.
	WorkflowUUID string
	// Agent is a transient inline agent definition.
	Agent map[string]any
}

// Create starts an outbound phone call, or runs a transient agent. The body is
// passed through as-is.
func (s *CallsService) Create(ctx context.Context, body map[string]any) (*Call, error) {
	var out Call
	err := s.client.do(ctx, "POST", "/calls", body, nil, &out)
	return &out, err
}

// List returns calls. Pass nil for no filters.
func (s *CallsService) List(ctx context.Context, query url.Values) ([]Call, error) {
	var out []Call
	err := s.client.do(ctx, "GET", "/calls", nil, query, &out)
	return out, err
}

// Retrieve returns a call by uuid.
func (s *CallsService) Retrieve(ctx context.Context, callUUID string) (*Call, error) {
	var out Call
	err := s.client.do(ctx, "GET", "/calls/"+esc(callUUID), nil, nil, &out)
	return &out, err
}

// WebToken mints a short-lived, workflow-scoped token for an in-browser web
// call. Hand the returned token and ws_url to the browser and connect with
// "@glytos/web".
func (s *CallsService) WebToken(ctx context.Context, params WebTokenParams) (*WebCallToken, error) {
	body := map[string]any{}
	if params.WorkflowUUID != "" {
		body["workflow_uuid"] = params.WorkflowUUID
	}
	if params.Agent != nil {
		body["agent"] = params.Agent
	}
	var out WebCallToken
	err := s.client.do(ctx, "POST", "/calls/web-token", body, nil, &out)
	return &out, err
}

// Control controls an in-progress call (for example transfer or hang up).
func (s *CallsService) Control(ctx context.Context, callUUID string, body map[string]any) (json.RawMessage, error) {
	var out json.RawMessage
	err := s.client.do(ctx, "POST", "/calls/"+esc(callUUID)+"/control", body, nil, &out)
	return out, err
}
