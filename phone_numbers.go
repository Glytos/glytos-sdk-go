package glytos

import (
	"context"
	"net/url"
)

// PhoneNumbersService manages telephony numbers and providers.
type PhoneNumbersService struct{ client *Client }

// ImportNumberParams are the fields for PhoneNumbersService.ImportNumber. Only
// E164 is required; the rest are sent only when set.
type ImportNumberParams struct {
	// E164 is the number in E.164 format (required).
	E164 string
	// Provider is the carrier (for example "twilio").
	Provider string
	// ProviderSID is the carrier-side identifier for the number.
	ProviderSID string
	// Credentials are the carrier credentials to import with.
	Credentials map[string]any
	// WorkflowUUID assigns the imported number to an agent.
	WorkflowUUID string
}

// Search searches carrier inventory for available numbers.
func (s *PhoneNumbersService) Search(ctx context.Context, query url.Values) ([]map[string]any, error) {
	var out []map[string]any
	err := s.client.do(ctx, "GET", "/telephony/numbers/search", nil, query, &out)
	return out, err
}

// List returns the numbers on your account.
func (s *PhoneNumbersService) List(ctx context.Context) ([]PhoneNumber, error) {
	var out []PhoneNumber
	err := s.client.do(ctx, "GET", "/telephony/numbers", nil, nil, &out)
	return out, err
}

// Providers lists the telephony providers available to your account.
func (s *PhoneNumbersService) Providers(ctx context.Context) ([]map[string]any, error) {
	var out []map[string]any
	err := s.client.do(ctx, "GET", "/telephony/providers", nil, nil, &out)
	return out, err
}

// Provision provisions (buys) a number by its E.164 value. Extra fields are
// merged into the request body.
func (s *PhoneNumbersService) Provision(ctx context.Context, e164 string, extra map[string]any) (*PhoneNumber, error) {
	body := map[string]any{"e164": e164}
	for k, v := range extra {
		body[k] = v
	}
	var out PhoneNumber
	err := s.client.do(ctx, "POST", "/telephony/numbers", body, nil, &out)
	return &out, err
}

// ImportNumber imports (connects) a number you already own at a carrier.
func (s *PhoneNumbersService) ImportNumber(ctx context.Context, params ImportNumberParams) (*PhoneNumber, error) {
	body := map[string]any{"e164": params.E164}
	if params.Provider != "" {
		body["provider"] = params.Provider
	}
	if params.ProviderSID != "" {
		body["provider_sid"] = params.ProviderSID
	}
	if params.Credentials != nil {
		body["credentials"] = params.Credentials
	}
	if params.WorkflowUUID != "" {
		body["workflow_uuid"] = params.WorkflowUUID
	}
	var out PhoneNumber
	err := s.client.do(ctx, "POST", "/telephony/numbers/import", body, nil, &out)
	return &out, err
}

// Instant provisions a platform "instant" number. Both arguments are optional
// and are sent as query parameters (there is no request body).
func (s *PhoneNumbersService) Instant(ctx context.Context, country, provider string) (*PhoneNumber, error) {
	query := url.Values{}
	if country != "" {
		query.Set("country", country)
	}
	if provider != "" {
		query.Set("provider", provider)
	}
	var out PhoneNumber
	err := s.client.do(ctx, "POST", "/telephony/numbers/instant", nil, query, &out)
	return &out, err
}

// Assign assigns a number to an agent.
func (s *PhoneNumbersService) Assign(ctx context.Context, numberUUID string, body map[string]any) (*PhoneNumber, error) {
	var out PhoneNumber
	err := s.client.do(ctx, "POST", "/telephony/numbers/"+esc(numberUUID)+"/assign", body, nil, &out)
	return &out, err
}

// Release releases (deletes) a number.
func (s *PhoneNumbersService) Release(ctx context.Context, numberUUID string) error {
	return s.client.do(ctx, "DELETE", "/telephony/numbers/"+esc(numberUUID), nil, nil, nil)
}
