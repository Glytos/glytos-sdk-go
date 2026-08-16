# Changelog

All notable changes to this project are documented in this file. The format is based
on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `SipTrunks` - connect a carrier directly over SIP, with no third party in
  between: `Presets`, `List`, `Create`, `Update`, `Delete`, `Test`. Numbers are
  attached to a registered trunk through `PhoneNumbers.ImportNumber`, whose
  params gained `SipTrunkUUID`.
- `Integrations` and `Integrations.Connections` - the destinations an agent or an
  automation can act on, and the named connections holding their credentials.
- `Automations` - fire an integration action when an event happens: `List`,
  `Create`, `Update`, `Delete`, `Runs`, `Test`.
- `TestSuites` - `List`, `Create`, `Delete`, `Run`.
- `Billing` - `Credits`, `Transactions`, `Usage`. Checking the balance before a
  long outbound run no longer needs a raw `Do` call.
- `Environments.List`, `Providers.List`, `Providers.Resources`, `APIKeys.List`
  /`Create`/`Delete`, `Organizations.Retrieve`/`Update`/`Regions`.
- `KnowledgeBase.RetrieveDocument` and `KnowledgeBase.DeleteDocument`. Documents
  could be created and listed but never read back or removed.
- `Tools.DiscoverMCP` - ask an MCP server what it publishes, instead of
  transcribing its schema by hand.
- `Imports.Connect` and `Imports.Pull` - list the agents on another platform with
  its API key, then bring over the ones you pick. The key is never stored.
- `Calls.Say`, `Calls.Transfer` and `Calls.End`, which spell out what each
  control action requires. `Calls.Control` still takes a raw map.
- `WorkflowCreateParams.PrimaryChannel`.

### Fixed

- `ToolsService` documented `kind` as http / static / mcp. The API has accepted
  `code`, `integration` and `client` since they shipped, and the doc comment now
  says what each of the six does.

## [0.3.0] - 2026-08-09

### Added

- `Dnc` - the numbers your organization must not call: `Dnc.List`, `Dnc.Add`,
  `Dnc.Import`, `Dnc.SetScope`, `Dnc.Remove`. Every outbound call is checked
  against this list, whether it comes from a campaign or from `Calls.Create`.
- `Campaigns.Stop`, `Campaigns.Delete` and `Campaigns.AddContacts` (upload a
  contact list as CSV text rather than serving it over HTTP).
- `Campaigns.PreviewSuppression` - how many of a contact list each suppression
  policy would reach, including how many of those people asked on a call not to
  be contacted again.
- `CampaignCreateParams` gained `ContactsCSV`, `ScheduledAt`, `CallWindowStart`
  /`CallWindowEnd`, `Timezone`, `SuppressionPolicy` and `OverrideCallerRequests`.
- `CampaignDetail`, `CampaignContact`, `SuppressionPreview`, `ContactSyncResult`
  and `DncEntry` types. `Campaign` gained its scheduling, calling-window and
  suppression fields.

### Changed

- `Campaigns.Retrieve` returns `*CampaignDetail`, and `Campaigns.SyncContacts`
  returns `*ContactSyncResult`, rather than the untyped shapes they had before.

### Fixed

- `CampaignCreateParams.Contacts` was `[]map[string]any`, which the API rejects
  with a 422. It is a `[]string` of phone numbers.

## [0.2.0] - 2026-08-02

### Added

- `Threads` - text conversations: `Threads.Create/Retrieve`,
  `Threads.Messages.Create/List`, `Threads.Runs.Create/Stream`.
- Streaming via a callback: `Threads.Runs.Stream`, `Workflows.StreamMessage` and
  `Chat.Stream` deliver `token` deltas and a terminal `done`.
- `TurnParams.Instructions` - extra context for one turn only.
- File uploads: `Chat.UploadFile`, `KnowledgeBase.UploadDocument`,
  `VectorStores.UploadDocument`, plus `Client.UploadFile` for any other endpoint.
- `Folders` and `Imports` services, plus `Agents.MoveToFolder` /
  `Agents.RemoveFromFolder` to file an agent and `Agents.Export` for the
  portable, secret-free JSON that imports back.
- `Agents` as an alias of `Workflows`.

## [0.1.0] - 2026-07-25

### Added

- Initial release.
- `Client` with `Workflows`, `Calls`, `PhoneNumbers`, `Campaigns`, `Sessions`,
  `Webhooks`, `Chat`, `Tools`, `KnowledgeBase`, `VectorStores` and `Analytics`
  resources, plus a generic `Do()` for any other endpoint.
- Functional options: `WithBaseURL`, `WithEnvironment`, `WithHTTPClient`.
- Context-aware methods and a typed `*Error` carrying HTTP status, API code,
  message and request id.
- `VerifyWebhook()` for constant-time webhook signature verification.
- Zero external dependencies (standard library only).
