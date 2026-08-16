package glytos

import (
	"context"
	"net/url"
)

// IntegrationsService exposes third-party destinations - Slack, Discord,
// Telegram, a generic webhook, Cal.com - and the connections that hold their
// credentials.
//
// A connection is reachable three ways: run it directly through Connections.Run,
// give an agent a tool of kind "integration" that names it so the model can act
// mid-conversation, or fire it from an automation when an event happens.
type IntegrationsService struct {
	client *Client
	// Connections manages the configured destinations.
	Connections *IntegrationConnectionsService
}

// IntegrationConnectionsService manages the configured destinations behind an
// integration. An organization can hold several per integration - two Slack
// workspaces, three calendars - which is why a rule names the connection rather
// than the integration.
type IntegrationConnectionsService struct{ client *Client }

// List returns the catalog: what can be connected, and the actions each offers.
func (s *IntegrationsService) List(ctx context.Context) ([]Integration, error) {
	var out []Integration
	err := s.client.do(ctx, "GET", "/integrations", nil, nil, &out)
	return out, err
}

// Run performs an action using whatever credentials the organization saved for
// this integration key.
//
// Prefer Connections.Run: this resolves the organization's stored credentials for
// the integration, which is ambiguous once there is more than one destination.
func (s *IntegrationsService) Run(ctx context.Context, integrationKey, action string, params map[string]any) (*IntegrationResult, error) {
	body := map[string]any{"action": action, "params": orEmpty(params)}
	var out IntegrationResult
	err := s.client.do(ctx, "POST", "/integrations/"+esc(integrationKey)+"/run", body, nil, &out)
	return &out, err
}

// List returns the configured connections, optionally for one integration.
func (s *IntegrationConnectionsService) List(ctx context.Context, integrationKey string) ([]IntegrationConnection, error) {
	query := url.Values{}
	if integrationKey != "" {
		query.Set("integration_key", integrationKey)
	}
	var out []IntegrationConnection
	err := s.client.do(ctx, "GET", "/integrations/connections", nil, query, &out)
	return out, err
}

// Create configures a destination. data carries the integration's required
// credentials (see IntegrationsService.List); they are encrypted at rest and
// masked when read back.
func (s *IntegrationConnectionsService) Create(ctx context.Context, integrationKey, name string, data map[string]any) (*IntegrationConnection, error) {
	body := map[string]any{
		"integration_key": integrationKey,
		"name":            name,
		"data":            orEmpty(data),
	}
	var out IntegrationConnection
	err := s.client.do(ctx, "POST", "/integrations/connections", body, nil, &out)
	return &out, err
}

// Update changes a connection. Only the fields present in body are changed.
func (s *IntegrationConnectionsService) Update(ctx context.Context, connectionUUID string, body map[string]any) (*IntegrationConnection, error) {
	var out IntegrationConnection
	err := s.client.do(ctx, "PATCH", "/integrations/connections/"+esc(connectionUUID), body, nil, &out)
	return &out, err
}

// Delete removes a connection. Automations pointing at it stop firing.
func (s *IntegrationConnectionsService) Delete(ctx context.Context, connectionUUID string) error {
	return s.client.do(ctx, "DELETE", "/integrations/connections/"+esc(connectionUUID), nil, nil, nil)
}

// Run performs one of the integration's actions through this connection.
func (s *IntegrationConnectionsService) Run(ctx context.Context, connectionUUID, action string, params map[string]any) (*IntegrationResult, error) {
	body := map[string]any{"action": action, "params": orEmpty(params)}
	var out IntegrationResult
	err := s.client.do(ctx, "POST", "/integrations/connections/"+esc(connectionUUID)+"/run", body, nil, &out)
	return &out, err
}

// orEmpty keeps a nil map out of the JSON body as null, which the API reads as a
// missing object rather than "no parameters".
func orEmpty(m map[string]any) map[string]any {
	if m == nil {
		return map[string]any{}
	}
	return m
}
