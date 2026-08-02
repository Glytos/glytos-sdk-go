package glytos

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

// StreamEvent is one Server-Sent Event from a streamed turn.
//
// Type is "token" (Delta carries the piece), "done" (Run carries the finished
// turn, the same payload the non-streamed call returns) or "error".
type StreamEvent struct {
	Type    string
	Delta   string
	Run     json.RawMessage
	Message string
}

// Stream calls a Server-Sent Events endpoint and invokes onEvent for each event
// in order. The reply arrives as it is written rather than after the last token,
// which is the whole difference on a long answer.
//
// Returning an error from onEvent stops reading and returns that error.
func (c *Client) Stream(
	ctx context.Context,
	method, path string,
	body any,
	onEvent func(StreamEvent) error,
) error {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("glytos: encode request body: %w", err)
		}
		reader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("glytos: build request: %w", err)
	}
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "text/event-stream")
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return parseError(resp.StatusCode, resp.Header.Get("X-Request-Id"), data)
	}

	scanner := bufio.NewScanner(resp.Body)
	// A single reply can exceed bufio's default 64KB line budget, and the "done"
	// event carries the whole turn, so give the scanner room.
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	var name string
	var data []string
	flush := func() error {
		if name == "" || len(data) == 0 {
			name, data = "", nil
			return nil
		}
		event, ok := parseSSE(name, strings.Join(data, "\n"))
		name, data = "", nil
		if !ok {
			return nil
		}
		return onEvent(event)
	}

	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case line == "":
			if err := flush(); err != nil {
				return err
			}
		case strings.HasPrefix(line, "event:"):
			name = strings.TrimSpace(line[len("event:"):])
		case strings.HasPrefix(line, "data:"):
			data = append(data, strings.TrimSpace(line[len("data:"):]))
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("glytos: read stream: %w", err)
	}
	// A stream that ends without a trailing blank line still has one event to give.
	return flush()
}

func parseSSE(name, payload string) (StreamEvent, bool) {
	switch name {
	case "token":
		var body struct {
			Delta string `json:"delta"`
		}
		_ = json.Unmarshal([]byte(payload), &body)
		return StreamEvent{Type: "token", Delta: body.Delta}, true
	case "error":
		var body struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal([]byte(payload), &body)
		if body.Message == "" {
			body.Message = "stream failed"
		}
		return StreamEvent{Type: "error", Message: body.Message}, true
	case "done":
		return StreamEvent{Type: "done", Run: json.RawMessage(payload)}, true
	}
	return StreamEvent{}, false
}

// UploadFile posts a multipart body. It is separate from Do because the
// Content-Type has to carry the boundary the writer generates - setting it by
// hand produces a body the server cannot parse.
func (c *Client) UploadFile(
	ctx context.Context,
	path string,
	fields map[string]string,
	filename string,
	content []byte,
	out any,
) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return fmt.Errorf("glytos: write form field: %w", err)
		}
	}
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return fmt.Errorf("glytos: write form file: %w", err)
	}
	if _, err := part.Write(content); err != nil {
		return fmt.Errorf("glytos: write form file: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("glytos: close form: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+path, &buf)
	if err != nil {
		return fmt.Errorf("glytos: build request: %w", err)
	}
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.environment != "" {
		req.Header.Set("X-Environment-Id", c.environment)
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
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseError(resp.StatusCode, resp.Header.Get("X-Request-Id"), data)
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("glytos: decode response: %w", err)
	}
	return nil
}
