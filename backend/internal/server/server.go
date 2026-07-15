package server

import (
	"net/http"

	"amentra/internal/chat"
	"amentra/internal/config"
	"amentra/internal/llm"
)

type Server struct {
	addr string
	svc  *chat.Service
}

func New(addr string) *Server {
	cfgLoader := config.NewLoader("configs")
	prompt := chat.NewPromptBuilder()
	svc := chat.NewService(cfgLoader, prompt, llm.DefaultClient)

	return &Server{addr: addr, svc: svc}
}

func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/chat", logRequest(s.handleChat))
	mux.HandleFunc("/chat-stream", logRequest(s.handleChatStream))
	mux.HandleFunc("/summary", logRequest(s.handleUpdateSummary))

	return http.ListenAndServe(s.addr, cors(mux))
}
