package glytos

import (
	"context"
	"encoding/json"
	"testing"
)

// The resources added after the first release. Same recording test server as
// the rest of the suite: assert the method, path and body, so a wrong path or a
// renamed field is caught without a live API.

func TestSipTrunkCreatePostsTheLogin(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `{"uuid":"trunk_1","name":"Main line","status":"pending"}`

	_, err := ts.client.SipTrunks.Create(context.Background(), SipTrunkCreateParams{
		Username: "line-1",
		Password: "secret",
		Preset:   "netgsm",
	})
	if err != nil {
		t.Fatal(err)
	}

	if ts.last.method != "POST" || ts.last.path != "/telephony/sip-trunks" {
		t.Fatalf("unexpected request: %s %s", ts.last.method, ts.last.path)
	}
	var body map[string]any
	if err := json.Unmarshal(ts.last.body, &body); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"username", "password", "preset"} {
		if _, ok := body[key]; !ok {
			t.Fatalf("%s missing from the body: %v", key, body)
		}
	}
	// Unset optional fields stay out rather than going as zero values, which the
	// API would read as a real port of 0.
	if _, ok := body["sip_port"]; ok {
		t.Fatalf("sip_port was sent unset: %v", body)
	}
}

func TestSipTrunkTestReportsReachableSeparatelyFromOK(t *testing.T) {
	// A carrier that refused the credentials is a different problem from one that
	// never answered, and only the first is worth changing the password over.
	ts := newTestServer(t)
	ts.body = `{"ok":false,"detail":"no reply","reachable":false}`

	result, err := ts.client.SipTrunks.Test(context.Background(), "trunk_1")
	if err != nil {
		t.Fatal(err)
	}
	if ts.last.path != "/telephony/sip-trunks/trunk_1/test" {
		t.Fatalf("unexpected path: %s", ts.last.path)
	}
	if result.OK || result.Reachable {
		t.Fatalf("expected a refused, unreachable result: %+v", result)
	}
}

func TestImportNumberCanNameASipTrunk(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `{"uuid":"num_1","e164":"+905321234567"}`

	_, err := ts.client.PhoneNumbers.ImportNumber(context.Background(), ImportNumberParams{
		E164:         "+905321234567",
		SipTrunkUUID: "trunk_1",
	})
	if err != nil {
		t.Fatal(err)
	}

	var body map[string]any
	if err := json.Unmarshal(ts.last.body, &body); err != nil {
		t.Fatal(err)
	}
	if body["sip_trunk_uuid"] != "trunk_1" {
		t.Fatalf("sip_trunk_uuid not sent: %v", body)
	}
	if _, ok := body["provider"]; ok {
		t.Fatalf("provider was sent unset: %v", body)
	}
}

func TestConnectionRunAddressesTheConnection(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `{"result":{"ok":true}}`

	_, err := ts.client.Integrations.Connections.Run(
		context.Background(), "conn_1", "post_message", map[string]any{"text": "A lead came in"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if ts.last.path != "/integrations/connections/conn_1/run" {
		t.Fatalf("unexpected path: %s", ts.last.path)
	}
	var body map[string]any
	if err := json.Unmarshal(ts.last.body, &body); err != nil {
		t.Fatal(err)
	}
	if body["action"] != "post_message" {
		t.Fatalf("action not sent: %v", body)
	}
}

func TestConnectionRunSendsAnObjectForNoParams(t *testing.T) {
	// A nil map would encode as null, which the API reads as a missing object
	// rather than "no parameters".
	ts := newTestServer(t)
	ts.body = `{"result":{}}`

	if _, err := ts.client.Integrations.Connections.Run(
		context.Background(), "conn_1", "ping", nil,
	); err != nil {
		t.Fatal(err)
	}
	var body struct {
		Params map[string]any `json:"params"`
	}
	if err := json.Unmarshal(ts.last.body, &body); err != nil {
		t.Fatal(err)
	}
	if body.Params == nil {
		t.Fatalf("params encoded as null: %s", ts.last.body)
	}
}

func TestAutomationCreateCarriesTheTriggerAndTemplate(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `{"uuid":"auto_1"}`

	_, err := ts.client.Automations.Create(context.Background(), AutomationCreateParams{
		Name:            "Tell sales",
		TriggerEvent:    "session.completed",
		ConnectionUUID:  "conn_1",
		Action:          "post_message",
		PayloadTemplate: map[string]any{"text": "Call from {{from_number}}"},
	})
	if err != nil {
		t.Fatal(err)
	}

	var body map[string]any
	if err := json.Unmarshal(ts.last.body, &body); err != nil {
		t.Fatal(err)
	}
	if body["trigger_event"] != "session.completed" {
		t.Fatalf("trigger not sent: %v", body)
	}
	// Conditions were not given, so they are absent rather than an empty object.
	if _, ok := body["conditions"]; ok {
		t.Fatalf("conditions sent unset: %v", body)
	}
}

func TestTestSuiteRunPostsToTheSuite(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `{"suite_uuid":"s1","passed":false,"total":3,"passed_count":2,"results":[]}`

	result, err := ts.client.TestSuites.Run(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}
	if ts.last.method != "POST" || ts.last.path != "/test-suites/s1/run" {
		t.Fatalf("unexpected request: %s %s", ts.last.method, ts.last.path)
	}
	if result.PassedCount != 2 {
		t.Fatalf("passed_count = %d, want 2", result.PassedCount)
	}
}

func TestBillingReadsTheBalanceAndFiltersTheLedger(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `{"balance":12.5,"currency":"USD"}`

	balance, err := ts.client.Billing.Credits(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if balance.Balance != 12.5 {
		t.Fatalf("balance = %v, want 12.5", balance.Balance)
	}

	ts.body = `[]`
	if _, err := ts.client.Billing.Transactions(context.Background(), "debit", 10); err != nil {
		t.Fatal(err)
	}
	if ts.last.rawQuery != "kind=debit&limit=10" {
		t.Fatalf("query = %q", ts.last.rawQuery)
	}
}

func TestAPIKeyCreateOmitsUnstatedLimits(t *testing.T) {
	// Omitting both is exactly the behaviour keys have always had, so an SDK that
	// sent zero values would change what an unchanged caller gets.
	ts := newTestServer(t)
	ts.body = `{"id":1,"name":"CI","key":"gly_x"}`

	if _, err := ts.client.APIKeys.Create(context.Background(), APIKeyCreateParams{Name: "CI"}); err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(ts.last.body, &body); err != nil {
		t.Fatal(err)
	}
	if len(body) != 1 || body["name"] != "CI" {
		t.Fatalf("unexpected body: %v", body)
	}

	if _, err := ts.client.APIKeys.Create(context.Background(), APIKeyCreateParams{
		Name:          "CI",
		ExpiresInDays: 90,
		Scopes:        []string{"workflow:read"},
	}); err != nil {
		t.Fatal(err)
	}
	body = nil
	if err := json.Unmarshal(ts.last.body, &body); err != nil {
		t.Fatal(err)
	}
	if body["expires_in_days"] != float64(90) {
		t.Fatalf("expires_in_days not sent: %v", body)
	}
}

func TestDiscoverMCPReturnsTheToolListNotTheEnvelope(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `{"tools":[{"name":"search"},{"name":"fetch"}]}`

	tools, err := ts.client.Tools.DiscoverMCP(context.Background(), "https://mcp.example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	if ts.last.path != "/tools/mcp/discover" {
		t.Fatalf("unexpected path: %s", ts.last.path)
	}
	if len(tools) != 2 || tools[0].Name != "search" {
		t.Fatalf("unexpected tools: %+v", tools)
	}
}

func TestKnowledgeBaseDocumentsCanBeReadAndDeleted(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `{"id":7,"name":"Refunds","content":"..."}`

	if _, err := ts.client.KnowledgeBase.RetrieveDocument(context.Background(), 7); err != nil {
		t.Fatal(err)
	}
	if ts.last.method != "GET" || ts.last.path != "/knowledge-base/documents/7" {
		t.Fatalf("unexpected request: %s %s", ts.last.method, ts.last.path)
	}

	ts.body = ``
	ts.status = 204
	if err := ts.client.KnowledgeBase.DeleteDocument(context.Background(), 7); err != nil {
		t.Fatal(err)
	}
	if ts.last.method != "DELETE" {
		t.Fatalf("method = %s, want DELETE", ts.last.method)
	}
}

func TestImportsConnectAndPullCarryTheOtherPlatformKey(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `{"agents":[]}`

	if _, err := ts.client.Imports.Connect(context.Background(), "vapi", "vapi_key"); err != nil {
		t.Fatal(err)
	}
	if ts.last.path != "/imports/vapi/connect" {
		t.Fatalf("unexpected path: %s", ts.last.path)
	}

	ts.body = `{"imports":[]}`
	if _, err := ts.client.Imports.Pull(context.Background(), "vapi", "vapi_key", []string{"a1"}); err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(ts.last.body, &body); err != nil {
		t.Fatal(err)
	}
	if body["api_key"] != "vapi_key" {
		t.Fatalf("api_key not sent: %v", body)
	}
}

func TestCallControlHelpersSpellOutTheAction(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `{"ok":true}`

	if _, err := ts.client.Calls.Say(context.Background(), "call_1", "One moment"); err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(ts.last.body, &body); err != nil {
		t.Fatal(err)
	}
	if body["action"] != "say" || body["text"] != "One moment" {
		t.Fatalf("unexpected body: %v", body)
	}

	if _, err := ts.client.Calls.End(context.Background(), "call_1"); err != nil {
		t.Fatal(err)
	}
	// A fresh map: Unmarshal merges into a non-nil one rather than replacing it,
	// so the previous turn's "text" would still be sitting there.
	body = nil
	if err := json.Unmarshal(ts.last.body, &body); err != nil {
		t.Fatal(err)
	}
	if len(body) != 1 || body["action"] != "end" {
		t.Fatalf("end carried more than the action: %v", body)
	}
}

func TestEnvironmentsAndProvidersRead(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `[]`

	if _, err := ts.client.Environments.List(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ts.last.path != "/environments" {
		t.Fatalf("unexpected path: %s", ts.last.path)
	}

	ts.body = `{"key":"cartesia","service_type":"tts","default_model":"","source":"","models":[],"voices":[]}`
	if _, err := ts.client.Providers.Resources(context.Background(), "tts", "cartesia", "tr"); err != nil {
		t.Fatal(err)
	}
	if ts.last.path != "/providers/tts/cartesia/resources" {
		t.Fatalf("unexpected path: %s", ts.last.path)
	}
	if ts.last.rawQuery != "language=tr" {
		t.Fatalf("query = %q", ts.last.rawQuery)
	}
}

func TestBackgroundToolFlagIsSentAndOmitted(t *testing.T) {
	ts := newTestServer(t)
	ts.body = `{"uuid":"t1","name":"Legal search","kind":"mcp"}`

	if _, err := ts.client.Tools.Create(context.Background(), ToolCreateParams{
		Name: "Legal search", Kind: "mcp", RunInBackground: true,
	}); err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(ts.last.body, &body); err != nil {
		t.Fatal(err)
	}
	if body["run_in_background"] != true {
		t.Fatalf("run_in_background not sent: %v", body)
	}

	if _, err := ts.client.Tools.Create(context.Background(), ToolCreateParams{
		Name: "Order status", Kind: "http",
	}); err != nil {
		t.Fatal(err)
	}
	body = nil
	if err := json.Unmarshal(ts.last.body, &body); err != nil {
		t.Fatal(err)
	}
	if _, sent := body["run_in_background"]; sent {
		t.Fatalf("run_in_background sent for a foreground tool: %v", body)
	}
}

func TestUpdatingLeavesTheBackgroundFlagAloneUnlessSet(t *testing.T) {
	// A pointer, so turning it OFF is expressible: a plain bool would make "off"
	// and "not mentioned" the same request.
	ts := newTestServer(t)
	ts.body = `{"uuid":"t1","name":"Legal search","kind":"mcp"}`

	if _, err := ts.client.Tools.Update(context.Background(), "t1", ToolUpdateParams{
		Name: "Renamed",
	}); err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(ts.last.body, &body); err != nil {
		t.Fatal(err)
	}
	if _, sent := body["run_in_background"]; sent {
		t.Fatalf("run_in_background sent when it was not set: %v", body)
	}

	off := false
	if _, err := ts.client.Tools.Update(context.Background(), "t1", ToolUpdateParams{
		RunInBackground: &off,
	}); err != nil {
		t.Fatal(err)
	}
	body = nil
	if err := json.Unmarshal(ts.last.body, &body); err != nil {
		t.Fatal(err)
	}
	if body["run_in_background"] != false {
		t.Fatalf("run_in_background=false not sent: %v", body)
	}
}
