package glytos

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
)

// WebhooksService manages webhook endpoints, deliveries, and signature
// verification.
type WebhooksService struct{ client *Client }

// WebhookCreateParams are the fields for WebhooksService.Create. URL and Events
// are required; the rest are sent only when set.
type WebhookCreateParams struct {
	// URL is the endpoint to deliver events to (required).
	URL string
	// Events are the event types to subscribe to (required).
	Events []string
	// IsActive enables or disables the endpoint.
	IsActive *bool
	// TimeoutSeconds is the per-delivery timeout.
	TimeoutSeconds *int
	// Headers are extra headers to send with each delivery.
	Headers map[string]string
	// AuthHeader is an Authorization header value to send with each delivery.
	AuthHeader string
}

// WebhookUpdateParams are the optional fields for WebhooksService.Update. Only
// the fields you set are changed.
type WebhookUpdateParams struct {
	URL            string
	Events         []string
	IsActive       *bool
	TimeoutSeconds *int
	Headers        map[string]string
	AuthHeader     string
}

// WebhookDeliveriesParams are the optional filters for WebhooksService.Deliveries.
type WebhookDeliveriesParams struct {
	EventType string
	Status    string
	Limit     *int
	Offset    *int
}

// List returns your webhook endpoints.
func (s *WebhooksService) List(ctx context.Context) ([]WebhookEndpoint, error) {
	var out []WebhookEndpoint
	err := s.client.do(ctx, "GET", "/webhooks/endpoints", nil, nil, &out)
	return out, err
}

// Create creates a webhook endpoint subscribed to the given events.
func (s *WebhooksService) Create(ctx context.Context, params WebhookCreateParams) (*WebhookEndpoint, error) {
	body := map[string]any{"url": params.URL, "events": params.Events}
	if params.IsActive != nil {
		body["is_active"] = *params.IsActive
	}
	if params.TimeoutSeconds != nil {
		body["timeout_seconds"] = *params.TimeoutSeconds
	}
	if params.Headers != nil {
		body["headers"] = params.Headers
	}
	if params.AuthHeader != "" {
		body["auth_header"] = params.AuthHeader
	}
	var out WebhookEndpoint
	err := s.client.do(ctx, "POST", "/webhooks/endpoints", body, nil, &out)
	return &out, err
}

// Update updates a webhook endpoint. Only the fields you set are changed.
func (s *WebhooksService) Update(ctx context.Context, endpointID int, params WebhookUpdateParams) (*WebhookEndpoint, error) {
	body := map[string]any{}
	if params.URL != "" {
		body["url"] = params.URL
	}
	if params.Events != nil {
		body["events"] = params.Events
	}
	if params.IsActive != nil {
		body["is_active"] = *params.IsActive
	}
	if params.TimeoutSeconds != nil {
		body["timeout_seconds"] = *params.TimeoutSeconds
	}
	if params.Headers != nil {
		body["headers"] = params.Headers
	}
	if params.AuthHeader != "" {
		body["auth_header"] = params.AuthHeader
	}
	var out WebhookEndpoint
	err := s.client.do(ctx, "PATCH", "/webhooks/endpoints/"+esc(strconv.Itoa(endpointID)), body, nil, &out)
	return &out, err
}

// Delete deletes a webhook endpoint.
func (s *WebhooksService) Delete(ctx context.Context, endpointID int) error {
	return s.client.do(ctx, "DELETE", "/webhooks/endpoints/"+esc(strconv.Itoa(endpointID)), nil, nil, nil)
}

// Events returns the catalog of webhook event types you can subscribe to.
func (s *WebhooksService) Events(ctx context.Context) ([]map[string]any, error) {
	var out []map[string]any
	err := s.client.do(ctx, "GET", "/webhooks/events", nil, nil, &out)
	return out, err
}

// Deliveries lists recent webhook deliveries. Pass nil for no filters.
func (s *WebhooksService) Deliveries(ctx context.Context, params *WebhookDeliveriesParams) ([]WebhookDelivery, error) {
	query := url.Values{}
	if params != nil {
		if params.EventType != "" {
			query.Set("event_type", params.EventType)
		}
		if params.Status != "" {
			query.Set("status", params.Status)
		}
		if params.Limit != nil {
			query.Set("limit", strconv.Itoa(*params.Limit))
		}
		if params.Offset != nil {
			query.Set("offset", strconv.Itoa(*params.Offset))
		}
	}
	var out []WebhookDelivery
	err := s.client.do(ctx, "GET", "/webhooks/deliveries", nil, query, &out)
	return out, err
}

// Redeliver re-sends a past webhook delivery.
func (s *WebhooksService) Redeliver(ctx context.Context, deliveryID int) (json.RawMessage, error) {
	var out json.RawMessage
	err := s.client.do(ctx, "POST", "/webhooks/deliveries/"+esc(strconv.Itoa(deliveryID))+"/redeliver", nil, nil, &out)
	return out, err
}

// Verify verifies a webhook delivery signature. It is a convenience wrapper for
// the package-level VerifyWebhook. Pass DefaultWebhookTolerance for the default
// replay window.
func (s *WebhooksService) Verify(payload []byte, signatureHeader, secret string, toleranceSeconds int) bool {
	return VerifyWebhook(payload, signatureHeader, secret, toleranceSeconds)
}
