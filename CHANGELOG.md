# Changelog

All notable changes to this project are documented in this file. The format is based
on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
