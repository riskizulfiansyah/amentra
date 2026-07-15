package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"amentra/internal/chat"
)

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, chat.ErrorResponse{Error: "method not allowed"})
		return
	}

	var req chat.Req
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, chat.ErrorResponse{Error: "invalid JSON body"})
		return
	}

	if err := req.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, chat.ErrorResponse{Error: err.Error()})
		return
	}

	resp, err := s.svc.Chat(r.Context(), &req)
	if err != nil {
		slog.Error("chat failed", "app_id", req.AppID, "error", err)
		writeJSON(w, http.StatusInternalServerError, chat.ErrorResponse{Error: "chat failed"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleChatStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, chat.ErrorResponse{Error: "method not allowed"})
		return
	}

	var req chat.Req
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, chat.ErrorResponse{Error: "invalid json"})
		return
	}

	if err := req.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, chat.ErrorResponse{Error: err.Error()})
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, chat.ErrorResponse{Error: "stream unsupported"})
		return
	}

	writeSSEHeaders(w)

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	_, _, err := s.svc.ChatStream(ctx, &req, w, flusher.Flush)
	if err != nil {
		slog.Error("chat-stream failed", "app_id", req.AppID, "error", err)
		writeSSEJSON(w, chat.ErrorEvent{Type: "error", Message: err.Error()})
		flusher.Flush()
		return
	}
}

func (s *Server) handleUpdateSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, chat.ErrorResponse{Error: "method not allowed"})
		return
	}

	var req chat.Req
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, chat.ErrorResponse{Error: "invalid JSON body"})
		return
	}

	newSummary, err := s.svc.UpdateSummary(r.Context(), &req)
	if err != nil {
		slog.Error("summary update failed", "app_id", req.AppID, "error", err)
		writeJSON(w, http.StatusBadRequest, chat.ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, chat.ChatResponse{Summary: newSummary})
}
