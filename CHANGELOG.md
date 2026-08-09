# Changelog

All notable changes to this project are documented in this file. The format is based
on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
