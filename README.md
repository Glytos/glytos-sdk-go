# glytos-sdk-go

[![CI](https://github.com/Glytos/glytos-sdk-go/actions/workflows/ci.yml/badge.svg)](https://github.com/Glytos/glytos-sdk-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/Glytos/glytos-sdk-go.svg)](https://pkg.go.dev/github.com/Glytos/glytos-sdk-go)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

The official [Glytos](https://glytos.com) server SDK for Go.

Call the Glytos API from your backend with an API key. Build agents once and run
them as **text** or as **voice**: hold a threaded conversation, stream a reply as it
is written, place phone calls, mint browser web-call tokens, manage numbers, and
verify webhooks. Zero dependencies (standard library only), fully typed,
context-aware.

> Never ship an API key to the browser. For in-browser voice, use the `@glytos/web`
> package with a short-lived token you mint here via `client.Calls.WebToken(...)`.

## Install

```bash
go get github.com/Glytos/glytos-sdk-go
```

Requires Go 1.21 or newer.

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"log"

	glytos "github.com/Glytos/glytos-sdk-go"
)

func main() {
	client := glytos.New("gly_...")
	ctx := context.Background()

	// List your agents
	agents, err := client.Workflows.List(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Mint a web-call token for the browser
	token, err := client.Calls.WebToken(ctx, glytos.WebTokenParams{
		WorkflowUUID: agents[0].UUID,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(token.Token, token.WSURL)
}
```

## Configuration

`New` takes your API key and optional functional options:

```go
client := glytos.New("gly_...",
	glytos.WithEnvironment("prod"),        // "dev" / "staging" / "prod" or an env uuid
	glytos.WithBaseURL("https://api.glytos.com/api/v1"),
	glytos.WithHTTPClient(&http.Client{Timeout: 60 * time.Second}),
)
```

Every method takes a `context.Context` as its first argument, so you control
timeouts and cancellation.

### Text conversations

An agent is one definition; nothing forces it to do both text and voice. For text,
a thread holds the conversation and a run is one turn on it:

```go
thread, _ := client.Threads.Create(ctx, agentUUID, nil)
reply, _ := client.Threads.Runs.Create(ctx, *thread, &glytos.TurnParams{
    Content: "What are your opening hours?",
})
```

Stream a long answer instead of waiting for it:

```go
err := client.Threads.Runs.Stream(ctx, *thread,
    &glytos.TurnParams{Content: "Summarise the policy"},
    func(event glytos.StreamEvent) error {
        if event.Type == "token" {
            fmt.Print(event.Delta)
        }
        return nil
    })
```

Returning an error from the callback stops the stream.

## Resources

| Namespace | Methods |
| --- | --- |
| `client.Agents` (alias `Workflows`) | `List`, `Retrieve`, `Create`, `Rename`, `Duplicate`, `Archive`, `Unarchive`, `Promote`, `Versions`, `UpdateDefinition`, `UpdateConfig`, `Publish`, `Delete`, `Templates`, `Export`, `MoveToFolder`, `RemoveFromFolder`, `StartSession`, `SendMessage`, `StreamMessage`, `RunText`, `Session`, `SessionEvents` |
| `client.Threads` | `Create`, `Retrieve`, `Messages.Create`, `Messages.List`, `Runs.Create`, `Runs.Stream` |
| `client.Folders` | `List`, `Create`, `Rename`, `Delete` |
| `client.Imports` | `Sources`, `Create`, `Connect`, `Pull`, `Assistant` |
| `client.Calls` | `Create`, `List`, `Retrieve`, `WebToken`, `Control`, `Say`, `Transfer`, `End` |
| `client.PhoneNumbers` | `Search`, `List`, `Providers`, `Provision`, `ImportNumber`, `Instant`, `Assign`, `Release` |
| `client.SipTrunks` | `Presets`, `List`, `Create`, `Update`, `Delete`, `Test` |
| `client.Campaigns` | `List`, `Create`, `Retrieve`, `Start`, `Stop`, `Delete`, `AddContacts`, `SyncContacts`, `PreviewSuppression` |
| `client.Dnc` | `List`, `Add`, `Import`, `SetScope`, `Remove` |
| `client.Integrations` | `List`, `Run`, `Connections.List`, `Connections.Create`, `Connections.Update`, `Connections.Delete`, `Connections.Run` |
| `client.Automations` | `List`, `Create`, `Update`, `Delete`, `Runs`, `Test` |
| `client.TestSuites` | `List`, `Create`, `Delete`, `Run` |
| `client.Sessions` | `List` |
| `client.Webhooks` | `List`, `Create`, `Update`, `Delete`, `Events`, `Deliveries`, `Redeliver`, `Verify` |
| `client.Chat` | `Token`, `Messages`, `Stream`, `UploadFile` |
| `client.Tools` | `List`, `Create`, `Update`, `Delete`, `DiscoverMCP` |
| `client.KnowledgeBase` | `ListDocuments`, `CreateDocument`, `UploadDocument`, `RetrieveDocument`, `DeleteDocument`, `Search` |
| `client.VectorStores` | `List`, `Create`, `Retrieve`, `Delete`, `UploadDocument` |
| `client.Analytics` | `Overview` |
| `client.Billing` | `Credits`, `Transactions`, `Usage` |
| `client.Environments` | `List` |
| `client.Providers` | `List`, `Resources` |
| `client.APIKeys` | `List`, `Create`, `Delete` |
| `client.Organizations` | `Retrieve`, `Update`, `Regions` |

`Agents` and `Workflows` are the same service under two names: the product calls
them agents, the API path is `/workflows`. Either works.

### Text and voice are separate

An agent is one definition. Nothing forces it to do both:

- A **text** agent needs only `Threads` (or `Chat` for a browser widget).
- A **voice** agent adds `Calls`, `PhoneNumbers` and `Campaigns`.
- The same agent can do both, if you want it to.

Any endpoint without a dedicated helper is one call away with `client.Do`:

```go
var out map[string]any
err := client.Do(ctx, "GET", "/analytics/overview", nil, nil, &out)
```

Optional parameters use pointer fields in the params structs. The `glytos.String`,
`glytos.Bool`, `glytos.Int`, and `glytos.Float64` helpers build those pointers:

```go
agents, err := client.Workflows.List(ctx, &glytos.WorkflowListParams{
	Archived:    glytos.Bool(true),
	Environment: "prod",
})
```

## Outbound calling

A campaign dials a list of contacts with one agent. Upload the list as CSV text:
the phone column is found by its header or by which column holds phone numbers,
and every other column travels with that contact, so `{{name}}` in the agent's
prompt means the person being called.

```go
csv, err := os.ReadFile("leads.csv")
if err != nil {
	return err
}
campaign, err := client.Campaigns.Create(ctx, glytos.CampaignCreateParams{
	Name:            "March outreach",
	WorkflowUUID:    agent.UUID,
	FromNumber:      "+15551230000", // must be a number you have connected
	ContactsCSV:     string(csv),
	ScheduledAt:     "2026-03-01T09:00:00Z",
	CallWindowStart: "09:00",
	CallWindowEnd:   "20:00",
	Timezone:        "Europe/Istanbul",
})
```

Left unscheduled, a campaign stays a draft until `Start`. `Stop` ends it at the
next contact, leaving the undialed ones ready to resume. `Retrieve` returns each
contact's outcome and, where one answered, the session it produced.

Every outbound call is checked against your do-not-call list first, whether it
comes from a campaign or from `Calls.Create`. Agents add to that list themselves
when someone asks not to be contacted again:

```go
_, err := client.Dnc.Add(ctx, "+15551230000", "asked on a call")
```

A campaign chooses how much of the list applies. The default, `strict`, honours
all of it. `transactional` still calls people who only refused marketing, which
is what you want for a call about someone's own order. `ignore` skips entries
your organization added for itself, but requests people made on a call still
apply unless you also set `OverrideCallerRequests`. Measure before you choose:

```go
preview, err := client.Campaigns.PreviewSuppression(ctx, nil, string(csv))
fmt.Println(preview.ReachedIfStrict, "of", preview.Contacts,
	"reachable;", preview.CallerRequested, "asked us not to call")
```

## Errors

Non-2xx responses return a `*glytos.Error` carrying the API error `Code`, HTTP
`Status`, `Message`, and the server `RequestID`:

```go
_, err := client.Workflows.Retrieve(ctx, "missing")
var apiErr *glytos.Error
if errors.As(err, &apiErr) {
	fmt.Println(apiErr.Status, apiErr.Code, apiErr.Message)
}
```

## Webhooks

Verify a delivery came from Glytos before trusting it. Pass the **raw** request
body, the `X-Glytos-Signature` header, and your endpoint secret:

```go
ok := glytos.VerifyWebhook(rawBody, r.Header.Get("X-Glytos-Signature"), webhookSecret, glytos.DefaultWebhookTolerance)
if !ok {
	http.Error(w, "invalid signature", http.StatusBadRequest)
	return
}
```

The scheme is HMAC-SHA256 over `"{timestamp}.{body}"`, sent as
`X-Glytos-Signature: t=<ts>,v1=<hex>`. The comparison is constant-time and the
`toleranceSeconds` window guards against replay. `client.Webhooks.Verify(...)` is
the same check as a method.

## License

MIT
