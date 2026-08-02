package glytos

import (
	"context"
	"encoding/json"
	"errors"
)

// ThreadsService holds conversations with a text agent, in the vocabulary the rest
// of the industry uses: a thread holds the conversation, a run is one turn on it.
//
// It is the same session API WorkflowsService exposes, shaped so code written
// against a thread/run model reads the same here.
type ThreadsService struct {
	client *Client
	// Messages adds messages to a thread and lists what has been said.
	Messages *ThreadMessagesService
	// Runs executes one turn, waiting for it or streaming it.
	Runs *ThreadRunsService
}

// ThreadMessagesService adds to and reads a thread's messages.
type ThreadMessagesService struct{ client *Client }

// ThreadRunsService runs one turn on a thread.
type ThreadRunsService struct{ client *Client }

// Thread is a conversation with an agent. It is created against one agent and
// carries its id, so no later call has to repeat it.
type Thread struct {
	// ID is the conversation id.
	ID string `json:"id"`
	// Agent is the agent this conversation belongs to.
	Agent  string `json:"agent"`
	Status string `json:"status,omitempty"`
	// Messages is anything the agent opened with; empty for a silent opening.
	Messages []Message `json:"messages,omitempty"`
}

// Message is one entry of a conversation.
type Message struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	TS      string   `json:"ts,omitempty"`
	NodeID  string   `json:"node_id,omitempty"`
	Images  []string `json:"images,omitempty"`
}

// TurnParams are the optional fields of one turn.
type TurnParams struct {
	// Content is the user message. It may be empty when Instructions carries the turn.
	Content string
	// Images are data: URIs or URLs for this turn only.
	Images []string
	// Instructions is extra context for THIS turn only, applied below the agent's
	// own instructions and never saved to it.
	Instructions string
}

// ErrIncompleteThread is returned when a thread reference is missing an id.
var ErrIncompleteThread = errors.New("glytos: a thread reference needs both ID and Agent")

func (t Thread) ids() (string, string, error) {
	if t.Agent == "" || t.ID == "" {
		return "", "", ErrIncompleteThread
	}
	return t.Agent, t.ID, nil
}

// turnBody is the request body shared by the plain and the streamed endpoint.
func turnBody(params *TurnParams) map[string]any {
	body := map[string]any{"content": ""}
	if params == nil {
		return body
	}
	body["content"] = params.Content
	if params.Images != nil {
		body["images"] = params.Images
	}
	if params.Instructions != "" {
		body["additional_instructions"] = params.Instructions
	}
	return body
}

// Create opens a conversation with an agent. Pass nil for defaults.
func (s *ThreadsService) Create(ctx context.Context, agent string, params *StartSessionParams) (*Thread, error) {
	session, err := (&WorkflowsService{client: s.client}).StartSession(ctx, agent, params)
	if err != nil {
		return nil, err
	}
	return &Thread{ID: session.SessionUUID, Agent: agent, Status: session.Status}, nil
}

// Retrieve returns the conversation so far, with its variables and cost.
func (s *ThreadsService) Retrieve(ctx context.Context, thread Thread) (json.RawMessage, error) {
	agent, id, err := thread.ids()
	if err != nil {
		return nil, err
	}
	var out json.RawMessage
	err = s.client.do(ctx, "GET", "/workflows/"+esc(agent)+"/sessions/"+esc(id), nil, nil, &out)
	return out, err
}

// Create adds a user message and runs the agent on it, returning that turn's reply.
func (s *ThreadMessagesService) Create(ctx context.Context, thread Thread, params *TurnParams) (json.RawMessage, error) {
	agent, id, err := thread.ids()
	if err != nil {
		return nil, err
	}
	var out json.RawMessage
	err = s.client.do(ctx, "POST", "/workflows/"+esc(agent)+"/sessions/"+esc(id)+"/messages", turnBody(params), nil, &out)
	return out, err
}

// List returns every message in the conversation, oldest first.
func (s *ThreadMessagesService) List(ctx context.Context, thread Thread) ([]Message, error) {
	agent, id, err := thread.ids()
	if err != nil {
		return nil, err
	}
	var detail struct {
		Transcript []Message `json:"transcript"`
	}
	if err := s.client.do(ctx, "GET", "/workflows/"+esc(agent)+"/sessions/"+esc(id), nil, nil, &detail); err != nil {
		return nil, err
	}
	return detail.Transcript, nil
}

// Create runs one turn and waits for it. A turn completes before it returns, so
// there is no run to poll: the reply is already in the result.
func (s *ThreadRunsService) Create(ctx context.Context, thread Thread, params *TurnParams) (json.RawMessage, error) {
	return (&ThreadMessagesService{client: s.client}).Create(ctx, thread, params)
}

// Stream runs the same turn, delivering it as it is written. onEvent is called for
// each event in order; returning an error from it stops the stream and is returned.
func (s *ThreadRunsService) Stream(ctx context.Context, thread Thread, params *TurnParams, onEvent func(StreamEvent) error) error {
	agent, id, err := thread.ids()
	if err != nil {
		return err
	}
	return s.client.Stream(ctx, "POST", "/workflows/"+esc(agent)+"/sessions/"+esc(id)+"/messages/stream", turnBody(params), onEvent)
}
