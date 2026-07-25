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
	UUID   string `json:"uuid"`
	Name   string `json:"name"`
	Status string `json:"status,omitempty"`
}

// Tool is a reusable tool an agent can call (kind = http / static / mcp).
type Tool struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

// Document is a knowledge-base document.
type Document struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
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
