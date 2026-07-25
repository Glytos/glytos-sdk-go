package glytos

import (
	"context"
	"encoding/json"
)

// CampaignsService manages outbound calling campaigns over a phone number.
type CampaignsService struct{ client *Client }

// CampaignCreateParams are the fields for CampaignsService.Create.
type CampaignCreateParams struct {
	// Name is the campaign name (required).
	Name string
	// WorkflowUUID is the agent to run for each contact (required).
	WorkflowUUID string
	// FromNumber is the caller-id number to dial from (required).
	FromNumber string
	// Contacts is the optional initial contact list.
	Contacts []map[string]any
}

// List returns your outbound calling campaigns.
func (s *CampaignsService) List(ctx context.Context) ([]Campaign, error) {
	var out []Campaign
	err := s.client.do(ctx, "GET", "/telephony/campaigns", nil, nil, &out)
	return out, err
}

// Create creates an outbound calling campaign.
func (s *CampaignsService) Create(ctx context.Context, params CampaignCreateParams) (*Campaign, error) {
	body := map[string]any{
		"name":          params.Name,
		"workflow_uuid": params.WorkflowUUID,
		"from_number":   params.FromNumber,
	}
	if params.Contacts != nil {
		body["contacts"] = params.Contacts
	}
	var out Campaign
	err := s.client.do(ctx, "POST", "/telephony/campaigns", body, nil, &out)
	return &out, err
}

// Retrieve returns a campaign by uuid.
func (s *CampaignsService) Retrieve(ctx context.Context, campaignUUID string) (*Campaign, error) {
	var out Campaign
	err := s.client.do(ctx, "GET", "/telephony/campaigns/"+esc(campaignUUID), nil, nil, &out)
	return &out, err
}

// Start starts a campaign (begins dialing its contacts).
func (s *CampaignsService) Start(ctx context.Context, campaignUUID string) (*Campaign, error) {
	var out Campaign
	err := s.client.do(ctx, "POST", "/telephony/campaigns/"+esc(campaignUUID)+"/start", nil, nil, &out)
	return &out, err
}

// SyncContacts loads a campaign's contacts from a remote source URL.
func (s *CampaignsService) SyncContacts(ctx context.Context, campaignUUID, sourceURL string) (json.RawMessage, error) {
	body := map[string]any{"source_url": sourceURL}
	var out json.RawMessage
	err := s.client.do(ctx, "POST", "/telephony/campaigns/"+esc(campaignUUID)+"/contacts/sync", body, nil, &out)
	return out, err
}
