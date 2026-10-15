package config

import (
	"os"
	"path/filepath"
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

func TestConfigPath(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() failed: %v", err)
	}
	expected := filepath.Join(tempDir, "whaeh", "config.yaml")
	if path != expected {
		t.Errorf("got %s, want %s", path, expected)
	}

	t.Setenv("XDG_CONFIG_HOME", "")
	home, err := os.UserHomeDir()
	if err == nil {
		path, err = ConfigPath()
		if err != nil {
			t.Fatalf("ConfigPath() without XDG failed: %v", err)
		}
		expectedHome := filepath.Join(home, ".whaeh", "config.yaml")
		if path != expectedHome {
			t.Errorf("got %s, want %s", path, expectedHome)
		}
	}
}
