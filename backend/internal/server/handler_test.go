package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"amentra/internal/chat"
	"amentra/internal/config"
	"amentra/internal/llm"
)

type mockHandlerLoader struct {
	cfg *config.AppConfig
	err error
}

func (m *mockHandlerLoader) Load(_ string) (*config.AppConfig, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.cfg, nil
}

type mockHandlerLLM struct {
	chatCompletion func(context.Context, []llm.Message) (string, error)
}

func (m *mockHandlerLLM) ChatCompletion(ctx context.Context, msgs []llm.Message) (string, error) {
	return m.chatCompletion(ctx, msgs)
}

func (m *mockHandlerLLM) StreamChat(_ context.Context, _ []llm.Message) (<-chan string, <-chan error) {
	tokenCh := make(chan string)
	errCh := make(chan error, 1)
	close(tokenCh)
	errCh <- nil
	return tokenCh, errCh
}

func TestHandleChat_Success(t *testing.T) {
	cfg := &config.AppConfig{AppID: "test", Name: "Test"}
	llm := &mockHandlerLLM{
		chatCompletion: func(_ context.Context, _ []llm.Message) (string, error) {
			return "Hello!", nil
		},
	}
	svc := chat.NewService(&mockHandlerLoader{cfg: cfg}, chat.NewPromptBuilder(), llm)
	srv := &Server{addr: ":0", svc: svc}

	body, _ := json.Marshal(chat.Req{AppID: "test", Message: "halo"})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/chat", bytes.NewReader(body))
	srv.handleChat(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp chat.ChatResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Reply != "Hello!" {
		t.Fatalf("expected 'Hello!', got %q", resp.Reply)
	}
}

func TestHandleChat_MethodNotAllowed(t *testing.T) {
	srv := &Server{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/chat", nil)
	srv.handleChat(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandleChat_BadJSON(t *testing.T) {
	srv := &Server{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/chat", strings.NewReader("not json"))
	srv.handleChat(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleChat_ValidationError(t *testing.T) {
	cfg := &config.AppConfig{AppID: "test", Name: "Test"}
	svc := chat.NewService(&mockHandlerLoader{cfg: cfg}, chat.NewPromptBuilder(), &mockHandlerLLM{})
	srv := &Server{addr: ":0", svc: svc}

	body, _ := json.Marshal(chat.Req{AppID: "", Message: "halo"})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/chat", bytes.NewReader(body))
	srv.handleChat(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleChat_ServiceError(t *testing.T) {
	svc := chat.NewService(&mockHandlerLoader{err: errors.New("not found")}, chat.NewPromptBuilder(), &mockHandlerLLM{})
	srv := &Server{addr: ":0", svc: svc}

	body, _ := json.Marshal(chat.Req{AppID: "missing", Message: "halo"})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/chat", bytes.NewReader(body))
	srv.handleChat(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleChatStream_Success(t *testing.T) {
	cfg := &config.AppConfig{AppID: "test", Name: "Test"}
	llm := &mockHandlerLLM{
		chatCompletion: func(_ context.Context, _ []llm.Message) (string, error) {
			return "", errors.New("should not call ChatCompletion")
		},
	}
	svc := chat.NewService(&mockHandlerLoader{cfg: cfg}, chat.NewPromptBuilder(), llm)
	srv := &Server{addr: ":0", svc: svc}

	body, _ := json.Marshal(chat.Req{AppID: "test", Message: "halo"})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/chat-stream", bytes.NewReader(body))
	srv.handleChatStream(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleChatStream_MethodNotAllowed(t *testing.T) {
	srv := &Server{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/chat-stream", nil)
	srv.handleChatStream(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandleChatStream_BadJSON(t *testing.T) {
	srv := &Server{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/chat-stream", strings.NewReader("not json"))
	srv.handleChatStream(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleUpdateSummary_Success(t *testing.T) {
	cfg := &config.AppConfig{AppID: "test", Name: "Test"}
	llm := &mockHandlerLLM{
		chatCompletion: func(_ context.Context, _ []llm.Message) (string, error) {
			return "updated summary", nil
		},
	}
	svc := chat.NewService(&mockHandlerLoader{cfg: cfg}, chat.NewPromptBuilder(), llm)
	srv := &Server{addr: ":0", svc: svc}

	body, _ := json.Marshal(chat.Req{AppID: "test", Message: "halo", Summary: "old summary"})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/summary", bytes.NewReader(body))
	srv.handleUpdateSummary(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp chat.ChatResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Summary != "updated summary" {
		t.Fatalf("expected 'updated summary', got %q", resp.Summary)
	}
}

func TestHandleUpdateSummary_ValidationError(t *testing.T) {
	cfg := &config.AppConfig{AppID: "test", Name: "Test"}
	svc := chat.NewService(&mockHandlerLoader{cfg: cfg}, chat.NewPromptBuilder(), &mockHandlerLLM{})
	srv := &Server{addr: ":0", svc: svc}

	body, _ := json.Marshal(chat.Req{AppID: "test", Message: "halo", Summary: ""})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/summary", bytes.NewReader(body))
	srv.handleUpdateSummary(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleUpdateSummary_MethodNotAllowed(t *testing.T) {
	srv := &Server{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/summary", nil)
	srv.handleUpdateSummary(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
