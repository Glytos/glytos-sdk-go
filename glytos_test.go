package glytos

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type recordedRequest struct {
	method      string
	path        string
	escapedPath string
	rawQuery    string
	header      http.Header
	body        []byte
	contentType string
}

type testServer struct {
	server *httptest.Server
	client *Client
	last   *recordedRequest
	status int
	body   string
	header map[string]string
}

func newTestServer(t *testing.T, opts ...Option) *testServer {
	t.Helper()
	ts := &testServer{last: &recordedRequest{}, status: http.StatusOK, body: ""}
	ts.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		*ts.last = recordedRequest{
			method:      r.Method,
			path:        r.URL.Path,
			escapedPath: r.URL.EscapedPath(),
			rawQuery:    r.URL.RawQuery,
			header:      r.Header.Clone(),
			body:        b,
			contentType: r.Header.Get("Content-Type"),
		}
		for k, v := range ts.header {
			w.Header().Set(k, v)
		}
		w.WriteHeader(ts.status)
		_, _ = io.WriteString(w, ts.body)
	}))
	t.Cleanup(ts.server.Close)
	all := append([]Option{WithBaseURL(ts.server.URL)}, opts...)
	ts.client = New("gly_test", all...)
	return ts
}

func TestAuthHeaders(t *testing.T) {
	ts := newTestServer(t)
	if _, err := ts.client.Workflows.List(context.Background(), nil); err != nil {
		t.Fatalf("List: %v", err)
	}
	if got := ts.last.header.Get("X-API-Key"); got != "gly_test" {
		t.Fatalf("X-API-Key = %q, want gly_test", got)
	}
	if got := ts.last.header.Get("Accept"); got != "application/json" {
		t.Fatalf("Accept = %q, want application/json", got)
	}
	if got := ts.last.header.Get("X-Environment-Id"); got != "" {
		t.Fatalf("X-Environment-Id = %q, want empty (no environment configured)", got)
	}
}

func TestEnvironmentHeader(t *testing.T) {
	ts := newTestServer(t, WithEnvironment("prod"))
	if _, err := ts.client.Workflows.List(context.Background(), nil); err != nil {
		t.Fatalf("List: %v", err)
	}
	if got := ts.last.header.Get("X-Environment-Id"); got != "prod" {
		t.Fatalf("X-Environment-Id = %q, want prod", got)
	}
}

func TestListDropsEmptyQuery(t *testing.T) {
	ts := newTestServer(t)
	if _, err := ts.client.Workflows.List(context.Background(), &WorkflowListParams{Environment: ""}); err != nil {
		t.Fatalf("List: %v", err)
	}
	if ts.last.rawQuery != "" {
		t.Fatalf("rawQuery = %q, want empty (no params provided)", ts.last.rawQuery)
	}

	if _, err := ts.client.Workflows.List(context.Background(), &WorkflowListParams{Archived: Bool(true)}); err != nil {
		t.Fatalf("List archived: %v", err)
	}
	if ts.last.rawQuery != "archived=true" {
		t.Fatalf("rawQuery = %q, want archived=true", ts.last.rawQuery)
	}
}

func TestErrorMapping(t *testing.T) {
	ts := newTestServer(t)
	ts.status = http.StatusNotFound
	ts.body = `{"error":{"code":"not_found","message":"agent not found"}}`
	ts.header = map[string]string{"X-Request-Id": "req_abc123"}

	_, err := ts.client.Workflows.Retrieve(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected an error on a 404 response")
	}
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is %T, want *glytos.Error", err)
	}
	if apiErr.Status != http.StatusNotFound {
		t.Errorf("Status = %d, want 404", apiErr.Status)
	}
	if apiErr.Code != "not_found" {
		t.Errorf("Code = %q, want not_found", apiErr.Code)
	}
	if apiErr.Message != "agent not found" {
		t.Errorf("Message = %q, want 'agent not found'", apiErr.Message)
	}
	if apiErr.RequestID != "req_abc123" {
		t.Errorf("RequestID = %q, want req_abc123", apiErr.RequestID)
	}
}

func TestPromoteSendsTargetEnvironment(t *testing.T) {
	ts := newTestServer(t)
	if _, err := ts.client.Workflows.Promote(context.Background(), "wf_1", "env_prod"); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if ts.last.method != http.MethodPost {
		t.Errorf("method = %s, want POST", ts.last.method)
	}
	if ts.last.path != "/workflows/wf_1/promote" {
		t.Errorf("path = %q, want /workflows/wf_1/promote", ts.last.path)
	}
	body := decodeBody(t, ts.last.body)
	if body["target_environment_id"] != "env_prod" {
		t.Errorf("target_environment_id = %v, want env_prod", body["target_environment_id"])
	}
}

func TestInstantSendsQueryNotBody(t *testing.T) {
	ts := newTestServer(t)
	if _, err := ts.client.PhoneNumbers.Instant(context.Background(), "US", "twilio"); err != nil {
		t.Fatalf("Instant: %v", err)
	}
	if ts.last.method != http.MethodPost {
		t.Errorf("method = %s, want POST", ts.last.method)
	}
	if ts.last.path != "/telephony/numbers/instant" {
		t.Errorf("path = %q", ts.last.path)
	}
	if len(ts.last.body) != 0 {
		t.Errorf("body = %q, want empty (instant uses query params)", ts.last.body)
	}
	if ts.last.contentType != "" {
		t.Errorf("Content-Type = %q, want empty (no request body)", ts.last.contentType)
	}
	q, err := url.ParseQuery(ts.last.rawQuery)
	if err != nil {
		t.Fatalf("parse query: %v", err)
	}
	if q.Get("country") != "US" || q.Get("provider") != "twilio" {
		t.Errorf("query = %v, want country=US provider=twilio", q)
	}
}

func TestPathParamsAreEscaped(t *testing.T) {
	ts := newTestServer(t)
	if _, err := ts.client.Workflows.Retrieve(context.Background(), "a/b?c#d"); err != nil {
		t.Fatalf("Retrieve: %v", err)
	}
	if ts.last.escapedPath != "/workflows/a%2Fb%3Fc%23d" {
		t.Errorf("escaped path = %q, want /workflows/a%%2Fb%%3Fc%%23d", ts.last.escapedPath)
	}
	// The raw query must stay empty: the "?c" from the uuid must not leak into it.
	if ts.last.rawQuery != "" {
		t.Errorf("rawQuery = %q, want empty", ts.last.rawQuery)
	}
}

func TestWebhookUpdateOmitsUnsetFields(t *testing.T) {
	ts := newTestServer(t)
	if _, err := ts.client.Webhooks.Update(context.Background(), 7, WebhookUpdateParams{URL: "https://example.com/hook"}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if ts.last.method != http.MethodPatch {
		t.Errorf("method = %s, want PATCH", ts.last.method)
	}
	if ts.last.path != "/webhooks/endpoints/7" {
		t.Errorf("path = %q, want /webhooks/endpoints/7", ts.last.path)
	}
	body := decodeBody(t, ts.last.body)
	if len(body) != 1 {
		t.Fatalf("body = %v, want only the url key", body)
	}
	if body["url"] != "https://example.com/hook" {
		t.Errorf("url = %v", body["url"])
	}
}

func TestRetrieveDecodesTypedResponse(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `{"uuid":"wf_9","name":"Support","mode":"prompt","extra":"ignored"}`
	wf, err := ts.client.Workflows.Retrieve(context.Background(), "wf_9")
	if err != nil {
		t.Fatalf("Retrieve: %v", err)
	}
	if wf.UUID != "wf_9" || wf.Name != "Support" || wf.Mode != "prompt" {
		t.Errorf("decoded = %+v", wf)
	}
}

func TestPaginatedListReturnsItems(t *testing.T) {
	// /calls and /webhooks/deliveries wrap results in an {items, total, ...}
	// envelope; the SDK must return the items, not fail to decode the object.
	ts := newTestServer(t)

	ts.body = `{"items":[{"uuid":"call_1","status":"completed"}],"total":1,"limit":50,"offset":0}`
	calls, err := ts.client.Calls.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("Calls.List: %v", err)
	}
	if len(calls) != 1 || calls[0].UUID != "call_1" {
		t.Fatalf("Calls.List = %+v, want one call call_1", calls)
	}

	ts.body = `{"items":[{"id":7,"event_type":"call.completed","status":"success"}],"total":1}`
	deliveries, err := ts.client.Webhooks.Deliveries(context.Background(), nil)
	if err != nil {
		t.Fatalf("Webhooks.Deliveries: %v", err)
	}
	if len(deliveries) != 1 || deliveries[0].ID != 7 {
		t.Fatalf("Webhooks.Deliveries = %+v, want one delivery id 7", deliveries)
	}
}

func TestStartSessionAlwaysSendsBody(t *testing.T) {
	ts := newTestServer(t)
	if _, err := ts.client.Workflows.StartSession(context.Background(), "wf_1", nil); err != nil {
		t.Fatalf("StartSession: %v", err)
	}
	if ts.last.contentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ts.last.contentType)
	}
	if string(ts.last.body) != "{}" {
		t.Errorf("body = %q, want {}", ts.last.body)
	}
}

func TestAnalyticsOverviewOmitsDefaultDays(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `{"total_calls":3,"currency":"USD","by_day":[{"day":"2026-07-25","calls":3,"cost":1.5}]}`
	ov, err := ts.client.Analytics.Overview(context.Background(), 0)
	if err != nil {
		t.Fatalf("Overview: %v", err)
	}
	if ts.last.rawQuery != "" {
		t.Errorf("rawQuery = %q, want empty for days<=0", ts.last.rawQuery)
	}
	if ov.TotalCalls != 3 || ov.Currency != "USD" || len(ov.ByDay) != 1 {
		t.Errorf("decoded overview = %+v", ov)
	}

	if _, err := ts.client.Analytics.Overview(context.Background(), 30); err != nil {
		t.Fatalf("Overview 30: %v", err)
	}
	if ts.last.rawQuery != "days=30" {
		t.Errorf("rawQuery = %q, want days=30", ts.last.rawQuery)
	}
}

func decodeBody(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("decode body %q: %v", raw, err)
	}
	return m
}
