package chat

import "errors"

type Req struct {
	AppID          string    `json:"app_id"`
	Summary        string    `json:"summary"`
	RecentMessages []Message `json:"recent_messages"`
	Message        string    `json:"message"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Reply   string `json:"reply"`
	Summary string `json:"summary"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type TokenEvent struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type DoneEvent struct {
	Type    string `json:"type"`
	Reply   string `json:"reply"`
	Summary string `json:"summary"`
}

type ErrorEvent struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (r *Req) Validate() error {
	if r.AppID == "" {
		return errors.New("app_id is required")
	}
	if len(r.Message) > 1000 {
		return errors.New("message too long (max 1000 chars)")
	}
	if len(r.RecentMessages) > 10 {
		return errors.New("too many recent_messages (max 10)")
	}
	return nil
}
