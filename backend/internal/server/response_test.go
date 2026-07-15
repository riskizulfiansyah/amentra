package server

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, 200, map[string]string{"reply": "hello"})

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %q", ct)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["reply"] != "hello" {
		t.Fatalf("expected reply=hello, got %v", body)
	}
}

func TestWriteJSON_Error(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, 400, map[string]string{"error": "bad request"})

	resp := w.Result()
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	if body["error"] != "bad request" {
		t.Fatalf("expected error=bad request, got %v", body)
	}
}

func TestWriteSSEHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	writeSSEHeaders(w)

	h := w.Header()
	if h.Get("Content-Type") != "text/event-stream" {
		t.Fatalf("bad Content-Type: %q", h.Get("Content-Type"))
	}
	if h.Get("Cache-Control") != "no-cache" {
		t.Fatalf("bad Cache-Control: %q", h.Get("Cache-Control"))
	}
	if h.Get("Connection") != "keep-alive" {
		t.Fatalf("bad Connection: %q", h.Get("Connection"))
	}
}

func TestWriteSSEJSON(t *testing.T) {
	w := httptest.NewRecorder()
	writeSSEJSON(w, map[string]string{"type": "token", "content": "hello"})

	body := w.Body.String()
	expected := "data: {\"content\":\"hello\",\"type\":\"token\"}\n\n"
	if body != expected {
		t.Fatalf("unexpected SSE body: %q", body)
	}
}

func TestWriteSSEJSON_TokenDoneError(t *testing.T) {
	w := httptest.NewRecorder()
	writeSSEJSON(w, map[string]string{"type": "token", "content": "Hello"})
	writeSSEJSON(w, map[string]string{"type": "token", "content": " world"})
	writeSSEJSON(w, map[string]string{"type": "done", "reply": "Hello world", "summary": "updated"})

	lines := strings.Split(strings.TrimRight(w.Body.String(), "\n"), "\n\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 SSE events, got %d: %v", len(lines), lines)
	}
}

func TestWriteSSEJSON_ErrorEvent(t *testing.T) {
	w := httptest.NewRecorder()
	writeSSEJSON(w, map[string]string{"type": "error", "message": "something went wrong"})

	body := w.Body.String()
	if !strings.Contains(body, "something went wrong") {
		t.Fatalf("expected error message in body, got: %q", body)
	}
}
