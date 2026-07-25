package glytos

import "context"

// ToolsService manages reusable tools an agent can call (kind = http / static / mcp).
type ToolsService struct{ client *Client }

// ToolCreateParams are the fields for ToolsService.Create.
type ToolCreateParams struct {
	// Name is the tool name (required).
	Name string
	// Kind is "http", "static", or "mcp" (required).
	Kind string
	// Description is an optional human description.
	Description string
	// Config is the kind-specific configuration.
	Config map[string]any
	// Parameters is the JSON-schema parameter definition.
	Parameters map[string]any
}

// ToolUpdateParams are the optional fields for ToolsService.Update. Only the
// fields you set are changed.
type ToolUpdateParams struct {
	Name        string
	Description string
	Kind        string
	Config      map[string]any
	Parameters  map[string]any
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
	var out Tool
	err := s.client.do(ctx, "PATCH", "/tools/"+esc(toolUUID), body, nil, &out)
	return &out, err
}

// Delete deletes a tool.
func (s *ToolsService) Delete(ctx context.Context, toolUUID string) error {
	return s.client.do(ctx, "DELETE", "/tools/"+esc(toolUUID), nil, nil, nil)
}
