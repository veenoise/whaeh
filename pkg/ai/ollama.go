package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type Ollama struct {
	baseURL string
	client  *http.Client
}

func NewOllama(baseURL string) *Ollama {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &Ollama{baseURL: strings.TrimRight(baseURL, "/"), client: http.DefaultClient}
}

func (p *Ollama) StreamCompletion(ctx context.Context, req ProviderRequest, out io.Writer) error {
	body := map[string]any{"model": req.Model, "prompt": promptWithContext(req), "stream": true}
	resp, err := doJSON(ctx, p.client, http.MethodPost, p.baseURL+"/api/generate", nil, body)
	if err != nil {
		return err
	}
	return scanLines(resp.Body, func(line string) error {
		var event struct {
			Response string `json:"response"`
			Error    string `json:"error"`
		}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return err
		}
		if event.Error != "" {
			return &providerError{message: event.Error}
		}
		_, err := io.WriteString(out, event.Response)
		return err
	})
}
