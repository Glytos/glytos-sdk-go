package glytos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// sseBody frames events exactly as the server writes them.
func sseBody(blocks ...[2]string) string {
	var b strings.Builder
	for _, block := range blocks {
		fmt.Fprintf(&b, "event: %s\ndata: %s\n\n", block[0], block[1])
	}
	return b.String()
}

func TestThreadCreateCarriesTheAgentID(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{"session_uuid":"ses_1","status":"in_progress"}`))
	}))
	defer server.Close()

	client := New("gly_test", WithBaseURL(server.URL))
	thread, err := client.Threads.Create(context.Background(), "wf_1", &StartSessionParams{
		Variables: map[string]any{"name": "Ada"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !strings.HasSuffix(gotPath, "/workflows/wf_1/sessions") {
		t.Fatalf("path = %q", gotPath)
	}
	// The agent id rides on the thread so no later call has to repeat it.
	if thread.ID != "ses_1" || thread.Agent != "wf_1" {
		t.Fatalf("thread = %+v", thread)
	}
}

func TestTurnSendsPerTurnInstructions(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := New("gly_test", WithBaseURL(server.URL))
	_, err := client.Threads.Messages.Create(context.Background(),
		Thread{ID: "ses_1", Agent: "wf_1"},
		&TurnParams{Content: "hello", Instructions: "answer in French"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if gotBody["content"] != "hello" || gotBody["additional_instructions"] != "answer in French" {
		t.Fatalf("body = %+v", gotBody)
	}
}

func TestAnIncompleteThreadIsRefused(t *testing.T) {
	client := New("gly_test")
	_, err := client.Threads.Messages.Create(context.Background(), Thread{Agent: "wf_1"}, nil)
	if !errors.Is(err, ErrIncompleteThread) {
		t.Fatalf("err = %v, want ErrIncompleteThread", err)
	}
}

func TestStreamYieldsTokensThenTheFinishedRun(t *testing.T) {
	body := sseBody(
		[2]string{"token", `{"delta":"He"}`},
		[2]string{"token", `{"delta":"llo"}`},
		[2]string{"done", `{"status":"completed"}`},
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	client := New("gly_test", WithBaseURL(server.URL))
	var text strings.Builder
	var last StreamEvent
	err := client.Threads.Runs.Stream(context.Background(), Thread{ID: "s", Agent: "w"}, nil,
		func(event StreamEvent) error {
			if event.Type == "token" {
				text.WriteString(event.Delta)
			}
			last = event
			return nil
		})
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	if text.String() != "Hello" {
		t.Fatalf("text = %q", text.String())
	}
	if last.Type != "done" || !strings.Contains(string(last.Run), "completed") {
		t.Fatalf("last = %+v", last)
	}
}

func TestStreamEmitsAFinalEventWithoutATrailingBlankLine(t *testing.T) {
	// The last block has no trailing blank line; it must still be delivered.
	body := "event: token\ndata: {\"delta\":\"x\"}\n\nevent: done\ndata: {\"status\":\"completed\"}"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	var kinds []string
	err := New("gly_test", WithBaseURL(server.URL)).Threads.Runs.Stream(
		context.Background(), Thread{ID: "s", Agent: "w"}, nil,
		func(event StreamEvent) error { kinds = append(kinds, event.Type); return nil })
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	if len(kinds) != 2 || kinds[0] != "token" || kinds[1] != "done" {
		t.Fatalf("kinds = %v", kinds)
	}
}

func TestStreamReturnsTheAPIErrorOnRejection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(402)
		_, _ = w.Write([]byte(`{"error":{"code":"insufficient_credit","message":"no credit"}}`))
	}))
	defer server.Close()

	err := New("gly_test", WithBaseURL(server.URL)).Threads.Runs.Stream(
		context.Background(), Thread{ID: "s", Agent: "w"}, nil,
		func(StreamEvent) error { return nil })
	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr.Status != 402 || apiErr.Code != "insufficient_credit" {
		t.Fatalf("err = %v", err)
	}
}

func TestAnErrorFromTheCallbackStopsTheStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(sseBody(
			[2]string{"token", `{"delta":"a"}`},
			[2]string{"token", `{"delta":"b"}`},
		)))
	}))
	defer server.Close()

	stop := errors.New("stop")
	seen := 0
	err := New("gly_test", WithBaseURL(server.URL)).Threads.Runs.Stream(
		context.Background(), Thread{ID: "s", Agent: "w"}, nil,
		func(StreamEvent) error { seen++; return stop })
	if !errors.Is(err, stop) || seen != 1 {
		t.Fatalf("err = %v, seen = %d", err, seen)
	}
}

func TestFoldersAndImports(t *testing.T) {
	var paths []string
	var bodies []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		bodies = append(bodies, body)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := New("gly_test", WithBaseURL(server.URL))
	ctx := context.Background()
	if _, err := client.Folders.Create(ctx, "Sales"); err != nil {
		t.Fatalf("folder create: %v", err)
	}
	if err := client.Folders.Delete(ctx, "fld_1"); err != nil {
		t.Fatalf("folder delete: %v", err)
	}
	if _, err := client.Imports.Assistant(ctx, map[string]any{"name": "Support"}); err != nil {
		t.Fatalf("import: %v", err)
	}

	if !strings.HasSuffix(paths[0], "/agent-folders") || bodies[0]["name"] != "Sales" {
		t.Fatalf("create = %q %+v", paths[0], bodies[0])
	}
	if !strings.HasPrefix(paths[1], "DELETE") || !strings.HasSuffix(paths[1], "/agent-folders/fld_1") {
		t.Fatalf("delete = %q", paths[1])
	}
}

func TestAgentExportAndFolderFiling(t *testing.T) {
	var paths []string
	var raw []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		buf := new(strings.Builder)
		_, _ = io.Copy(buf, r.Body)
		raw = append(raw, buf.String())
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := New("gly_test", WithBaseURL(server.URL))
	ctx := context.Background()
	if _, err := client.Agents.Export(ctx, "wf_1"); err != nil {
		t.Fatalf("export: %v", err)
	}
	if _, err := client.Agents.MoveToFolder(ctx, "wf_1", "fld_1"); err != nil {
		t.Fatalf("move: %v", err)
	}
	if _, err := client.Agents.RemoveFromFolder(ctx, "wf_1"); err != nil {
		t.Fatalf("unfile: %v", err)
	}

	if !strings.HasPrefix(paths[0], "GET") || !strings.HasSuffix(paths[0], "/workflows/wf_1/export") {
		t.Fatalf("export path = %q", paths[0])
	}
	if !strings.Contains(raw[1], `"folder_uuid":"fld_1"`) {
		t.Fatalf("move body = %q", raw[1])
	}
	// Sent as null is what unfiles an agent; not sent would leave it where it is.
	if !strings.Contains(raw[2], `"folder_uuid":null`) {
		t.Fatalf("unfile body = %q", raw[2])
	}
}

func TestUploadIsMultipartNotJSON(t *testing.T) {
	var contentType string
	var sawFilename bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		if err := r.ParseMultipartForm(1 << 20); err == nil {
			_, header, err := r.FormFile("file")
			sawFilename = err == nil && header.Filename == "notes.txt"
		}
		_, _ = w.Write([]byte(`{"file_uuid":"f_1"}`))
	}))
	defer server.Close()

	client := New("gly_test", WithBaseURL(server.URL))
	if _, err := client.Chat.UploadFile(context.Background(), "tok", "ses_1", "notes.txt", []byte("hi")); err != nil {
		t.Fatalf("upload: %v", err)
	}
	if !strings.HasPrefix(contentType, "multipart/form-data") || !strings.Contains(contentType, "boundary=") {
		t.Fatalf("content-type = %q", contentType)
	}
	if !sawFilename {
		t.Fatal("the file part did not arrive with its filename")
	}
}

func TestAgentsIsTheSameServiceAsWorkflows(t *testing.T) {
	client := New("gly_test")
	if client.Agents != client.Workflows {
		t.Fatal("Agents and Workflows should be the same service")
	}
}
