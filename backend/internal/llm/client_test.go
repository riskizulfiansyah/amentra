package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestBaseURL_Default(t *testing.T) {
	os.Unsetenv("AI_BASE_URL")
	defer os.Unsetenv("AI_BASE_URL")
	if got := baseURL(); got != "https://openrouter.ai/api/v1" {
		t.Fatalf("expected default, got %q", got)
	}
}

func TestBaseURL_Custom(t *testing.T) {
	os.Setenv("AI_BASE_URL", "http://localhost:11434/v1")
	defer os.Unsetenv("AI_BASE_URL")
	if got := baseURL(); got != "http://localhost:11434/v1" {
		t.Fatalf("expected custom, got %q", got)
	}
}

func TestBaseURL_TrailingSlash(t *testing.T) {
	os.Setenv("AI_BASE_URL", "http://localhost:11434/v1/")
	defer os.Unsetenv("AI_BASE_URL")
	if got := baseURL(); got != "http://localhost:11434/v1" {
		t.Fatalf("expected trailing slash trimmed, got %q", got)
	}
}

func TestModel_Default(t *testing.T) {
	os.Unsetenv("LLM_MODEL")
	defer os.Unsetenv("LLM_MODEL")
	if got := model(); got != "gpt-3.5-turbo" {
		t.Fatalf("expected default, got %q", got)
	}
}

func TestModel_Custom(t *testing.T) {
	os.Setenv("LLM_MODEL", "ollama/llama3")
	defer os.Unsetenv("LLM_MODEL")
	if got := model(); got != "ollama/llama3" {
		t.Fatalf("expected custom, got %q", got)
	}
}

func TestChatCompletion_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("expected Bearer test-key, got %q", r.Header.Get("Authorization"))
		}
		json.NewEncoder(w).Encode(chatResponse{
			Choices: []struct {
				Message      Message `json:"message"`
				FinishReason string  `json:"finish_reason"`
			}{{Message: Message{Role: "assistant", Content: "Hello!"}}},
		})
	}))
	defer srv.Close()

	os.Setenv("AI_BASE_URL", srv.URL)
	os.Setenv("AI_API_KEY", "test-key")
	defer func() {
		os.Unsetenv("AI_BASE_URL")
		os.Unsetenv("AI_API_KEY")
	}()

	reply, err := ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply != "Hello!" {
		t.Fatalf("expected 'Hello!', got %q", reply)
	}
}

func TestChatCompletion_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(chatResponse{
			Error: &struct {
				Message string `json:"message"`
			}{Message: "rate limited"},
		})
	}))
	defer srv.Close()

	os.Setenv("AI_BASE_URL", srv.URL)
	defer os.Unsetenv("AI_BASE_URL")

	_, err := ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err == nil || !strings.Contains(err.Error(), "rate limited") {
		t.Fatalf("expected 'rate limited' error, got %v", err)
	}
}

func TestChatCompletion_NoChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(chatResponse{Choices: nil})
	}))
	defer srv.Close()

	os.Setenv("AI_BASE_URL", srv.URL)
	defer os.Unsetenv("AI_BASE_URL")

	_, err := ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err == nil || !strings.Contains(err.Error(), "no choices") {
		t.Fatalf("expected 'no choices' error, got %v", err)
	}
}

func TestChatCompletion_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	os.Setenv("AI_BASE_URL", srv.URL)
	defer os.Unsetenv("AI_BASE_URL")

	_, err := ChatCompletion(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err == nil || !strings.Contains(err.Error(), "server error") {
		t.Fatalf("expected server error, got %v", err)
	}
}

func TestStreamChat_Tokens(t *testing.T) {
	chunks := []string{
		`data: {"choices":[{"delta":{"content":"Hello"}}]}` + "\n\n",
		`data: {"choices":[{"delta":{"content":" world"}}]}` + "\n\n",
		`data: [DONE]` + "\n\n",
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, c := range chunks {
			w.Write([]byte(c))
			w.(http.Flusher).Flush()
		}
	}))
	defer srv.Close()

	os.Setenv("AI_BASE_URL", srv.URL)
	defer os.Unsetenv("AI_BASE_URL")

	tokenCh, errCh := StreamChat(context.Background(), []Message{{Role: "user", Content: "hi"}})

	var tokens []string
	for t := range tokenCh {
		tokens = append(tokens, t)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := strings.Join(tokens, "")
	if got != "Hello world" {
		t.Fatalf("expected 'Hello world', got %q", got)
	}
}

func TestStreamChat_ErrorChunk(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`data: {"error":{"message":"quota exceeded"}}` + "\n\n"))
		w.(http.Flusher).Flush()
	}))
	defer srv.Close()

	os.Setenv("AI_BASE_URL", srv.URL)
	defer os.Unsetenv("AI_BASE_URL")

	tokenCh, errCh := StreamChat(context.Background(), []Message{{Role: "user", Content: "hi"}})

	for range tokenCh {
	}

	err := <-errCh
	if err == nil || !strings.Contains(err.Error(), "quota exceeded") {
		t.Fatalf("expected 'quota exceeded', got %v", err)
	}
}

func TestStreamChat_HTTPClientError(t *testing.T) {
	old := httpClient
	httpClient = &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, http.ErrAbortHandler
	})}
	defer func() { httpClient = old }()

	os.Setenv("AI_BASE_URL", "http://127.0.0.1:1")
	defer os.Unsetenv("AI_BASE_URL")

	tokenCh, errCh := StreamChat(context.Background(), []Message{{Role: "user", Content: "hi"}})
	for range tokenCh {
	}
	err := <-errCh
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestDefaultClient_ImplementsInterface(t *testing.T) {
	var _ Client = DefaultClient
	var _ Client = (*mockClient)(nil)
}

type mockClient struct{}

func (m *mockClient) ChatCompletion(_ context.Context, _ []Message) (string, error) {
	return "", nil
}
func (m *mockClient) StreamChat(_ context.Context, _ []Message) (<-chan string, <-chan error) {
	return nil, nil
}
