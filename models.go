package glytos

// The entity types below carry the fields you commonly rely on. JSON decoding
// ignores any additional fields the API returns, so they stay forward-compatible
// as the API grows. When you need the full, untyped payload, call Client.Do with
// a *json.RawMessage (or a map[string]any) destination.

// Workflow is an agent: a prompt agent or a visual workflow.
type Workflow struct {
	UUID     string `json:"uuid"`
	Name     string `json:"name"`
	Mode     string `json:"mode"`
	Status   string `json:"status,omitempty"`
	Archived bool   `json:"archived,omitempty"`
}

// WorkflowVersion is a saved version of an agent.
type WorkflowVersion struct {
	Version int `json:"version,omitempty"`
}

// Call is a phone or web call.
type Call struct {
	UUID   string `json:"uuid"`
	Status string `json:"status"`
}

// WebCallToken is a short-lived, workflow-scoped token for an in-browser web
// call. Hand it and WSURL to the browser and connect with "@glytos/web".
type WebCallToken struct {
	Token string `json:"token"`
	WSURL string `json:"ws_url"`
}

// PhoneNumber is a telephony number on your account.
type PhoneNumber struct {
	UUID string `json:"uuid"`
	E164 string `json:"e164"`
}

// Session is a text or voice session against an agent.
type Session struct {
	SessionUUID  string `json:"session_uuid"`
	WorkflowUUID string `json:"workflow_uuid,omitempty"`
	Mode         string `json:"mode,omitempty"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at,omitempty"`
}

// WebhookEndpoint is a subscribed webhook endpoint.
type WebhookEndpoint struct {
	ID     int      `json:"id"`
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

// WebhookDelivery is a single webhook delivery attempt.
type WebhookDelivery struct {
	ID        int    `json:"id"`
	EventType string `json:"event_type,omitempty"`
	Status    string `json:"status,omitempty"`
}

// Campaign is an outbound calling campaign over a phone number.
type Campaign struct {
	UUID         string `json:"uuid"`
	Name         string `json:"name"`
	WorkflowUUID string `json:"workflow_uuid,omitempty"`
	FromNumber   string `json:"from_number,omitempty"`
	// Status is one of draft, scheduled, running, waiting, stopped, completed,
	// halted, out_of_credit, failed.
	Status string `json:"status,omitempty"`
	// StatusDetail says why a campaign stopped somewhere other than the end of
	// its list, so "halted" and "out_of_credit" are actionable.
	StatusDetail string `json:"status_detail,omitempty"`
	ScheduledAt  string `json:"scheduled_at,omitempty"`
	StartedAt    string `json:"started_at,omitempty"`
	FinishedAt   string `json:"finished_at,omitempty"`
	// CallWindowStart and CallWindowEnd bound dialing, read in Timezone.
	CallWindowStart        string `json:"call_window_start,omitempty"`
	CallWindowEnd          string `json:"call_window_end,omitempty"`
	Timezone               string `json:"timezone,omitempty"`
	SuppressionPolicy      string `json:"suppression_policy,omitempty"`
	OverrideCallerRequests bool   `json:"override_caller_requests,omitempty"`
	// WorkflowName is the agent that does the dialing, named so a row need not
	// resolve the uuid against the agent list.
	WorkflowName string `json:"workflow_name,omitempty"`
	// Counts is how far the campaign has got.
	Counts CampaignCounts `json:"counts"`
	// Imported is only set on the create response: what reading the supplied
	// contact list did.
	Imported *ContactSyncResult `json:"imported,omitempty"`
}

// CampaignCounts is how far a campaign has got, sent with every campaign so a
// row can draw its progress without fetching the contact list to count it.
type CampaignCounts struct {
	Total     int `json:"total"`
	Pending   int `json:"pending"`
	Dialing   int `json:"dialing"`
	Answered  int `json:"answered"`
	Voicemail int `json:"voicemail"`
	NoAnswer  int `json:"no_answer"`
	Failed    int `json:"failed"`
	// Suppressed are on the do-not-call list, so never dialed.
	Suppressed int `json:"suppressed"`
	// Dialed were handed to the carrier, including calls still in flight. It
	// excludes Suppressed.
	Dialed int `json:"dialed"`
	// Dialable is what the campaign can ever dial: Total minus Suppressed.
	// Measure progress against this, not Total, or a finished campaign stops
	// short of complete by however many numbers were suppressed.
	Dialable int `json:"dialable"`
}

// CampaignContact is one dial target and what became of it.
type CampaignContact struct {
	Phone string `json:"phone"`
	// Status is one of pending, dialing, answered, voicemail, no_answer,
	// failed, suppressed. Busy is not reported separately from no_answer: it
	// needs per-carrier callbacks the platform does not collect.
	Status string `json:"status"`
	// CallSID is the carrier's own id for the call.
	CallSID string `json:"call_sid,omitempty"`
	// Error is the carrier's own words when it refused the number.
	Error string `json:"error,omitempty"`
	// SessionUUID is the conversation this contact produced, if it answered.
	SessionUUID string `json:"session_uuid,omitempty"`
	// Variables are the contact's other CSV columns, which reach the agent's
	// prompt, so {{name}} means this person.
	Variables map[string]string `json:"variables,omitempty"`
}

// CampaignDetail is a campaign with its contact list.
type CampaignDetail struct {
	Campaign
	Contacts []CampaignContact `json:"contacts"`
}

// DncEntry is a number this organization must not call.
type DncEntry struct {
	UUID  string `json:"uuid"`
	Phone string `json:"phone"`
	// Source is how it got here: agent (the person asked on a call), manual,
	// import or api.
	Source string `json:"source"`
	// Scope is how far it reaches: "all" or "marketing".
	Scope         string `json:"scope"`
	Reason        string `json:"reason,omitempty"`
	LastMatchedAt string `json:"last_matched_at,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
}

// Tool is a reusable tool an agent can call (kind = http / static / mcp).
type Tool struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
	// Kind is one of static, http, mcp, code, integration, client.
	Kind string `json:"kind"`
}

// McpTool is one tool an MCP server publishes, as discovered from the server.
type McpTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

// Document is a knowledge-base document.
type Document struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	// Content is returned when a single document is retrieved, not when listing.
	Content string `json:"content,omitempty"`
	Status  string `json:"status,omitempty"`
}

// SipTrunk is a BYO SIP trunk registered with a carrier.
type SipTrunk struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	Preset    string `json:"preset"`
	SipServer string `json:"sip_server"`
	SipPort   int    `json:"sip_port"`
	Transport string `json:"transport"`
	Username  string `json:"username"`
	// Status is registered, pending or failed. Only a registered trunk takes calls.
	Status           string `json:"status,omitempty"`
	StatusDetail     string `json:"status_detail,omitempty"`
	LastRegisteredAt string `json:"last_registered_at,omitempty"`
	NumberCount      int    `json:"number_count,omitempty"`
}

// SipPreset is a carrier whose connection settings are already known, so only
// the login has to be supplied.
type SipPreset struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	SipServer string `json:"sip_server"`
	SipPort   int    `json:"sip_port"`
	Transport string `json:"transport"`
	Country   string `json:"country"`
	// Verified says whether these settings have been confirmed against the live
	// carrier, rather than taken from its documentation.
	Verified bool   `json:"verified"`
	Note     string `json:"note"`
}

// SipTrunkTest is the result of re-checking a trunk against its carrier.
type SipTrunkTest struct {
	OK     bool   `json:"ok"`
	Detail string `json:"detail"`
	// Reachable separates "the carrier refused these credentials" from "nobody
	// answered". Only the first is worth changing the password over.
	Reachable bool `json:"reachable,omitempty"`
}

// TestSuite is a set of saved conversations replayed against one agent.
type TestSuite struct {
	UUID         string           `json:"uuid"`
	WorkflowUUID string           `json:"workflow_uuid"`
	Name         string           `json:"name"`
	Cases        []map[string]any `json:"cases"`
}

// TestSuiteRun is the outcome of running every case in a suite.
type TestSuiteRun struct {
	SuiteUUID   string           `json:"suite_uuid"`
	Passed      bool             `json:"passed"`
	Total       int              `json:"total"`
	PassedCount int              `json:"passed_count"`
	Results     []map[string]any `json:"results"`
}

// Integration is a third-party destination the platform can act on.
type Integration struct {
	Key                 string           `json:"key"`
	Name                string           `json:"name"`
	RequiredCredentials []string         `json:"required_credentials"`
	PublicCredentials   []string         `json:"public_credentials,omitempty"`
	SupportsAutomation  bool             `json:"supports_automation,omitempty"`
	Actions             []map[string]any `json:"actions"`
}

// IntegrationConnection is one configured destination. An organization can hold
// several per integration, so an agent or automation names the connection.
type IntegrationConnection struct {
	UUID           string `json:"uuid"`
	IntegrationKey string `json:"integration_key"`
	Name           string `json:"name"`
	IsActive       bool   `json:"is_active"`
	// Data comes back masked; secrets are never returned.
	Data            map[string]any `json:"data"`
	AutomationCount int            `json:"automation_count,omitempty"`
}

// IntegrationResult is what an integration action returned.
type IntegrationResult struct {
	Result map[string]any `json:"result"`
}

// Automation fires an integration action when an event happens.
type Automation struct {
	UUID     string `json:"uuid"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
	// TriggerEvent is a webhook event type, for example session.completed.
	TriggerEvent    string         `json:"trigger_event"`
	ConnectionUUID  string         `json:"connection_uuid"`
	IntegrationKey  string         `json:"integration_key"`
	Action          string         `json:"action"`
	PayloadTemplate map[string]any `json:"payload_template"`
	Conditions      map[string]any `json:"conditions"`
}

// AutomationRun is one firing of an automation.
type AutomationRun struct {
	EventType  string `json:"event_type"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
	DurationMs int    `json:"duration_ms"`
	CreatedAt  string `json:"created_at"`
}

// AutomationTest is a trial firing: the rendered parameters and the reply.
type AutomationTest struct {
	Params map[string]any `json:"params"`
	Result map[string]any `json:"result"`
}

// CreditBalance is the organization's prepaid balance.
type CreditBalance struct {
	Balance  float64 `json:"balance"`
	Currency string  `json:"currency"`
}

// CreditTransaction is one entry in the credit ledger.
type CreditTransaction struct {
	Amount       float64 `json:"amount"`
	Kind         string  `json:"kind"`
	Description  string  `json:"description"`
	BalanceAfter float64 `json:"balance_after"`
	CreatedAt    string  `json:"created_at"`
}

// UsageSummary is aggregate usage and cost for the organization.
type UsageSummary struct {
	TotalUnits  float64 `json:"total_units"`
	TotalCost   float64 `json:"total_cost"`
	RecordCount int     `json:"record_count"`
	Currency    string  `json:"currency"`
}

// Environment is one of Development, Staging or Production.
type Environment struct {
	UUID string `json:"uuid"`
	// Kind is the stable id to pass to WithEnvironment: dev, staging or prod.
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

// Provider is one entry in the model, transcriber and voice catalog.
type Provider struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	// ServiceType is llm, stt, tts or realtime.
	ServiceType  string           `json:"service_type"`
	DefaultModel string           `json:"default_model"`
	Models       []map[string]any `json:"models"`
	Voices       []map[string]any `json:"voices"`
	Languages    []map[string]any `json:"languages"`
	// Available reports whether it can be selected; an unavailable provider is
	// shown as "Soon" rather than hidden.
	Available bool `json:"available"`
}

// ProviderResources is one provider's live models and voices.
type ProviderResources struct {
	Key          string           `json:"key"`
	ServiceType  string           `json:"service_type"`
	DefaultModel string           `json:"default_model"`
	Source       string           `json:"source"`
	Models       []map[string]any `json:"models"`
	Voices       []map[string]any `json:"voices"`
}

// APIKey is a key for calling this API. The secret is never returned after
// creation; see CreatedAPIKey.
type APIKey struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	KeyPrefix  string   `json:"key_prefix"`
	IsActive   bool     `json:"is_active"`
	LastUsedAt string   `json:"last_used_at,omitempty"`
	CreatedAt  string   `json:"created_at,omitempty"`
	ExpiresAt  string   `json:"expires_at,omitempty"`
	Scopes     []string `json:"scopes,omitempty"`
}

// CreatedAPIKey is a newly created key, the one and only time the secret is
// returned.
type CreatedAPIKey struct {
	APIKey
	Key string `json:"key"`
}

// Organization is a workspace.
type Organization struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	// Region is immutable: where this organization's data lives.
	Region string `json:"region"`
}

// Region is one deployed stack data can live in.
type Region struct {
	Code  string `json:"code"`
	Label string `json:"label"`
	// APIBaseURL is empty for the stack you are already talking to.
	APIBaseURL string `json:"api_base_url"`
}

// VectorStore is a vector store over knowledge-base documents.
type VectorStore struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

// ChatToken is a short-lived chat token scoped to a workflow.
type ChatToken struct {
	Token        string `json:"token"`
	WorkflowUUID string `json:"workflow_uuid"`
	ExpiresIn    int    `json:"expires_in"`
}

// AnalyticsDayPoint is one day's call count and cost in an AnalyticsOverview.
type AnalyticsDayPoint struct {
	Day   string  `json:"day"`
	Calls int     `json:"calls"`
	Cost  float64 `json:"cost"`
}

// AnalyticsOverview is a high-level usage and cost summary over a time window.
type AnalyticsOverview struct {
	TotalCalls     int                 `json:"total_calls"`
	VoiceCalls     int                 `json:"voice_calls"`
	TextCalls      int                 `json:"text_calls"`
	CompletedCalls int                 `json:"completed_calls"`
	TotalMinutes   float64             `json:"total_minutes"`
	TotalCost      float64             `json:"total_cost"`
	CreditBalance  float64             `json:"credit_balance"`
	Currency       string              `json:"currency"`
	ByStatus       map[string]int      `json:"by_status"`
	ByDay          []AnalyticsDayPoint `json:"by_day"`
}
