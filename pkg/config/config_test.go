package config

import (
	"testing"
)

func TestConfigLoadDefaultSystemPrompt(t *testing.T) {
	t.Setenv("WHAEH_SYSTEM_PROMPT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.SystemPrompt != DefaultSystemPrompt {
		t.Errorf("expected default system prompt, got %q", cfg.SystemPrompt)
	}
}

func TestConfigEnvSystemPrompt(t *testing.T) {
	t.Setenv("WHAEH_SYSTEM_PROMPT", "Custom prompt")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.SystemPrompt != "Custom prompt" {
		t.Errorf("expected 'Custom prompt', got %q", cfg.SystemPrompt)
	}
}
