package chat

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"ai-chat/internal/config"
	"ai-chat/internal/llm"
)

type PromptBuilder struct{}

func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{}
}

func (b *PromptBuilder) Build(cfg *config.AppConfig, summary string, recent []Message, userMsg string) []llm.Message {
	systemPrompt := fmt.Sprintf("You are an assistant for %s.", cfg.Name)

	if len(cfg.Scope) > 0 {
		var scopeLines []string
		for _, s := range cfg.Scope {
			scopeLines = append(scopeLines, "- "+s)
		}
		systemPrompt += fmt.Sprintf(`
		
You ONLY answer questions related to:
%s

If the question is outside the scope:
Respond with:
"%s"`, strings.Join(scopeLines, "\n"), cfg.FallbackMsg)
	}

	systemPrompt += `

Guidelines:
- Use markdown formatting: bullet points (-) or numbered lists for lists, **bold** for emphasis, inline code for technical terms
- Be factual, do not hallucinate
- If unsure, say you don't know
- Respond only in Indonesian or English`

	if knowledge := loadKnowledge(cfg.AppID); knowledge != "" {
		systemPrompt += "\n\nKnowledge:\n" + knowledge
	}

	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
	}

	if summary != "" {
		messages = append(messages, llm.Message{Role: "system", Content: "Conversation summary:\n" + summary})
	}

	for _, m := range recent {
		messages = append(messages, llm.Message{Role: m.Role, Content: m.Content})
	}

	messages = append(messages, llm.Message{Role: "user", Content: userMsg})

	return messages
}

func loadKnowledge(appID string) string {
	dir := filepath.Join("data/knowledge", appID)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	var parts []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".md")

		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}

		parts = append(parts, "--- "+name+" ---\n"+strings.TrimSpace(string(data)))
	}

	return strings.Join(parts, "\n\n")
}
