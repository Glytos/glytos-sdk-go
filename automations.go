package glytos

import (
	"context"
	"net/url"
	"strconv"
)

// AutomationsService manages "when this happens, do that" rules: a webhook event
// fires an integration action, with no server of your own.
//
// Automations run in the background job that already handles the event, never
// during a call, and a failure is recorded rather than allowed to affect the
// conversation.
type AutomationsService struct{ client *Client }

// AutomationCreateParams are the fields for AutomationsService.Create.
type AutomationCreateParams struct {
	// Name labels the rule (required).
	Name string
	// TriggerEvent is a webhook event type, for example "session.completed"
	// (required). WebhooksService.Events lists them.
	TriggerEvent string
	// ConnectionUUID is the destination to act on (required).
	ConnectionUUID string
	// Action is one of the integration's actions (required).
	Action string
	// PayloadTemplate holds the action's parameters. Values may reference the
	// event with {{placeholders}}.
	PayloadTemplate map[string]any
	// Conditions narrow which events fire the rule.
	Conditions map[string]any
}

// List returns your automations.
func (s *AutomationsService) List(ctx context.Context) ([]Automation, error) {
	var out []Automation
	err := s.client.do(ctx, "GET", "/automations", nil, nil, &out)
	return out, err
}

// Create adds a rule.
func (s *AutomationsService) Create(ctx context.Context, params AutomationCreateParams) (*Automation, error) {
	body := map[string]any{
		"name":            params.Name,
		"trigger_event":   params.TriggerEvent,
		"connection_uuid": params.ConnectionUUID,
		"action":          params.Action,
	}
	if params.PayloadTemplate != nil {
		body["payload_template"] = params.PayloadTemplate
	}
	if params.Conditions != nil {
		body["conditions"] = params.Conditions
	}
	var out Automation
	err := s.client.do(ctx, "POST", "/automations", body, nil, &out)
	return &out, err
}

// Update changes an automation, including pausing it with is_active false. Only
// the fields present in body are changed.
func (s *AutomationsService) Update(ctx context.Context, automationUUID string, body map[string]any) (*Automation, error) {
	var out Automation
	err := s.client.do(ctx, "PATCH", "/automations/"+esc(automationUUID), body, nil, &out)
	return &out, err
}

// Delete removes an automation.
func (s *AutomationsService) Delete(ctx context.Context, automationUUID string) error {
	return s.client.do(ctx, "DELETE", "/automations/"+esc(automationUUID), nil, nil, nil)
}

// Runs returns recent firings, newest first: what ran, and what came back. Pass
// a limit of zero for the server default.
func (s *AutomationsService) Runs(ctx context.Context, automationUUID string, limit int) ([]AutomationRun, error) {
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	var out []AutomationRun
	err := s.client.do(ctx, "GET", "/automations/"+esc(automationUUID)+"/runs", nil, query, &out)
	return out, err
}

// Test fires the automation once against a payload you supply, so the rendered
// parameters and the destination's reply can be checked before a real event is
// trusted to it.
func (s *AutomationsService) Test(ctx context.Context, automationUUID string, payload map[string]any) (*AutomationTest, error) {
	body := map[string]any{"payload": orEmpty(payload)}
	var out AutomationTest
	err := s.client.do(ctx, "POST", "/automations/"+esc(automationUUID)+"/test", body, nil, &out)
	return &out, err
}
