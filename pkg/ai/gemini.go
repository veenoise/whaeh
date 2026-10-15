package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Gemini struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewGemini(apiKey, baseURL string) *Gemini {
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta"
	}
	return &Gemini{apiKey: apiKey, baseURL: strings.TrimRight(baseURL, "/"), client: http.DefaultClient}
}

func (p *Gemini) StreamCompletion(ctx context.Context, req ProviderRequest, out io.Writer) error {
	body := map[string]any{"contents": []map[string]any{{"parts": []map[string]string{{"text": promptWithContext(req)}}}}}
	if req.SystemPrompt != "" {
		body["system_instruction"] = map[string]any{
			"parts": []map[string]string{{"text": req.SystemPrompt}},
		}
	}
	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?alt=sse&key=%s", p.baseURL, req.Model, p.apiKey)
	resp, err := doJSON(ctx, p.client, http.MethodPost, url, nil, body)
	if err != nil {
		return err
	}
	return scanLines(resp.Body, func(line string) error {
		if !strings.HasPrefix(line, "data: ") {
			return nil
		}
		var event struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
		}
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &event); err != nil {
			return err
		}
		if len(event.Candidates) > 0 && len(event.Candidates[0].Content.Parts) > 0 {
			_, err := io.WriteString(out, event.Candidates[0].Content.Parts[0].Text)
			return err
		}
		return nil
	})
}

type providerError struct{ message string }

func (e *providerError) Error() string { return e.message }

func promptWithContext(req ProviderRequest) string {
	if req.Context == "" {
		return req.Prompt
	}
	return "Use the following context to answer the request.\n\n<context>\n" + req.Context +
		"\n</context>\n\nRequest: " + req.Prompt
}
