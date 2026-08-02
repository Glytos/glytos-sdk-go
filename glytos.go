// Package glytos is the official Glytos server SDK for Go.
//
// Call the Glytos API from your backend with an API key: build and run voice
// agents, start phone calls, mint browser web-call tokens, manage phone numbers,
// and verify webhooks. It depends only on the standard library.
//
// Never ship an API key to the browser. For in-browser voice, use the
// "@glytos/web" package with a short-lived token you mint here via
// Client.Calls.WebToken.
//
//	client := glytos.New("gly_...")
//
//	agents, err := client.Workflows.List(context.Background(), nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	token, err := client.Calls.WebToken(context.Background(), glytos.WebTokenParams{
//		WorkflowUUID: agents[0].UUID,
//	})
//
// Every resource method is a thin, typed wrapper over the REST API. For any
// endpoint without a dedicated helper, use Client.Do.
package glytos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultBaseURL is the public Glytos API base. Override it with WithBaseURL,
// for example to target a regional stack.
const DefaultBaseURL = "https://api.glytos.com/api/v1"

const defaultTimeout = 30 * time.Second

// Error is returned for any non-2xx API response. It carries the HTTP status,
// the API error code and message from the {"error":{...}} envelope, and the
// server request id (when present) for support.
type Error struct {
	// Status is the HTTP status code.
	Status int
	// Code is the machine-readable API error code (e.g. "not_found").
	Code string
	// Message is the human-readable error message.
	Message string
	// RequestID is the X-Request-Id response header, if the server sent one.
	RequestID string
}

func (e *Error) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf(
			"glytos: %s (status %d, code %q, request %s)",
			e.Message, e.Status, e.Code, e.RequestID,
		)
	}
	return fmt.Sprintf("glytos: %s (status %d, code %q)", e.Message, e.Status, e.Code)
}

// Client is a Glytos API client. Construct it with New. It is safe for
// concurrent use by multiple goroutines.
type Client struct {
	apiKey      string
	baseURL     string
	environment string
	httpClient  *http.Client

	// Workflows manages agents (prompt agents and visual workflows).
	Workflows *WorkflowsService
	// Agents is the same service as Workflows, under the word the product uses.
	Agents *WorkflowsService
	// Threads holds text conversations: a thread per conversation, a run per turn.
	Threads *ThreadsService
	// Folders groups agents inside an environment.
	Folders *FoldersService
	// Imports brings an agent over from another platform.
	Imports *ImportsService
	// Calls manages phone and web calls.
	Calls *CallsService
	// PhoneNumbers manages telephony numbers and providers.
	PhoneNumbers *PhoneNumbersService
	// Campaigns manages outbound calling campaigns.
	Campaigns *CampaignsService
	// Sessions lists sessions across agents.
	Sessions *SessionsService
	// Webhooks manages webhook endpoints, deliveries, and signature verification.
	Webhooks *WebhooksService
	// Chat mints widget chat tokens and exchanges messages.
	Chat *ChatService
	// Tools manages reusable agent tools.
	Tools *ToolsService
	// KnowledgeBase manages knowledge-base documents and search.
	KnowledgeBase *KnowledgeBaseService
	// VectorStores manages vector stores over knowledge-base documents.
	VectorStores *VectorStoresService
	// Analytics exposes usage and activity analytics.
	Analytics *AnalyticsService
}

// Option configures a Client. Pass options to New.
type Option func(*Client)

// WithBaseURL overrides the API base URL (for example a regional stack).
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		if baseURL != "" {
			c.baseURL = strings.TrimRight(baseURL, "/")
		}
	}
}

// WithEnvironment sets the environment to act in: "dev", "staging", "prod", or
// an environment uuid. It defaults to the organization's default environment
// (Development). Agents are still created in Development regardless; this scopes
// reads and calls. Sent as the X-Environment-Id header.
func WithEnvironment(environment string) Option {
	return func(c *Client) { c.environment = environment }
}

// WithHTTPClient sets a custom *http.Client (for example to configure timeouts,
// proxies, or transport-level instrumentation).
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// New creates a Glytos client. apiKey is your organization API key (it starts
// with "gly_"). Configure it with functional options.
//
//	client := glytos.New("gly_...", glytos.WithEnvironment("prod"))
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:     apiKey,
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}

	c.Workflows = &WorkflowsService{client: c}
	c.Agents = c.Workflows
	c.Threads = &ThreadsService{
		client:   c,
		Messages: &ThreadMessagesService{client: c},
		Runs:     &ThreadRunsService{client: c},
	}
	c.Folders = &FoldersService{client: c}
	c.Imports = &ImportsService{client: c}
	c.Calls = &CallsService{client: c}
	c.PhoneNumbers = &PhoneNumbersService{client: c}
	c.Campaigns = &CampaignsService{client: c}
	c.Sessions = &SessionsService{client: c}
	c.Webhooks = &WebhooksService{client: c}
	c.Chat = &ChatService{client: c}
	c.Tools = &ToolsService{client: c}
	c.KnowledgeBase = &KnowledgeBaseService{client: c}
	c.VectorStores = &VectorStoresService{client: c}
	c.Analytics = &AnalyticsService{client: c}
	return c
}

// Do performs a request against any API endpoint. path is relative to the API
// base (for example "/workflows"). body, when non-nil, is JSON-encoded; query,
// when non-empty, is appended (empty values are dropped); out, when non-nil, is
// a pointer the JSON response is decoded into. A non-2xx response returns *Error.
//
// Use it for endpoints without a dedicated helper.
func (c *Client) Do(
	ctx context.Context,
	method, path string,
	body any,
	query url.Values,
	out any,
) error {
	return c.do(ctx, method, path, body, query, out)
}

func (c *Client) do(
	ctx context.Context,
	method, path string,
	body any,
	query url.Values,
	out any,
) error {
	target := c.baseURL + path
	if q := encodeQuery(query); q != "" {
		target += "?" + q
	}

	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("glytos: encode request body: %w", err)
		}
		reader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, target, reader)
	if err != nil {
		return fmt.Errorf("glytos: build request: %w", err)
	}
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	if c.environment != "" {
		req.Header.Set("X-Environment-Id", c.environment)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("glytos: request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("glytos: read response: %w", err)
	}
	requestID := resp.Header.Get("X-Request-Id")

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseError(resp.StatusCode, requestID, data)
	}

	if out == nil || len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("glytos: decode response: %w", err)
	}
	return nil
}

// encodeQuery renders query values, dropping any with an empty string value so
// callers can pass optionals freely. It returns "" when nothing remains.
func encodeQuery(query url.Values) string {
	if len(query) == 0 {
		return ""
	}
	filtered := url.Values{}
	for key, values := range query {
		for _, v := range values {
			if v != "" {
				filtered.Add(key, v)
			}
		}
	}
	return filtered.Encode()
}

// parseError maps a non-2xx response to a typed *Error, reading the API
// {"error":{"code","message"}} envelope when present.
func parseError(status int, requestID string, data []byte) *Error {
	e := &Error{Status: status, Code: "error", Message: http.StatusText(status), RequestID: requestID}
	if e.Message == "" {
		e.Message = "request failed"
	}
	var envelope struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(data, &envelope) == nil {
		if envelope.Error.Code != "" {
			e.Code = envelope.Error.Code
		}
		if envelope.Error.Message != "" {
			e.Message = envelope.Error.Message
		}
	}
	return e
}

// String returns a pointer to v, for optional string fields in params structs.
func String(v string) *string { return &v }

// Bool returns a pointer to v, for optional bool fields in params structs.
func Bool(v bool) *bool { return &v }

// Int returns a pointer to v, for optional int fields in params structs.
func Int(v int) *int { return &v }

// Float64 returns a pointer to v, for optional float64 fields in params structs.
func Float64(v float64) *float64 { return &v }
