package glytos

import (
	"context"
	"net/url"
	"strconv"
)

// DncService manages the numbers your organization must not call.
//
// Every outbound call checks this list first, whether it comes from a campaign,
// from this SDK, or from an agent. Agents add to it themselves when a caller
// asks not to be contacted again.
type DncService struct{ client *Client }

// DncList is a page of the do-not-call list.
type DncList struct {
	Items []DncEntry `json:"items"`
	Total int        `json:"total"`
}

// DncListParams filter a page of the list.
type DncListParams struct {
	// Search is normalized before matching, so a number typed the way it
	// appears on a contact list finds the entry stored in international form.
	Search string
	Limit  int
	Offset int
}

// DncImportResult is what a bulk add did.
type DncImportResult struct {
	Added int `json:"added"`
	// Duplicates were already on the list; Rejected were not phone numbers.
	Duplicates int `json:"duplicates"`
	Rejected   int `json:"rejected"`
}

// List returns the numbers on the list, newest first.
func (s *DncService) List(ctx context.Context, params DncListParams) (*DncList, error) {
	query := url.Values{}
	if params.Search != "" {
		query.Set("search", params.Search)
	}
	if params.Limit > 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Offset > 0 {
		query.Set("offset", strconv.Itoa(params.Offset))
	}
	var out DncList
	err := s.client.do(ctx, "GET", "/dnc", nil, query, &out)
	return &out, err
}

// Add suppresses a number. Any spelling is accepted and stored in
// international form. Adding one already on the list returns the existing
// entry rather than failing.
func (s *DncService) Add(ctx context.Context, phone, reason string) (*DncEntry, error) {
	var out DncEntry
	body := map[string]any{"phone": phone, "reason": reason}
	err := s.client.do(ctx, "POST", "/dnc", body, nil, &out)
	return &out, err
}

// Import suppresses many numbers at once, e.g. a list exported from your CRM.
func (s *DncService) Import(ctx context.Context, phones []string, reason string) (*DncImportResult, error) {
	var out DncImportResult
	body := map[string]any{"phones": phones, "reason": reason}
	err := s.client.do(ctx, "POST", "/dnc/import", body, nil, &out)
	return &out, err
}

// SetScope changes how far a suppression reaches: "all" for every call, or
// "marketing" to allow a transactional call about the person's own order.
func (s *DncService) SetScope(ctx context.Context, phone, scope string) (*DncEntry, error) {
	var out DncEntry
	body := map[string]any{"scope": scope}
	err := s.client.do(ctx, "PATCH", "/dnc/"+esc(phone), body, nil, &out)
	return &out, err
}

// Remove takes a number off the list, so it can be called again.
func (s *DncService) Remove(ctx context.Context, phone string) error {
	return s.client.do(ctx, "DELETE", "/dnc/"+esc(phone), nil, nil, nil)
}
