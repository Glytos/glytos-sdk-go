package glytos

import "context"

// FoldersService manages folders that group agents inside an environment.
type FoldersService struct{ client *Client }

// AgentFolder groups agents within one environment.
type AgentFolder struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

// List returns the folders in the active environment.
func (s *FoldersService) List(ctx context.Context) ([]AgentFolder, error) {
	var out []AgentFolder
	err := s.client.do(ctx, "GET", "/agent-folders", nil, nil, &out)
	return out, err
}

// Create adds a folder to the active environment.
func (s *FoldersService) Create(ctx context.Context, name string) (*AgentFolder, error) {
	var out AgentFolder
	err := s.client.do(ctx, "POST", "/agent-folders", map[string]any{"name": name}, nil, &out)
	return &out, err
}

// Rename renames a folder.
func (s *FoldersService) Rename(ctx context.Context, folderUUID, name string) (*AgentFolder, error) {
	var out AgentFolder
	err := s.client.do(ctx, "PATCH", "/agent-folders/"+esc(folderUUID), map[string]any{"name": name}, nil, &out)
	return &out, err
}

// Delete removes a folder. The agents filed in it are deleted with it.
func (s *FoldersService) Delete(ctx context.Context, folderUUID string) error {
	return s.client.do(ctx, "DELETE", "/agent-folders/"+esc(folderUUID), nil, nil, nil)
}

// ImportsService brings an agent over from another platform.
type ImportsService struct{ client *Client }

// Sources lists the platforms an agent can be brought over from.
func (s *ImportsService) Sources(ctx context.Context) ([]map[string]any, error) {
	var out []map[string]any
	err := s.client.do(ctx, "GET", "/imports/sources", nil, nil, &out)
	return out, err
}

// Create brings an agent over from another platform's export.
func (s *ImportsService) Create(ctx context.Context, source string, payload map[string]any) (map[string]any, error) {
	var out map[string]any
	err := s.client.do(ctx, "POST", "/imports/"+esc(source), map[string]any{"payload": payload}, nil, &out)
	return out, err
}

// Assistant brings over an assistant definition, tools and all.
func (s *ImportsService) Assistant(ctx context.Context, assistant map[string]any) (map[string]any, error) {
	var out map[string]any
	err := s.client.do(ctx, "POST", "/imports/openai-assistant", map[string]any{"assistant": assistant}, nil, &out)
	return out, err
}
