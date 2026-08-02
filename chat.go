package glytos

import (
	"context"
	"encoding/json"
)

// ChatService mints embeddable chat tokens and exchanges messages with them.
type ChatService struct{ client *Client }

// ChatMessagesParams are the fields for ChatService.Messages. Token and Content
// are required.
type ChatMessagesParams struct {
	// Token is the chat token from ChatService.Token (authenticates the turn).
	Token string
	// Content is the user message text.
	Content string
	// SessionUUID continues an existing chat session.
	SessionUUID string
	// Images are optional data: URIs or URLs attached to the turn.
	Images []string
}

// Token mints a short-lived chat token scoped to a workflow.
func (s *ChatService) Token(ctx context.Context, workflowUUID string) (*ChatToken, error) {
	var out ChatToken
	err := s.client.do(ctx, "POST", "/chat/token", map[string]any{"workflow_uuid": workflowUUID}, nil, &out)
	return &out, err
}

// Messages sends a chat turn. It is authenticated by the body token from Token,
// not the API key.
func (s *ChatService) Messages(ctx context.Context, params ChatMessagesParams) (json.RawMessage, error) {
	body := map[string]any{"token": params.Token, "content": params.Content}
	if params.SessionUUID != "" {
		body["session_uuid"] = params.SessionUUID
	}
	if params.Images != nil {
		body["images"] = params.Images
	}
	var out json.RawMessage
	err := s.client.do(ctx, "POST", "/chat/messages", body, nil, &out)
	return out, err
}

// Stream sends the same turn, delivering it as it is written. onEvent is called
// for each event in order; returning an error from it stops the stream.
func (s *ChatService) Stream(ctx context.Context, params ChatMessagesParams, onEvent func(StreamEvent) error) error {
	body := map[string]any{"token": params.Token, "content": params.Content}
	if params.SessionUUID != "" {
		body["session_uuid"] = params.SessionUUID
	}
	if params.Images != nil {
		body["images"] = params.Images
	}
	return s.client.Stream(ctx, "POST", "/chat/stream", body, onEvent)
}

// ChatFile is a file attached to one conversation.
type ChatFile struct {
	FileUUID   string `json:"file_uuid"`
	Filename   string `json:"filename"`
	Characters int    `json:"characters"`
}

// UploadFile attaches a file to one conversation. Its text is put in front of the
// agent for that conversation only - it does not join the knowledge base.
func (s *ChatService) UploadFile(ctx context.Context, token, sessionUUID, filename string, content []byte) (*ChatFile, error) {
	var out ChatFile
	err := s.client.UploadFile(ctx, "/chat/files",
		map[string]string{"token": token, "session_uuid": sessionUUID}, filename, content, &out)
	return &out, err
}
