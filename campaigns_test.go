package glytos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestCampaignContactsAreSentAsPlainNumbers(t *testing.T) {
	// The API takes a list of strings. Sending objects was rejected with a 422,
	// so the shape of this field is worth a test of its own.
	ts := newTestServer(t)
	ts.body = `{"uuid":"c1","name":"March","status":"draft"}`

	_, err := ts.client.Campaigns.Create(context.Background(), CampaignCreateParams{
		Name:         "March",
		WorkflowUUID: "wf1",
		FromNumber:   "+15551230000",
		Contacts:     []string{"+15551230001", "0532 123 45 67"},
	})
	if err != nil {
		t.Fatal(err)
	}

	var body struct {
		Contacts []string `json:"contacts"`
	}
	if err := json.Unmarshal(ts.last.body, &body); err != nil {
		t.Fatalf("contacts did not decode as a list of strings: %v", err)
	}
	if len(body.Contacts) != 2 || body.Contacts[0] != "+15551230001" {
		t.Fatalf("unexpected contacts: %v", body.Contacts)
	}
}

func TestCampaignCreateOmitsUnsetScheduling(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `{"uuid":"c1"}`

	_, err := ts.client.Campaigns.Create(context.Background(), CampaignCreateParams{
		Name:         "March",
		WorkflowUUID: "wf1",
		FromNumber:   "+15551230000",
	})
	if err != nil {
		t.Fatal(err)
	}

	var body map[string]any
	if err := json.Unmarshal(ts.last.body, &body); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{
		"contacts", "contacts_csv", "scheduled_at", "call_window_start",
		"call_window_end", "timezone", "suppression_policy",
		"override_caller_requests",
	} {
		if _, present := body[key]; present {
			t.Fatalf("%q was sent although it was never set", key)
		}
	}
}

func TestCampaignCreateCarriesTheCallWindow(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `{"uuid":"c1"}`

	_, err := ts.client.Campaigns.Create(context.Background(), CampaignCreateParams{
		Name:                   "March",
		WorkflowUUID:           "wf1",
		FromNumber:             "+15551230000",
		ContactsCSV:            "phone,name\n+15551230001,Ada\n",
		ScheduledAt:            "2026-03-01T09:00:00Z",
		CallWindowStart:        "09:00",
		CallWindowEnd:          "20:00",
		Timezone:               "Europe/Istanbul",
		SuppressionPolicy:      "ignore",
		OverrideCallerRequests: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	var body map[string]any
	if err := json.Unmarshal(ts.last.body, &body); err != nil {
		t.Fatal(err)
	}
	if body["timezone"] != "Europe/Istanbul" || body["call_window_end"] != "20:00" {
		t.Fatalf("call window not sent: %v", body)
	}
	if body["override_caller_requests"] != true {
		t.Fatalf("override not sent: %v", body)
	}
}

func TestCampaignDetailDecodesContactOutcomes(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `{"uuid":"c1","name":"March","status":"completed",
		"suppression_policy":"strict","contacts":[
		{"phone":"+15551230001","status":"answered","session_uuid":"s1",
		 "variables":{"name":"Ada"}},
		{"phone":"+15551230002","status":"suppressed"}]}`

	campaign, err := ts.client.Campaigns.Retrieve(context.Background(), "c1")
	if err != nil {
		t.Fatal(err)
	}
	if campaign.Name != "March" || campaign.SuppressionPolicy != "strict" {
		t.Fatalf("campaign fields lost: %+v", campaign.Campaign)
	}
	if len(campaign.Contacts) != 2 {
		t.Fatalf("expected 2 contacts, got %d", len(campaign.Contacts))
	}
	if campaign.Contacts[0].SessionUUID != "s1" {
		t.Fatal("the answered contact lost its session")
	}
	if campaign.Contacts[0].Variables["name"] != "Ada" {
		t.Fatal("contact variables lost")
	}
}

func TestCampaignStopAndDelete(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `{"uuid":"c1","status":"stopped"}`

	campaign, err := ts.client.Campaigns.Stop(context.Background(), "c1")
	if err != nil {
		t.Fatal(err)
	}
	if ts.last.method != http.MethodPost || ts.last.path != "/telephony/campaigns/c1/stop" {
		t.Fatalf("unexpected request: %s %s", ts.last.method, ts.last.path)
	}
	if campaign.Status != "stopped" {
		t.Fatalf("unexpected status %q", campaign.Status)
	}

	ts.status = http.StatusNoContent
	ts.body = ""
	if err := ts.client.Campaigns.Delete(context.Background(), "c1"); err != nil {
		t.Fatal(err)
	}
	if ts.last.method != http.MethodDelete || ts.last.path != "/telephony/campaigns/c1" {
		t.Fatalf("unexpected request: %s %s", ts.last.method, ts.last.path)
	}
}

func TestAddContactsSendsCSVText(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `{"added":2,"skipped":1,"rejected":0,"phone_column":"telefon"}`

	result, err := ts.client.Campaigns.AddContacts(
		context.Background(), "c1", "telefon;isim\n+905321234567;Ada\n")
	if err != nil {
		t.Fatal(err)
	}
	if ts.last.path != "/telephony/campaigns/c1/contacts/sync" {
		t.Fatalf("unexpected path %q", ts.last.path)
	}
	var body map[string]any
	if err := json.Unmarshal(ts.last.body, &body); err != nil {
		t.Fatal(err)
	}
	if _, present := body["source_url"]; present {
		t.Fatal("a CSV upload must not also claim a source url")
	}
	// The column actually read is what separates a file parsed from the wrong
	// column from one that could not be parsed at all.
	if result.PhoneColumn != "telefon" || result.Added != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestSuppressionPreviewReportsCallerRequests(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `{"contacts":100,"suppressed_total":12,"caller_requested":4,
		"reached_if_strict":88,"reached_if_transactional":92,
		"reached_if_ignore":96,"reached_if_override":100}`

	preview, err := ts.client.Campaigns.PreviewSuppression(
		context.Background(), []string{"+15551230001"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if ts.last.path != "/telephony/campaigns/suppression-preview" {
		t.Fatalf("unexpected path %q", ts.last.path)
	}
	if preview.CallerRequested != 4 || preview.ReachedIfStrict != 88 {
		t.Fatalf("unexpected preview: %+v", preview)
	}
}

func TestDncListPassesSearchAndPaging(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `{"items":[{"uuid":"d1","phone":"+15551230001","source":"agent",
		"scope":"all"}],"total":1}`

	list, err := ts.client.Dnc.List(context.Background(), DncListParams{
		Search: "0555", Limit: 50, Offset: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if ts.last.rawQuery != "limit=50&offset=100&search=0555" {
		t.Fatalf("unexpected query %q", ts.last.rawQuery)
	}
	if len(list.Items) != 1 || list.Items[0].Source != "agent" {
		t.Fatalf("unexpected entries: %+v", list.Items)
	}
}

func TestDncListOmitsUnsetFilters(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `{"items":[],"total":0}`

	if _, err := ts.client.Dnc.List(context.Background(), DncListParams{}); err != nil {
		t.Fatal(err)
	}
	if ts.last.rawQuery != "" {
		t.Fatalf("expected no query, got %q", ts.last.rawQuery)
	}
}

func TestDncScopeAndRemovalCarryThePhoneNumberIntact(t *testing.T) {
	// A phone number is a path parameter, so the leading "+" has to reach the
	// server as itself and not as a space or an escape the server re-reads.
	ts := newTestServer(t)
	ts.body = `{"uuid":"d1","phone":"+15551230001","source":"manual","scope":"marketing"}`

	entry, err := ts.client.Dnc.SetScope(context.Background(), "+15551230001", "marketing")
	if err != nil {
		t.Fatal(err)
	}
	if ts.last.method != http.MethodPatch {
		t.Fatalf("unexpected method %q", ts.last.method)
	}
	if ts.last.path != "/dnc/+15551230001" {
		t.Fatalf("phone did not survive the path: %q", ts.last.path)
	}
	if entry.Scope != "marketing" {
		t.Fatalf("unexpected scope %q", entry.Scope)
	}

	ts.status = http.StatusNoContent
	ts.body = ""
	if err := ts.client.Dnc.Remove(context.Background(), "+15551230001"); err != nil {
		t.Fatal(err)
	}
	if ts.last.method != http.MethodDelete || ts.last.path != "/dnc/+15551230001" {
		t.Fatalf("unexpected request: %s %s", ts.last.method, ts.last.path)
	}
}

func TestDncSendsAnEmptyReasonRatherThanNull(t *testing.T) {
	// The server takes a plain string, not a nullable one, so a null is a 422.
	// A plain Go string cannot be null, which is why reason is not a pointer.
	ts := newTestServer(t)
	ts.body = `{"uuid":"d1","phone":"+15551230001"}`

	if _, err := ts.client.Dnc.Add(context.Background(), "+15551230001", ""); err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(ts.last.body, &body); err != nil {
		t.Fatal(err)
	}
	if body["reason"] != "" {
		t.Fatalf("expected an empty reason, got %v", body["reason"])
	}
}

func TestDncImportReportsWhatItDid(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `{"added":8,"duplicates":1,"rejected":2}`

	result, err := ts.client.Dnc.Import(
		context.Background(), []string{"+15551230001", "not a number"}, "CRM export")
	if err != nil {
		t.Fatal(err)
	}
	if ts.last.path != "/dnc/import" {
		t.Fatalf("unexpected path %q", ts.last.path)
	}
	if result.Added != 8 || result.Rejected != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
}
