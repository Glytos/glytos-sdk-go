package glytos

import (
	"context"
	"net/url"
	"strconv"
)

// TestSuitesService manages saved conversations replayed against an agent, to
// catch prompt regressions.
type TestSuitesService struct{ client *Client }

// List returns your test suites.
func (s *TestSuitesService) List(ctx context.Context) ([]TestSuite, error) {
	var out []TestSuite
	err := s.client.do(ctx, "GET", "/test-suites", nil, nil, &out)
	return out, err
}

// Create adds a suite of cases against one agent.
func (s *TestSuitesService) Create(ctx context.Context, workflowUUID, name string, cases []map[string]any) (*TestSuite, error) {
	body := map[string]any{"workflow_uuid": workflowUUID, "name": name}
	if cases != nil {
		body["cases"] = cases
	}
	var out TestSuite
	err := s.client.do(ctx, "POST", "/test-suites", body, nil, &out)
	return &out, err
}

// TestSuiteUpdate carries the editable parts of a suite. A nil field is left
// alone; cases are replaced whole rather than merged.
type TestSuiteUpdate struct {
	Name         *string          `json:"name,omitempty"`
	WorkflowUUID *string          `json:"workflow_uuid,omitempty"`
	Cases        []map[string]any `json:"cases,omitempty"`
}

// Update renames a suite, repoints it at another agent, or rewrites its cases.
func (s *TestSuitesService) Update(ctx context.Context, suiteUUID string, body TestSuiteUpdate) (*TestSuite, error) {
	var out TestSuite
	err := s.client.do(ctx, "PUT", "/test-suites/"+esc(suiteUUID), body, nil, &out)
	return &out, err
}

// Delete removes a suite.
func (s *TestSuitesService) Delete(ctx context.Context, suiteUUID string) error {
	return s.client.do(ctx, "DELETE", "/test-suites/"+esc(suiteUUID), nil, nil, nil)
}

// Run runs every case and reports which passed. It runs the agent, so it spends
// credit.
func (s *TestSuitesService) Run(ctx context.Context, suiteUUID string) (*TestSuiteRun, error) {
	var out TestSuiteRun
	err := s.client.do(ctx, "POST", "/test-suites/"+esc(suiteUUID)+"/run", nil, nil, &out)
	return &out, err
}

// BillingService exposes the credit balance, the ledger and usage.
type BillingService struct{ client *Client }

// Credits returns the current prepaid balance. Worth checking before a large
// outbound run: a call is refused below the minimum, so a campaign that runs out
// simply stops.
func (s *BillingService) Credits(ctx context.Context) (*CreditBalance, error) {
	var out CreditBalance
	err := s.client.do(ctx, "GET", "/billing/credits", nil, nil, &out)
	return &out, err
}

// Transactions returns the credit ledger, newest first. Pass an empty kind or a
// limit of zero for the server default.
func (s *BillingService) Transactions(ctx context.Context, kind string, limit int) ([]CreditTransaction, error) {
	query := url.Values{}
	if kind != "" {
		query.Set("kind", kind)
	}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	var out []CreditTransaction
	err := s.client.do(ctx, "GET", "/billing/credits/transactions", nil, query, &out)
	return out, err
}

// Usage returns aggregate usage and cost for the organization.
func (s *BillingService) Usage(ctx context.Context) (*UsageSummary, error) {
	var out UsageSummary
	err := s.client.do(ctx, "GET", "/billing/usage", nil, nil, &out)
	return &out, err
}

// EnvironmentsService lists Development, Staging and Production.
//
// Pass a Kind or a UUID to WithEnvironment to scope reads and calls; agents are
// created in Development whatever it is set to.
type EnvironmentsService struct{ client *Client }

// List returns the organization's three environments.
func (s *EnvironmentsService) List(ctx context.Context) ([]Environment, error) {
	var out []Environment
	err := s.client.do(ctx, "GET", "/environments", nil, nil, &out)
	return out, err
}

// ProvidersService exposes the model, transcriber and voice catalog.
type ProvidersService struct{ client *Client }

// List returns every provider and model, and whether it is available to you.
func (s *ProvidersService) List(ctx context.Context) ([]Provider, error) {
	var out []Provider
	err := s.client.do(ctx, "GET", "/providers", nil, nil, &out)
	return out, err
}

// Resources returns one provider's live models and voices, fetched from the
// provider itself where it publishes them. Pass an empty language for all of
// them; a long voice list is worth narrowing.
func (s *ProvidersService) Resources(ctx context.Context, serviceType, key, language string) (*ProviderResources, error) {
	query := url.Values{}
	if language != "" {
		query.Set("language", language)
	}
	var out ProviderResources
	err := s.client.do(ctx, "GET", "/providers/"+esc(serviceType)+"/"+esc(key)+"/resources", nil, query, &out)
	return &out, err
}

// APIKeysService manages keys for calling this API.
type APIKeysService struct{ client *Client }

// APIKeyCreateParams are the fields for APIKeysService.Create.
type APIKeyCreateParams struct {
	// Name labels the key (required).
	Name string
	// ExpiresInDays retires the key on its own. Zero means it never expires.
	ExpiresInDays int
	// Scopes bounds what the key may do, and cannot exceed what you hold. Leave
	// it nil and the key inherits your permissions, which means it stops working
	// if you leave the organization.
	Scopes []string
}

// List returns the keys on the organization. Secrets are never returned.
func (s *APIKeysService) List(ctx context.Context) ([]APIKey, error) {
	var out []APIKey
	err := s.client.do(ctx, "GET", "/api-keys", nil, nil, &out)
	return out, err
}

// Create makes a key. The secret is in the response and nowhere else, so store
// it now.
func (s *APIKeysService) Create(ctx context.Context, params APIKeyCreateParams) (*CreatedAPIKey, error) {
	body := map[string]any{"name": params.Name}
	if params.ExpiresInDays > 0 {
		body["expires_in_days"] = params.ExpiresInDays
	}
	if params.Scopes != nil {
		body["scopes"] = params.Scopes
	}
	var out CreatedAPIKey
	err := s.client.do(ctx, "POST", "/api-keys", body, nil, &out)
	return &out, err
}

// Delete revokes a key immediately.
func (s *APIKeysService) Delete(ctx context.Context, keyID int) error {
	return s.client.do(ctx, "DELETE", "/api-keys/"+esc(strconv.Itoa(keyID)), nil, nil, nil)
}

// OrganizationsService exposes the organization this key belongs to, and the
// regions data can live in.
type OrganizationsService struct{ client *Client }

// Retrieve returns the organization behind this API key.
func (s *OrganizationsService) Retrieve(ctx context.Context) (*Organization, error) {
	var out Organization
	err := s.client.do(ctx, "GET", "/organization", nil, nil, &out)
	return &out, err
}

// Update renames the organization. Its region is fixed at creation and cannot
// be changed.
func (s *OrganizationsService) Update(ctx context.Context, name string) (*Organization, error) {
	var out Organization
	err := s.client.do(ctx, "PATCH", "/organization", map[string]any{"name": name}, nil, &out)
	return &out, err
}

// Regions returns the regions this deployment offers. Each is a separate stack
// with its own base URL, so reaching an organization in another region means
// pointing WithBaseURL there with a key issued there.
func (s *OrganizationsService) Regions(ctx context.Context) ([]Region, error) {
	var out []Region
	err := s.client.do(ctx, "GET", "/regions", nil, nil, &out)
	return out, err
}
