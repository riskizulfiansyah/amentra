package chat

import (
	"strings"
	"testing"
)

func TestValidate_EmptyAppID(t *testing.T) {
	r := &Req{Message: "halo"}
	err := r.Validate()
	if err == nil || err.Error() != "app_id is required" {
		t.Fatalf("expected app_id error, got: %v", err)
	}
}

func TestValidate_MessageTooLong(t *testing.T) {
	r := &Req{AppID: "x", Message: strings.Repeat("a", 1001)}
	err := r.Validate()
	if err == nil || err.Error() != "message too long (max 1000 chars)" {
		t.Fatalf("expected message too long error, got: %v", err)
	}
}

func TestValidate_TooManyRecent(t *testing.T) {
	r := &Req{
		AppID:          "x",
		Message:        "hi",
		RecentMessages: make([]Message, 11),
	}
	err := r.Validate()
	if err == nil || err.Error() != "too many recent_messages (max 10)" {
		t.Fatalf("expected recent_messages error, got: %v", err)
	}
}

func TestValidate_Valid(t *testing.T) {
	r := &Req{
		AppID:          "my-app",
		Message:        "halo",
		RecentMessages: make([]Message, 3),
	}
	if err := r.Validate(); err != nil {
		t.Fatalf("expected nil, got: %v", err)
	}
}

func TestValidate_ValidNoRecent(t *testing.T) {
	r := &Req{
		AppID:   "my-app",
		Message: "halo",
	}
	if err := r.Validate(); err != nil {
		t.Fatalf("expected nil, got: %v", err)
	}
}
