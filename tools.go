package glytos

import "context"

// ToolsService manages reusable tools an agent can call.
//
// Kind is one of "static", "http", "mcp", "code", "integration" or "client". An
// integration tool names its connection in Config, so the model fills in
// arguments but never chooses the destination. A code tool runs only in an
// operator-configured sandbox, and a client tool is resolved by the browser
// during a web call, so both are inert unless that side is set up.
type ToolsService struct{ client *Client }

// ToolCreateParams are the fields for ToolsService.Create.
type ToolCreateParams struct {
	// Name is the tool name (required).
	Name string
	// Kind is one of static, http, mcp, code, integration, client (required).
	Kind string
	// Description is an optional human description.
	Description string
	// Config is the kind-specific configuration.
	Config map[string]any
	// Parameters is the JSON-schema parameter definition.
	Parameters map[string]any
	// RunInBackground lets a voice agent keep talking while this tool runs and
	// share the result when it lands, instead of holding the caller in silence.
	// Allowed only for the server-side kinds (http, mcp, integration, code), and
	// takes effect only where the operator has enabled background tools.
	RunInBackground bool
}

// ToolUpdateParams are the optional fields for ToolsService.Update. Only the
// fields you set are changed.
type ToolUpdateParams struct {
	Name        string
	Description string
	Kind        string
	Config      map[string]any
	Parameters  map[string]any
	// RunInBackground is a pointer because false is a real value here: nil means
	// "leave it as it is", &false turns it off.
	RunInBackground *bool
}

// List returns your saved tools.
func (s *ToolsService) List(ctx context.Context) ([]Tool, error) {
	var out []Tool
	err := s.client.do(ctx, "GET", "/tools", nil, nil, &out)
	return out, err
}

// Create creates a tool.
func (s *ToolsService) Create(ctx context.Context, params ToolCreateParams) (*Tool, error) {
	body := map[string]any{"name": params.Name, "kind": params.Kind}
	if params.Description != "" {
		body["description"] = params.Description
	}
	if params.Config != nil {
		body["config"] = params.Config
	}
	if params.Parameters != nil {
		body["parameters"] = params.Parameters
	}
	if params.RunInBackground {
		body["run_in_background"] = true
	}
	var out Tool
	err := s.client.do(ctx, "POST", "/tools", body, nil, &out)
	return &out, err
}

// Update updates a tool. Only the fields you set are changed.
func (s *ToolsService) Update(ctx context.Context, toolUUID string, params ToolUpdateParams) (*Tool, error) {
	body := map[string]any{}
	if params.Name != "" {
		body["name"] = params.Name
	}
	if params.Description != "" {
		body["description"] = params.Description
	}
	if params.Kind != "" {
		body["kind"] = params.Kind
	}
	if params.Config != nil {
		body["config"] = params.Config
	}
	if params.Parameters != nil {
		body["parameters"] = params.Parameters
	}
	if params.RunInBackground != nil {
		body["run_in_background"] = *params.RunInBackground
	}
	var out Tool
	err := s.client.do(ctx, "PATCH", "/tools/"+esc(toolUUID), body, nil, &out)
	return &out, err
}

// Delete deletes a tool.
func (s *ToolsService) Delete(ctx context.Context, toolUUID string) error {
	return s.client.do(ctx, "DELETE", "/tools/"+esc(toolUUID), nil, nil, nil)
}

// DiscoverMCP asks an MCP server what it publishes, so a tool can be built from
// the server's own schema rather than one transcribed by hand. headers may be
// nil. It returns the tool list itself, not the response envelope.
func (s *ToolsService) DiscoverMCP(ctx context.Context, serverURL string, headers map[string]string) ([]McpTool, error) {
	body := map[string]any{"server_url": serverURL}
	if headers != nil {
		body["headers"] = headers
	}
	var out struct {
		Tools []McpTool `json:"tools"`
	}
	err := s.client.do(ctx, "POST", "/tools/mcp/discover", body, nil, &out)
	return out.Tools, err
}
