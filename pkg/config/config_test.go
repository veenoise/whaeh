package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadDefaultSystemPrompt(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.SystemPrompt != DefaultSystemPrompt {
		t.Errorf("expected default system prompt, got %q", cfg.SystemPrompt)
	}
}

func TestConfigEnvironmentSystemPromptIgnored(t *testing.T) {
	t.Setenv("WHAEH_SYSTEM_PROMPT", "Custom prompt")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.SystemPrompt != DefaultSystemPrompt {
		t.Errorf("expected environment value to be ignored, got %q", cfg.SystemPrompt)
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
	_, err = os.UserHomeDir()
	if err == nil {
		path, err = ConfigPath()
		if err != nil {
			t.Fatalf("ConfigPath() without XDG failed: %v", err)
		}
		var expectedHome string
		expectedHome, err = os.UserConfigDir()
		if err != nil {
			t.Fatalf("UserConfigDir() failed: %v", err)
		}
		expectedHome = filepath.Join(expectedHome, "whaeh", "config.yaml")
		if path != expectedHome {
			t.Errorf("got %s, want %s", path, expectedHome)
		}
	}

	custom := "/tmp/custom-config.yaml"
	path, err = ConfigPath(custom)
	if err != nil || path != custom {
		t.Errorf("got %s, want %s", path, custom)
	}
}

func TestConfigOverrideFile(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "custom.yaml")
	want := Config{
		Provider:     "openai",
		Model:        "gpt-4o",
		APIKey:       "custom-key",
		BaseURL:      "https://custom-url",
		SystemPrompt: "custom prompt",
	}

	savedPath, err := Save(want, tempFile)
	if err != nil {
		t.Fatalf("Save with override failed: %v", err)
	}
	if savedPath != tempFile {
		t.Fatalf("expected savedPath %s, got %s", tempFile, savedPath)
	}

	got, err := Load(tempFile)
	if err != nil {
		t.Fatalf("Load with override failed: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLoadActiveProfile(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "config.yaml")
	doc := File{
		ActiveProfile: "work",
		Profiles: map[string]Profile{
			"default": {
				Provider: "ollama",
				Model:    "llama3.2",
			},
			"work": {
				Provider:     "openai",
				Model:        "gpt-4o",
				BaseURL:      "https://api.openai.com/v1",
				SystemPrompt: "work prompt",
			},
		},
	}
	if _, err := SaveFile(doc, tempFile); err != nil {
		t.Fatalf("SaveFile() failed: %v", err)
	}

	got, err := Load(tempFile)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if got.Provider != "openai" || got.Model != "gpt-4o" || got.SystemPrompt != "work prompt" {
		t.Fatalf("Load() returned wrong active profile: %+v", got)
	}
}

func TestProfileRoundTripPreservesProfiles(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "config.yaml")
	doc := File{
		ActiveProfile: "default",
		Profiles: map[string]Profile{
			"default": {Provider: "openai", Model: "gpt-4o", APIKey: "secret"},
			"local":   {Provider: "ollama", Model: "llama3.2"},
		},
	}
	if _, err := SaveFile(doc, tempFile); err != nil {
		t.Fatalf("SaveFile() failed: %v", err)
	}
	if _, err := Save(Config{Provider: "openai", Model: "gpt-4.1"}, tempFile); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}
	got, err := LoadFile(tempFile)
	if err != nil {
		t.Fatalf("LoadFile() failed: %v", err)
	}
	if len(got.Profiles) != 2 {
		t.Fatalf("expected two profiles, got %d", len(got.Profiles))
	}
	if got.Profiles["local"].Provider != "ollama" {
		t.Fatal("Save() changed inactive profile")
	}
	if got.Profiles["default"].Model != "gpt-4.1" {
		t.Fatalf("Save() did not update active profile: %+v", got.Profiles["default"])
	}
}

func TestLoadMissingActiveProfileFallsBackToSortedName(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "config.yaml")
	doc := File{
		ActiveProfile: "missing",
		Profiles: map[string]Profile{
			"zulu":  {Provider: "z"},
			"alpha": {Provider: "a"},
		},
	}
	if _, err := SaveFile(doc, tempFile); err != nil {
		t.Fatalf("SaveFile() failed: %v", err)
	}
	got, err := Load(tempFile)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if got.Provider != "a" {
		t.Fatalf("expected alpha fallback, got %q", got.Provider)
	}
}
