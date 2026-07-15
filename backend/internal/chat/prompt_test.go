package chat

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"amentra/internal/config"
)

func TestBuild_Basic(t *testing.T) {
	b := NewPromptBuilder()
	cfg := &config.AppConfig{
		AppID:       "test-app",
		Name:        "Test App",
		Scope:       []string{"feature1"},
		FallbackMsg: "I only answer about feature1.",
	}

	msgs := b.Build(cfg, "", nil, "hello")

	if len(msgs) < 2 {
		t.Fatalf("expected at least 2 messages (system + user), got %d", len(msgs))
	}

	if msgs[0].Role != "system" {
		t.Fatalf("first message should be system, got %q", msgs[0].Role)
	}
	if msgs[len(msgs)-1].Role != "user" {
		t.Fatalf("last message should be user, got %q", msgs[len(msgs)-1].Role)
	}
	if msgs[len(msgs)-1].Content != "hello" {
		t.Fatalf("user content should be 'hello', got %q", msgs[len(msgs)-1].Content)
	}
}

func TestBuild_WithSummary(t *testing.T) {
	b := NewPromptBuilder()
	cfg := &config.AppConfig{
		AppID:       "test-app",
		Name:        "Test App",
		Scope:       []string{"feature1"},
		FallbackMsg: "I only answer about feature1.",
	}

	msgs := b.Build(cfg, "previous context", nil, "hello")

	systemCount := 0
	for _, m := range msgs {
		if m.Role == "system" {
			systemCount++
		}
	}
	if systemCount != 2 {
		t.Fatalf("expected 2 system messages (prompt + summary), got %d", systemCount)
	}
}

func TestBuild_WithRecent(t *testing.T) {
	b := NewPromptBuilder()
	cfg := &config.AppConfig{
		AppID:       "test-app",
		Name:        "Test App",
		Scope:       []string{"feature1"},
		FallbackMsg: "I only answer about feature1.",
	}
	recent := []Message{
		{Role: "user", Content: "previous question"},
		{Role: "assistant", Content: "previous answer"},
	}

	msgs := b.Build(cfg, "", recent, "hello")

	if len(msgs) != 4 {
		t.Fatalf("expected 4 messages (system + 2 recent + user), got %d", len(msgs))
	}
	if msgs[1].Role != "user" || msgs[1].Content != "previous question" {
		t.Fatalf("unexpected msg[1]: role=%q content=%q", msgs[1].Role, msgs[1].Content)
	}
	if msgs[2].Role != "assistant" || msgs[2].Content != "previous answer" {
		t.Fatalf("unexpected msg[2]: role=%q content=%q", msgs[2].Role, msgs[2].Content)
	}
}

func TestBuild_NoScope(t *testing.T) {
	b := NewPromptBuilder()
	cfg := &config.AppConfig{
		AppID: "test-app",
		Name:  "Test App",
	}

	msgs := b.Build(cfg, "", nil, "hello")

	sys := msgs[0].Content
	if !strings.Contains(sys, "Test App") {
		t.Fatalf("expected system prompt to contain app name, got:\n%s", sys)
	}
	if strings.Contains(sys, "ONLY answer") {
		t.Fatal("expected no scope restriction when scope is empty")
	}
}

func TestBuild_WithKnowledge(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldDir)

	knowDir := filepath.Join("data/knowledge/test-app")
	os.MkdirAll(knowDir, 0755)
	os.WriteFile(filepath.Join(knowDir, "intro.md"), []byte("# About\nThis is test knowledge."), 0644)

	b := NewPromptBuilder()
	cfg := &config.AppConfig{
		AppID: "test-app",
		Name:  "Test App",
	}

	msgs := b.Build(cfg, "", nil, "hello")

	sys := msgs[0].Content
	if !strings.Contains(sys, "test knowledge") {
		t.Fatalf("expected knowledge injected into system prompt, got:\n%s", sys)
	}
}

func TestBuild_SystemPromptContainsScope(t *testing.T) {
	b := NewPromptBuilder()
	cfg := &config.AppConfig{
		AppID:       "test-app",
		Name:        "My App",
		Scope:       []string{"about", "projects", "skills"},
		FallbackMsg: "Only about my app.",
	}

	msgs := b.Build(cfg, "", nil, "halo")

	sys := msgs[0].Content
	if !containsAll(sys, "My App", "about", "projects", "skills", "Only about my app.") {
		t.Fatalf("system prompt missing required parts:\n%s", sys)
	}
}

func TestLoadKnowledge_NotFound(t *testing.T) {
	got := loadKnowledge("nonexistent-app")
	if got != "" {
		t.Fatalf("expected empty string for missing knowledge dir, got %q", got)
	}
}

func TestLoadKnowledge_SkipsNonMarkdown(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldDir)

	knowDir := filepath.Join("data/knowledge/test-app")
	os.MkdirAll(knowDir, 0755)
	os.WriteFile(filepath.Join(knowDir, "notes.txt"), []byte("should be ignored"), 0644)
	os.WriteFile(filepath.Join(knowDir, "facts.md"), []byte("real content"), 0644)

	got := loadKnowledge("test-app")
	if !strings.Contains(got, "real content") {
		t.Fatalf("expected 'real content', got %q", got)
	}
	if strings.Contains(got, "should be ignored") {
		t.Fatal("should not include non-markdown files")
	}
}

func containsAll(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && containsStr(s, sub)
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
