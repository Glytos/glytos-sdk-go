package glytos

import "context"

// CampaignsService manages outbound calling campaigns over a phone number.
type CampaignsService struct{ client *Client }

// CampaignCreateParams are the fields for CampaignsService.Create.
type CampaignCreateParams struct {
	// Name is the campaign name (required).
	Name string
	// WorkflowUUID is the agent to run for each contact (required).
	WorkflowUUID string
	// FromNumber is the caller id to dial from (required). It must be a number
	// your organization has connected, or the API refuses the campaign.
	FromNumber string
	// Contacts is an optional initial contact list. Any spelling works: numbers
	// are converted to international form and duplicates are dropped.
	Contacts []string
	// ContactsCSV is the contents of a CSV file. The phone column is found by
	// its header or by which column holds phone numbers, and every other column
	// travels with that contact's call as a variable.
	ContactsCSV string
	// ScheduledAt starts the campaign at a moment in the future (RFC 3339). Left
	// empty the campaign is a draft until you Start it.
	ScheduledAt string
	// CallWindowStart and CallWindowEnd bound dialing to a range of hours
	// ("09:00", "20:00"), read in Timezone. Set both or neither.
	CallWindowStart string
	CallWindowEnd   string
	// Timezone is an IANA name, e.g. "Europe/Istanbul". Defaults to UTC.
	Timezone string
	// SuppressionPolicy is how much of the do-not-call list this campaign
	// honours: "strict" (default, all of it), "transactional" (skip entries that
	// only refused marketing), or "ignore" (skip entries the organization added
	// for itself; requests people made on a call still apply).
	SuppressionPolicy string
	// OverrideCallerRequests also calls people who asked, on a call, not to be
	// contacted again. Only valid with SuppressionPolicy "ignore".
	OverrideCallerRequests bool
}

// SuppressionPreview is how many of a contact list each policy would reach.
type SuppressionPreview struct {
	Contacts        int `json:"contacts"`
	SuppressedTotal int `json:"suppressed_total"`
	// CallerRequested is how many of them asked, on a call, not to be contacted.
	CallerRequested        int `json:"caller_requested"`
	ReachedIfStrict        int `json:"reached_if_strict"`
	ReachedIfTransactional int `json:"reached_if_transactional"`
	ReachedIfIgnore        int `json:"reached_if_ignore"`
	ReachedIfOverride      int `json:"reached_if_override"`
}

// ContactSyncResult is what adding contacts to a campaign did.
type ContactSyncResult struct {
	Added int `json:"added"`
	// Skipped were already on the list; Duplicates appeared more than once in
	// the file; Rejected held no usable phone number.
	Skipped    int `json:"skipped"`
	Duplicates int `json:"duplicates"`
	Rejected   int `json:"rejected"`
	// PhoneColumn is the column read as the phone number, so a file read from
	// the wrong one is distinguishable from one that could not be read.
	PhoneColumn string `json:"phone_column"`
	// OnDoNotCall is how many of the added contacts are suppressed and so will
	// never be dialed. Counted at import purely to report it.
	OnDoNotCall int `json:"on_do_not_call"`
}

// CampaignUpdateParams are the fields for CampaignsService.Update. Anything left
// empty is left alone. To remove a schedule entirely use Unschedule: omitting a
// field and clearing it are different instructions, and only one of them can be
// expressed by an empty string.
type CampaignUpdateParams struct {
	Name string
	// ScheduledAt is RFC 3339. Changing it is only allowed before the campaign
	// starts.
	ScheduledAt     string
	CallWindowStart string
	CallWindowEnd   string
	Timezone        string
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
	if params.ContactsCSV != "" {
		body["contacts_csv"] = params.ContactsCSV
	}
	if params.ScheduledAt != "" {
		body["scheduled_at"] = params.ScheduledAt
	}
	if params.CallWindowStart != "" {
		body["call_window_start"] = params.CallWindowStart
	}
	if params.CallWindowEnd != "" {
		body["call_window_end"] = params.CallWindowEnd
	}
	if params.Timezone != "" {
		body["timezone"] = params.Timezone
	}
	if params.SuppressionPolicy != "" {
		body["suppression_policy"] = params.SuppressionPolicy
	}
	if params.OverrideCallerRequests {
		body["override_caller_requests"] = true
	}
	var out Campaign
	err := s.client.do(ctx, "POST", "/telephony/campaigns", body, nil, &out)
	return &out, err
}

// Retrieve returns a campaign by uuid, with its contacts and their outcomes.
func (s *CampaignsService) Retrieve(ctx context.Context, campaignUUID string) (*CampaignDetail, error) {
	var out CampaignDetail
	err := s.client.do(ctx, "GET", "/telephony/campaigns/"+esc(campaignUUID), nil, nil, &out)
	return &out, err
}

// Start begins dialing, from the contacts that have not been called yet. A
// campaign that is already running is refused.
func (s *CampaignsService) Start(ctx context.Context, campaignUUID string) (*Campaign, error) {
	var out Campaign
	err := s.client.do(ctx, "POST", "/telephony/campaigns/"+esc(campaignUUID)+"/start", nil, nil, &out)
	return &out, err
}

// Stop ends dialing at the next contact. Calls already handed to the carrier
// run to their end; undialed contacts stay ready, so Start resumes.
func (s *CampaignsService) Stop(ctx context.Context, campaignUUID string) (*Campaign, error) {
	var out Campaign
	err := s.client.do(ctx, "POST", "/telephony/campaigns/"+esc(campaignUUID)+"/stop", nil, nil, &out)
	return &out, err
}

// Update renames a campaign, or changes when and within what hours it dials.
//
// A rename is accepted at any point. The schedule and the calling window can
// only be changed before the campaign starts: moving the start of one already
// dialing would say nothing about the calls it has placed.
func (s *CampaignsService) Update(ctx context.Context, campaignUUID string, params CampaignUpdateParams) (*Campaign, error) {
	body := map[string]any{}
	if params.Name != "" {
		body["name"] = params.Name
	}
	if params.ScheduledAt != "" {
		body["scheduled_at"] = params.ScheduledAt
	}
	if params.CallWindowStart != "" {
		body["call_window_start"] = params.CallWindowStart
	}
	if params.CallWindowEnd != "" {
		body["call_window_end"] = params.CallWindowEnd
	}
	if params.Timezone != "" {
		body["timezone"] = params.Timezone
	}
	var out Campaign
	err := s.client.do(ctx, "PATCH", "/telephony/campaigns/"+esc(campaignUUID), body, nil, &out)
	return &out, err
}

// Unschedule clears a campaign's schedule, returning it to a draft that waits
// for Start.
func (s *CampaignsService) Unschedule(ctx context.Context, campaignUUID string) (*Campaign, error) {
	var out Campaign
	body := map[string]any{"scheduled_at": nil}
	err := s.client.do(ctx, "PATCH", "/telephony/campaigns/"+esc(campaignUUID), body, nil, &out)
	return &out, err
}

// Duplicate copies a campaign and its contact list into a fresh draft. Nothing
// dials and no outcome is copied, so this is how you run the same list again or
// reuse a setup against a new one. An empty name keeps the original's, suffixed.
func (s *CampaignsService) Duplicate(ctx context.Context, campaignUUID, name string) (*Campaign, error) {
	body := map[string]any{}
	if name != "" {
		body["name"] = name
	}
	var out Campaign
	err := s.client.do(ctx, "POST", "/telephony/campaigns/"+esc(campaignUUID)+"/duplicate", body, nil, &out)
	return &out, err
}

// Export returns the contacts and what came of each, as CSV: phone, outcome,
// dialed_at, error, session_uuid. The session uuid joins a result back to the
// conversation that produced it.
func (s *CampaignsService) Export(ctx context.Context, campaignUUID string) ([]byte, error) {
	return s.client.doRaw(ctx, "GET", "/telephony/campaigns/"+esc(campaignUUID)+"/export")
}

// Delete removes a campaign and its contact list. A running campaign is
// stopped first.
func (s *CampaignsService) Delete(ctx context.Context, campaignUUID string) error {
	return s.client.do(ctx, "DELETE", "/telephony/campaigns/"+esc(campaignUUID), nil, nil, nil)
}

// AddContacts appends contacts from the contents of a CSV file.
func (s *CampaignsService) AddContacts(ctx context.Context, campaignUUID, contactsCSV string) (*ContactSyncResult, error) {
	var out ContactSyncResult
	body := map[string]any{"contacts_csv": contactsCSV}
	err := s.client.do(ctx, "POST", "/telephony/campaigns/"+esc(campaignUUID)+"/contacts/sync", body, nil, &out)
	return &out, err
}

// SyncContacts appends contacts from a CSV your own system serves over HTTP.
func (s *CampaignsService) SyncContacts(ctx context.Context, campaignUUID, sourceURL string) (*ContactSyncResult, error) {
	var out ContactSyncResult
	body := map[string]any{"source_url": sourceURL}
	err := s.client.do(ctx, "POST", "/telephony/campaigns/"+esc(campaignUUID)+"/contacts/sync", body, nil, &out)
	return &out, err
}

// PreviewSuppression reports how many of a contact list each suppression policy
// would reach, including how many of those people asked on a call not to be
// contacted. Measure before choosing anything other than the default.
func (s *CampaignsService) PreviewSuppression(ctx context.Context, contacts []string, contactsCSV string) (*SuppressionPreview, error) {
	body := map[string]any{}
	if contacts != nil {
		body["contacts"] = contacts
	}
	if contactsCSV != "" {
		body["contacts_csv"] = contactsCSV
	}
	var out SuppressionPreview
	err := s.client.do(ctx, "POST", "/telephony/campaigns/suppression-preview", body, nil, &out)
	return &out, err
}
