package chat

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"ai-chat/internal/config"
	"ai-chat/internal/llm"
)

type configLoader interface {
	Load(string) (*config.AppConfig, error)
}

type Service struct {
	cfgLoader configLoader
	prompt    *PromptBuilder
	llm       llm.Client
}

func NewService(cfgLoader configLoader, prompt *PromptBuilder, llm llm.Client) *Service {
	return &Service{
		cfgLoader: cfgLoader,
		prompt:    prompt,
		llm:       llm,
	}
}

func (s *Service) Chat(ctx context.Context, req *Req) (*ChatResponse, error) {
	appCfg, err := s.cfgLoader.Load(req.AppID)
	if err != nil {
		return nil, err
	}

	slog.Info("llm call", "app_id", req.AppID, "msg_len", len(req.Message))

	messages := s.prompt.Build(appCfg, req.Summary, req.RecentMessages, req.Message)
	reply, err := s.llm.ChatCompletion(ctx, messages)
	if err != nil {
		return nil, err
	}

	return &ChatResponse{Reply: reply, Summary: req.Summary}, nil
}

func (s *Service) ChatStream(ctx context.Context, req *Req, w http.ResponseWriter, flush func()) (string, string, error) {
	appCfg, err := s.cfgLoader.Load(req.AppID)
	if err != nil {
		return "", "", err
	}

	slog.Info("llm stream", "app_id", req.AppID, "msg_len", len(req.Message))

	messages := s.prompt.Build(appCfg, req.Summary, req.RecentMessages, req.Message)
	tokenCh, errCh := s.llm.StreamChat(ctx, messages)

	fullReply := ""
	for token := range tokenCh {
		fullReply += token
		evt, _ := json.Marshal(TokenEvent{Type: "token", Content: token})
		_, _ = w.Write([]byte("data: " + string(evt) + "\n\n"))
		flush()
	}

	if err := <-errCh; err != nil {
		slog.Error("llm stream error", "error", err)
		evt, _ := json.Marshal(ErrorEvent{Type: "error", Message: err.Error()})
		_, _ = w.Write([]byte("data: " + string(evt) + "\n\n"))
		flush()
		return fullReply, req.Summary, nil
	}

	evt, _ := json.Marshal(DoneEvent{Type: "done", Reply: fullReply, Summary: req.Summary})
	_, _ = w.Write([]byte("data: " + string(evt) + "\n\n"))
	flush()
	return fullReply, req.Summary, nil
}

func (s *Service) UpdateSummary(ctx context.Context, req *Req) (string, error) {
	if req.Summary == "" {
		return "", errors.New("summary is required")
	}

	messages := []llm.Message{
		{Role: "system", Content: "Update the conversation summary. Keep it under 100 words. Retain key facts, preferences, constraints."},
		{Role: "user", Content: "Previous summary:\n" + req.Summary},
	}

	reply, err := s.llm.ChatCompletion(ctx, messages)
	if err != nil {
		return req.Summary, nil
	}

	return reply, nil
}

func (s *Service) updateSummary(ctx context.Context, oldSummary, userMsg, reply string) string {
	if userMsg == "" && reply == "" {
		return oldSummary
	}

	messages := []llm.Message{
		{Role: "system", Content: "Update the conversation summary. Keep it under 100 words. Retain key facts, preferences, constraints."},
		{Role: "user", Content: "Previous summary:\n" + oldSummary},
		{Role: "user", Content: "New exchange:\nUser: " + userMsg + "\nAssistant: " + reply},
	}

	newSummary, err := s.llm.ChatCompletion(ctx, messages)
	if err != nil {
		return oldSummary
	}

	return newSummary
}
