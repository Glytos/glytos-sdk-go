# glytos-sdk-go

[![CI](https://github.com/Glytos/glytos-sdk-go/actions/workflows/ci.yml/badge.svg)](https://github.com/Glytos/glytos-sdk-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/Glytos/glytos-sdk-go.svg)](https://pkg.go.dev/github.com/Glytos/glytos-sdk-go)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

The official [Glytos](https://glytos.com) server SDK for Go.

Call the Glytos API from your backend with an API key: build and run voice agents,
start phone calls, mint browser web-call tokens, manage phone numbers, and verify
webhooks. Zero dependencies (standard library only), fully typed, context-aware.

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

## Resources

| Namespace | Methods |
| --- | --- |
| `client.Workflows` | `List`, `Retrieve`, `Create`, `Rename`, `Duplicate`, `Archive`, `Unarchive`, `Promote`, `Versions`, `UpdateDefinition`, `UpdateConfig`, `Publish`, `Delete`, `Templates`, `StartSession`, `SendMessage`, `RunText`, `Session`, `SessionEvents` |
| `client.Calls` | `Create`, `List`, `Retrieve`, `WebToken`, `Control` |
| `client.PhoneNumbers` | `Search`, `List`, `Providers`, `Provision`, `ImportNumber`, `Instant`, `Assign`, `Release` |
| `client.Campaigns` | `List`, `Create`, `Retrieve`, `Start`, `SyncContacts` |
| `client.Sessions` | `List` |
| `client.Webhooks` | `List`, `Create`, `Update`, `Delete`, `Events`, `Deliveries`, `Redeliver`, `Verify` |
| `client.Chat` | `Token`, `Messages` |
| `client.Tools` | `List`, `Create`, `Update`, `Delete` |
| `client.KnowledgeBase` | `ListDocuments`, `CreateDocument`, `Search` |
| `client.VectorStores` | `List`, `Create`, `Retrieve`, `Delete` |
| `client.Analytics` | `Overview` |

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
