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

// List returns calls. Pass nil for no filters. The endpoint is paginated and
// wraps results in an {items, ...} envelope; the items are returned here.
func (s *CallsService) List(ctx context.Context, query url.Values) ([]Call, error) {
	var page struct {
		Items []Call `json:"items"`
	}
	err := s.client.do(ctx, "GET", "/calls", nil, query, &page)
	return page.Items, err
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

// Control acts on a call that is happening right now.
//
// body takes an "action" of "say", "transfer" or "end". A say needs "text", a
// transfer needs "to_number", and an end needs neither. Prefer the Say, Transfer
// and End helpers, which spell that out.
func (s *CallsService) Control(ctx context.Context, callUUID string, body map[string]any) (json.RawMessage, error) {
	var out json.RawMessage
	err := s.client.do(ctx, "POST", "/calls/"+esc(callUUID)+"/control", body, nil, &out)
	return out, err
}

// Say makes the agent speak a line on a call in progress.
func (s *CallsService) Say(ctx context.Context, callUUID, text string) (json.RawMessage, error) {
	return s.Control(ctx, callUUID, map[string]any{"action": "say", "text": text})
}

// Transfer hands a call in progress to a person.
func (s *CallsService) Transfer(ctx context.Context, callUUID, toNumber string) (json.RawMessage, error) {
	return s.Control(ctx, callUUID, map[string]any{"action": "transfer", "to_number": toNumber})
}

// End hangs up a call in progress.
func (s *CallsService) End(ctx context.Context, callUUID string) (json.RawMessage, error) {
	return s.Control(ctx, callUUID, map[string]any{"action": "end"})
}
