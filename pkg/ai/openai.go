package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type OpenAI struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewOpenAI(apiKey, baseURL string) *OpenAI {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAI{apiKey: apiKey, baseURL: strings.TrimRight(baseURL, "/"), client: http.DefaultClient}
}

func (p *OpenAI) StreamCompletion(ctx context.Context, req ProviderRequest, out io.Writer) error {
	body := map[string]any{
		"model": req.Model, "stream": true,
		"messages": []map[string]string{{"role": "user", "content": promptWithContext(req)}},
	}
	resp, err := doJSON(ctx, p.client, http.MethodPost, p.baseURL+"/chat/completions",
		map[string]string{"Authorization": "Bearer " + p.apiKey}, body)
	if err != nil {
		return err
	}
	return scanLines(resp.Body, func(line string) error {
		if !strings.HasPrefix(line, "data: ") {
			return nil
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			return nil
		}
		var event struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return err
		}
		if len(event.Choices) > 0 {
			_, err := io.WriteString(out, event.Choices[0].Delta.Content)
			return err
		}
		return nil
	})
}
