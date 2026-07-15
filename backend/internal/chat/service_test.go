package chat

import (
	"context"
	"errors"
	"testing"

	"ai-chat/internal/config"
	"ai-chat/internal/llm"
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
