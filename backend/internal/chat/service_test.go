package chat

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"amentra/internal/config"

	"amentra/internal/llm"
)

type mockLLM struct {
	chatCompletion func(ctx context.Context, messages []llm.Message) (string, error)
	streamChat     func(ctx context.Context, messages []llm.Message) (<-chan string, <-chan error)
}

func (m *mockLLM) ChatCompletion(ctx context.Context, messages []llm.Message) (string, error) {
	return m.chatCompletion(ctx, messages)
}

func (m *mockLLM) StreamChat(ctx context.Context, messages []llm.Message) (<-chan string, <-chan error) {
	return m.streamChat(ctx, messages)
}

func TestChat_Success(t *testing.T) {
	cfg := &config.AppConfig{AppID: "test", Name: "Test"}
	cfgLoader := &mockLoader{cfg: cfg}
	prompt := NewPromptBuilder()
	llm := &mockLLM{
		chatCompletion: func(_ context.Context, _ []llm.Message) (string, error) {
			return "Hello from LLM", nil
		},
	}
	svc := NewService(cfgLoader, prompt, llm)

	resp, err := svc.Chat(context.Background(), &Req{
		AppID:   "test",
		Message: "halo",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Reply != "Hello from LLM" {
		t.Fatalf("expected 'Hello from LLM', got %q", resp.Reply)
	}
}

func TestChat_LLMError(t *testing.T) {
	cfg := &config.AppConfig{AppID: "test", Name: "Test"}
	cfgLoader := &mockLoader{cfg: cfg}
	prompt := NewPromptBuilder()
	llm := &mockLLM{
		chatCompletion: func(_ context.Context, _ []llm.Message) (string, error) {
			return "", errors.New("api error")
		},
	}
	svc := NewService(cfgLoader, prompt, llm)

	_, err := svc.Chat(context.Background(), &Req{
		AppID:   "test",
		Message: "halo",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestChat_CfgLoadError(t *testing.T) {
	cfgLoader := &mockLoader{err: errors.New("not found")}
	prompt := NewPromptBuilder()
	svc := NewService(cfgLoader, prompt, &mockLLM{})

	_, err := svc.Chat(context.Background(), &Req{AppID: "missing", Message: "halo"})
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestUpdateSummary_Success(t *testing.T) {
	cfgLoader := &mockLoader{cfg: &config.AppConfig{AppID: "test", Name: "Test"}}
	prompt := NewPromptBuilder()
	llm := &mockLLM{
		chatCompletion: func(_ context.Context, _ []llm.Message) (string, error) {
			return "updated summary", nil
		},
	}
	svc := NewService(cfgLoader, prompt, llm)

	summary, err := svc.UpdateSummary(context.Background(), &Req{
		AppID:   "test",
		Summary: "old summary",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary != "updated summary" {
		t.Fatalf("expected 'updated summary', got %q", summary)
	}
}

func TestChatStream_Success(t *testing.T) {
	cfg := &config.AppConfig{AppID: "test", Name: "Test"}
	cfgLoader := &mockLoader{cfg: cfg}
	prompt := NewPromptBuilder()
	llm := &mockLLM{
		streamChat: func(_ context.Context, _ []llm.Message) (<-chan string, <-chan error) {
			tokenCh := make(chan string)
			errCh := make(chan error, 1)
			go func() {
				tokenCh <- "Hello"
				tokenCh <- " world"
				close(tokenCh)
				errCh <- nil
			}()
			return tokenCh, errCh
		},
	}
	svc := NewService(cfgLoader, prompt, llm)

	w := httptest.NewRecorder()
	_, _, err := svc.ChatStream(context.Background(), &Req{
		AppID:   "test",
		Message: "halo",
	}, w, w.Flush)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Hello") || !strings.Contains(body, "world") {
		t.Fatalf("expected SSE events with Hello world, got: %q", body)
	}
	if !strings.Contains(body, "data: ") {
		t.Fatalf("expected SSE data events, got: %q", body)
	}
}

func TestChatStream_LLMError(t *testing.T) {
	cfg := &config.AppConfig{AppID: "test", Name: "Test"}
	cfgLoader := &mockLoader{cfg: cfg}
	prompt := NewPromptBuilder()
	llm := &mockLLM{
		streamChat: func(_ context.Context, _ []llm.Message) (<-chan string, <-chan error) {
			tokenCh := make(chan string)
			errCh := make(chan error, 1)
			go func() {
				close(tokenCh)
				errCh <- errors.New("stream failed")
			}()
			return tokenCh, errCh
		},
	}
	svc := NewService(cfgLoader, prompt, llm)

	w := httptest.NewRecorder()
	_, _, err := svc.ChatStream(context.Background(), &Req{
		AppID:   "test",
		Message: "halo",
	}, w, w.Flush)
	if err != nil {
		t.Fatalf("expected nil error (stream error wrapped in SSE), got %v", err)
	}

	body := w.Body.String()
	if !strings.Contains(body, "stream failed") {
		t.Fatalf("expected SSE error event with 'stream failed', got: %q", body)
	}
}

func TestChatStream_CfgLoadError(t *testing.T) {
	cfgLoader := &mockLoader{err: errors.New("not found")}
	prompt := NewPromptBuilder()
	svc := NewService(cfgLoader, prompt, &mockLLM{})

	_, _, err := svc.ChatStream(context.Background(), &Req{AppID: "missing", Message: "halo"}, httptest.NewRecorder(), httptest.NewRecorder().Flush)
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestUpdateSummary_LLMFallback(t *testing.T) {
	cfgLoader := &mockLoader{cfg: &config.AppConfig{AppID: "test", Name: "Test"}}
	prompt := NewPromptBuilder()
	llm := &mockLLM{
		chatCompletion: func(_ context.Context, _ []llm.Message) (string, error) {
			return "", errors.New("api error")
		},
	}
	svc := NewService(cfgLoader, prompt, llm)

	summary, err := svc.UpdateSummary(context.Background(), &Req{
		AppID:   "test",
		Summary: "existing summary",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary != "existing summary" {
		t.Fatalf("expected fallback to existing summary, got %q", summary)
	}
}

func TestUpdateSummary_Empty(t *testing.T) {
	cfgLoader := &mockLoader{cfg: &config.AppConfig{AppID: "test", Name: "Test"}}
	prompt := NewPromptBuilder()
	svc := NewService(cfgLoader, prompt, &mockLLM{})

	_, err := svc.UpdateSummary(context.Background(), &Req{
		AppID:   "test",
		Summary: "",
	})
	if err == nil {
		t.Fatal("expected error for empty summary")
	}
}

type mockLoader struct {
	cfg *config.AppConfig
	err error
}

func (m *mockLoader) Load(_ string) (*config.AppConfig, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.cfg, nil
}
