package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeConfig(t *testing.T, dir, appID, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, appID+".json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestLoad_Success(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "test-app", `{
		"app_id": "test-app",
		"name": "Test App",
		"scope": ["feature1", "feature2"],
		"tone": "casual",
		"fallback_message": "Sorry.",
		"system_prompt": "You are test."
	}`)

	loader := NewLoader(dir)
	cfg, err := loader.Load("test-app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AppID != "test-app" {
		t.Fatalf("expected test-app, got %q", cfg.AppID)
	}
	if cfg.Name != "Test App" {
		t.Fatalf("expected Test App, got %q", cfg.Name)
	}
	if len(cfg.Scope) != 2 || cfg.Scope[0] != "feature1" {
		t.Fatalf("unexpected scope: %v", cfg.Scope)
	}
	if cfg.Tone != "casual" {
		t.Fatalf("expected casual, got %q", cfg.Tone)
	}
	if cfg.FallbackMsg != "Sorry." {
		t.Fatalf("expected Sorry., got %q", cfg.FallbackMsg)
	}
	if cfg.SystemPrompt != "You are test." {
		t.Fatalf("unexpected system_prompt: %q", cfg.SystemPrompt)
	}
}

func TestLoad_Cache(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "cached-app", `{"app_id":"cached-app","name":"Cached"}`)

	loader := NewLoader(dir)
	cfg, err := loader.Load("cached-app")
	if err != nil {
		t.Fatalf("first load: %v", err)
	}
	if cfg.Name != "Cached" {
		t.Fatalf("expected Cached, got %q", cfg.Name)
	}

	// delete file, should still return cached
	os.Remove(filepath.Join(dir, "cached-app.json"))
	cfg2, err := loader.Load("cached-app")
	if err != nil {
		t.Fatalf("cached load: %v", err)
	}
	if cfg2.Name != "Cached" {
		t.Fatalf("expected Cached from cache, got %q", cfg2.Name)
	}
}

func TestLoad_NotFound(t *testing.T) {
	dir := t.TempDir()
	loader := NewLoader(dir)
	_, err := loader.Load("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing app config")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "broken", "this is not json")

	loader := NewLoader(dir)
	_, err := loader.Load("broken")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
