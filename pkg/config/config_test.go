package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadDefaults(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("WHAEH_PROVIDER", "")
	t.Setenv("WHAEH_MODEL", "")
	t.Setenv("WHAEH_API_KEY", "")
	t.Setenv("WHAEH_BASE_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.Provider != "ollama" {
		t.Errorf("expected provider ollama, got %s", cfg.Provider)
	}
	if cfg.Model != "llama3.2" {
		t.Errorf("expected model llama3.2, got %s", cfg.Model)
	}
}

func TestConfigSaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)

	want := Config{
		Provider: "openai",
		Model:    "gpt-4o",
		APIKey:   "test-key",
		BaseURL:  "https://custom.api",
	}

	if err := Save(want); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	expectedFile := filepath.Join(tempDir, "whaeh", "config.yaml")
	if _, err := os.Stat(expectedFile); err != nil {
		t.Fatalf("config file not created at %s: %v", expectedFile, err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestConfigEnvOverrides(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempDir)
	t.Setenv("WHAEH_PROVIDER", "gemini")
	t.Setenv("WHAEH_MODEL", "gemini-2.5-flash")
	t.Setenv("GEMINI_API_KEY", "env-gemini-key")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.Provider != "gemini" {
		t.Errorf("expected provider gemini, got %s", cfg.Provider)
	}
	if cfg.Model != "gemini-2.5-flash" {
		t.Errorf("expected model gemini-2.5-flash, got %s", cfg.Model)
	}
	if cfg.APIKey != "env-gemini-key" {
		t.Errorf("expected API key env-gemini-key, got %s", cfg.APIKey)
	}
}
