package glytos

import (
	"context"
	"encoding/json"
	"net/url"
)

// WorkflowsService manages agents: prompt agents and visual workflows.
type WorkflowsService struct{ client *Client }

// WorkflowListParams are the optional filters for WorkflowsService.List.
type WorkflowListParams struct {
	// Archived filters by archived state when set.
	Archived *bool
	// Environment scopes the list to an environment ("dev"/"staging"/"prod" or
	// a uuid), or "all" to list across environments.
	Environment string
}

// WorkflowCreateParams are the fields for WorkflowsService.Create.
type WorkflowCreateParams struct {
	// Name is the agent name (required).
	Name string
	// Mode is "prompt" (default) or "workflow".
	Mode string
	// Config is the optional initial agent config.
	Config map[string]any
}

// StartSessionParams are the optional fields for WorkflowsService.StartSession.
type StartSessionParams struct {
	// Variables seeds the session's template variables.
	Variables map[string]any
	// Version pins a specific saved version (an int or a string).
	Version any
}

// List returns your agents. Pass nil for no filters.
func (s *WorkflowsService) List(ctx context.Context, params *WorkflowListParams) ([]Workflow, error) {
	query := url.Values{}
	if params != nil {
		if params.Archived != nil {
			query.Set("archived", boolString(*params.Archived))
		}
		if params.Environment != "" {
			query.Set("environment", params.Environment)
		}
	}
	var out []Workflow
	err := s.client.do(ctx, "GET", "/workflows", nil, query, &out)
	return out, err
}

// Retrieve returns a single agent by uuid.
func (s *WorkflowsService) Retrieve(ctx context.Context, workflowUUID string) (*Workflow, error) {
	var out Workflow
	err := s.client.do(ctx, "GET", "/workflows/"+esc(workflowUUID), nil, nil, &out)
	return &out, err
}

// Create creates an agent.
func (s *WorkflowsService) Create(ctx context.Context, params WorkflowCreateParams) (*Workflow, error) {
	mode := params.Mode
	if mode == "" {
		mode = "prompt"
	}
	body := map[string]any{"name": params.Name, "mode": mode}
	if params.Config != nil {
		body["config"] = params.Config
	}
	var out Workflow
	err := s.client.do(ctx, "POST", "/workflows", body, nil, &out)
	return &out, err
}

// Rename changes an agent's name only (use UpdateConfig/UpdateDefinition for the rest).
func (s *WorkflowsService) Rename(ctx context.Context, workflowUUID, name string) (*Workflow, error) {
	var out Workflow
	err := s.client.do(ctx, "PATCH", "/workflows/"+esc(workflowUUID), map[string]any{"name": name}, nil, &out)
	return &out, err
}

// Export returns an agent as portable, secret-free JSON. It imports back through
// Imports.Create(ctx, "glytos", ...), on this account or another.
func (s *WorkflowsService) Export(ctx context.Context, workflowUUID string) (map[string]any, error) {
	var out map[string]any
	err := s.client.do(ctx, "GET", "/workflows/"+esc(workflowUUID)+"/export", nil, nil, &out)
	return out, err
}

// MoveToFolder files an agent into a folder. Both must be in the same environment.
func (s *WorkflowsService) MoveToFolder(ctx context.Context, workflowUUID, folderUUID string) (*Workflow, error) {
	var out Workflow
	err := s.client.do(ctx, "PATCH", "/workflows/"+esc(workflowUUID), map[string]any{"folder_uuid": folderUUID}, nil, &out)
	return &out, err
}

// RemoveFromFolder takes an agent out of its folder, leaving it ungrouped.
func (s *WorkflowsService) RemoveFromFolder(ctx context.Context, workflowUUID string) (*Workflow, error) {
	var out Workflow
	// Sent as null is what unfiles it; not sent at all would leave it where it is.
	err := s.client.do(ctx, "PATCH", "/workflows/"+esc(workflowUUID), map[string]any{"folder_uuid": nil}, nil, &out)
	return &out, err
}

// Duplicate copies an agent and returns the new copy.
func (s *WorkflowsService) Duplicate(ctx context.Context, workflowUUID string) (*Workflow, error) {
	var out Workflow
	err := s.client.do(ctx, "POST", "/workflows/"+esc(workflowUUID)+"/duplicate", nil, nil, &out)
	return &out, err
}

// Archive hides an agent from the default list.
func (s *WorkflowsService) Archive(ctx context.Context, workflowUUID string) (*Workflow, error) {
	var out Workflow
	err := s.client.do(ctx, "POST", "/workflows/"+esc(workflowUUID)+"/archive", nil, nil, &out)
	return &out, err
}

// Unarchive restores an archived agent.
func (s *WorkflowsService) Unarchive(ctx context.Context, workflowUUID string) (*Workflow, error) {
	var out Workflow
	err := s.client.do(ctx, "POST", "/workflows/"+esc(workflowUUID)+"/unarchive", nil, nil, &out)
	return &out, err
}

// Promote moves an agent into another environment (a move, not a copy).
func (s *WorkflowsService) Promote(ctx context.Context, workflowUUID, targetEnvironmentID string) (*Workflow, error) {
	body := map[string]any{"target_environment_id": targetEnvironmentID}
	var out Workflow
	err := s.client.do(ctx, "POST", "/workflows/"+esc(workflowUUID)+"/promote", body, nil, &out)
	return &out, err
}

// Versions lists the saved versions of an agent.
func (s *WorkflowsService) Versions(ctx context.Context, workflowUUID string) ([]WorkflowVersion, error) {
	var out []WorkflowVersion
	err := s.client.do(ctx, "GET", "/workflows/"+esc(workflowUUID)+"/versions", nil, nil, &out)
	return out, err
}

// UpdateDefinition replaces an agent's graph definition.
func (s *WorkflowsService) UpdateDefinition(ctx context.Context, workflowUUID string, graph map[string]any) (*Workflow, error) {
	var out Workflow
	err := s.client.do(ctx, "PUT", "/workflows/"+esc(workflowUUID)+"/definition", map[string]any{"graph": graph}, nil, &out)
	return &out, err
}

// UpdateConfig replaces an agent's config.
func (s *WorkflowsService) UpdateConfig(ctx context.Context, workflowUUID string, config map[string]any) (*Workflow, error) {
	var out Workflow
	err := s.client.do(ctx, "PUT", "/workflows/"+esc(workflowUUID)+"/config", map[string]any{"config": config}, nil, &out)
	return &out, err
}

// Publish publishes the current draft so the agent goes live.
func (s *WorkflowsService) Publish(ctx context.Context, workflowUUID string) (*Workflow, error) {
	var out Workflow
	err := s.client.do(ctx, "POST", "/workflows/"+esc(workflowUUID)+"/publish", nil, nil, &out)
	return &out, err
}

// Delete deletes an agent.
func (s *WorkflowsService) Delete(ctx context.Context, workflowUUID string) error {
	return s.client.do(ctx, "DELETE", "/workflows/"+esc(workflowUUID), nil, nil, nil)
}

// Templates returns ready-made starter workflow graphs.
func (s *WorkflowsService) Templates(ctx context.Context) ([]Workflow, error) {
	var out []Workflow
	err := s.client.do(ctx, "GET", "/workflows/templates", nil, nil, &out)
	return out, err
}

// StartSession starts a text/chat session against an agent. Pass nil for defaults.
func (s *WorkflowsService) StartSession(ctx context.Context, workflowUUID string, params *StartSessionParams) (*Session, error) {
	body := map[string]any{}
	if params != nil {
		if params.Variables != nil {
			body["variables"] = params.Variables
		}
		if params.Version != nil {
			body["version"] = params.Version
		}
	}
	var out Session
	err := s.client.do(ctx, "POST", "/workflows/"+esc(workflowUUID)+"/sessions", body, nil, &out)
	return &out, err
}

// SendMessage sends one user message to an existing session and returns that
// turn's reply. Pass nil images to omit them.
func (s *WorkflowsService) SendMessage(ctx context.Context, workflowUUID, sessionUUID, content string, images []string) (json.RawMessage, error) {
	return s.SendTurn(ctx, workflowUUID, sessionUUID, &TurnParams{Content: content, Images: images})
}

// SendTurn is SendMessage with the full turn, including per-turn Instructions.
func (s *WorkflowsService) SendTurn(ctx context.Context, workflowUUID, sessionUUID string, params *TurnParams) (json.RawMessage, error) {
	body := turnBody(params)
	var out json.RawMessage
	err := s.client.do(ctx, "POST", "/workflows/"+esc(workflowUUID)+"/sessions/"+esc(sessionUUID)+"/messages", body, nil, &out)
	return out, err
}

// StreamMessage runs the same turn, delivering it as it is written.
func (s *WorkflowsService) StreamMessage(ctx context.Context, workflowUUID, sessionUUID string, params *TurnParams, onEvent func(StreamEvent) error) error {
	return s.client.Stream(ctx, "POST", "/workflows/"+esc(workflowUUID)+"/sessions/"+esc(sessionUUID)+"/messages/stream", turnBody(params), onEvent)
}

// RunText runs a one-shot text conversation (a list of {"role","content"}
// messages) against an agent.
func (s *WorkflowsService) RunText(ctx context.Context, workflowUUID string, messages []map[string]any) (json.RawMessage, error) {
	var out json.RawMessage
	err := s.client.do(ctx, "POST", "/workflows/"+esc(workflowUUID)+"/runs/text", map[string]any{"messages": messages}, nil, &out)
	return out, err
}

// Session returns full detail for one session of an agent (transcript, cost,
// latency, and so on).
func (s *WorkflowsService) Session(ctx context.Context, workflowUUID, sessionUUID string) (*Session, error) {
	var out Session
	err := s.client.do(ctx, "GET", "/workflows/"+esc(workflowUUID)+"/sessions/"+esc(sessionUUID), nil, nil, &out)
	return &out, err
}

// SessionEvents returns the run-event log for a session (routing decisions,
// tool calls, and so on).
func (s *WorkflowsService) SessionEvents(ctx context.Context, workflowUUID, sessionUUID string) ([]map[string]any, error) {
	var out []map[string]any
	err := s.client.do(ctx, "GET", "/workflows/"+esc(workflowUUID)+"/sessions/"+esc(sessionUUID)+"/events", nil, nil, &out)
	return out, err
}

func boolString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
