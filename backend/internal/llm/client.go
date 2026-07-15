package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

var httpClient = http.DefaultClient

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Client interface {
	ChatCompletion(ctx context.Context, messages []Message) (string, error)
	StreamChat(ctx context.Context, messages []Message) (<-chan string, <-chan error)
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func baseURL() string {
	if v := os.Getenv("AI_BASE_URL"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return "https://openrouter.ai/api/v1"
}

func apiKey() string {
	return os.Getenv("AI_API_KEY")
}

func model() string {
	if v := os.Getenv("LLM_MODEL"); v != "" {
		return v
	}
	return "gpt-3.5-turbo"
}

func ChatCompletion(ctx context.Context, messages []Message) (string, error) {
	body := chatRequest{
		Model:    model(),
		Messages: messages,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL()+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey())

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return "", fmt.Errorf("server error: %d", resp.StatusCode)
	}

	var result chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("api error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return result.Choices[0].Message.Content, nil
}

func StreamChat(ctx context.Context, messages []Message) (<-chan string, <-chan error) {
	tokenCh := make(chan string)
	errCh := make(chan error, 1)

	go func() {
		defer close(tokenCh)
		defer close(errCh)

		body := chatRequest{
			Model:    model(),
			Messages: messages,
			Stream:   true,
		}

		payload, err := json.Marshal(body)
		if err != nil {
			errCh <- fmt.Errorf("marshal request: %w", err)
			return
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL()+"/chat/completions", bytes.NewReader(payload))
		if err != nil {
			errCh <- fmt.Errorf("create request: %w", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+apiKey())

		resp, err := httpClient.Do(req)
		if err != nil {
			errCh <- fmt.Errorf("http request: %w", err)
			return
		}
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				continue
			}

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			if data == "[DONE]" {
				errCh <- nil
				return
			}

			var chunk streamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if chunk.Error != nil {
				errCh <- fmt.Errorf("api error: %s", chunk.Error.Message)
				return
			}

			for _, choice := range chunk.Choices {
				if choice.Delta.Content != "" {
					tokenCh <- choice.Delta.Content
				}
			}
		}

		if err := scanner.Err(); err != nil {
			errCh <- fmt.Errorf("scan stream: %w", err)
			return
		}

		errCh <- nil
	}()

	return tokenCh, errCh
}


