package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type AppConfig struct {
	AppID          string   `json:"app_id"`
	Name           string   `json:"name"`
	Scope          []string `json:"scope"`
	Tone           string   `json:"tone"`
	FallbackMsg    string   `json:"fallback_message"`
	SystemPrompt   string   `json:"system_prompt"`
}

type Loader struct {
	mu   sync.RWMutex
	dir  string
	cache map[string]*AppConfig
}

func NewLoader(dir string) *Loader {
	return &Loader{
		dir:   dir,
		cache: make(map[string]*AppConfig),
	}
}

func (l *Loader) Load(appID string) (*AppConfig, error) {
	l.mu.RLock()
	cfg, ok := l.cache[appID]
	l.mu.RUnlock()
	if ok {
		return cfg, nil
	}

	path := fmt.Sprintf("%s/%s.json", l.dir, appID)
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("app config not found for %s: %w", appID, err)
	}
	defer f.Close()

	cfg = &AppConfig{}
	if err := json.NewDecoder(f).Decode(cfg); err != nil {
		return nil, fmt.Errorf("invalid app config %s: %w", appID, err)
	}

	l.mu.Lock()
	l.cache[appID] = cfg
	l.mu.Unlock()

	return cfg, nil
}
