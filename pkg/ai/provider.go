package ai

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type ProviderRequest struct {
	Prompt  string
	Context string
	Model   string
}

type Provider interface {
	StreamCompletion(context.Context, ProviderRequest, io.Writer) error
}

func NewProvider(name, apiKey, baseURL string) (Provider, error) {
	switch strings.ToLower(name) {
	case "gemini":
		return NewGemini(apiKey, baseURL), nil
	case "openai":
		return NewOpenAI(apiKey, baseURL), nil
	case "ollama":
		return NewOllama(baseURL), nil
	default:
		return nil, fmt.Errorf("unknown provider %q (choose gemini, openai, or ollama)", name)
	}
}
