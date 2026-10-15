package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

const DefaultSystemPrompt = `You are a Unix CLI assistant embedded in a command-line pipeline.
- Be direct and concise. Skip pleasantries, introductions, and filler.
- When analyzing code, diffs, or logs: pinpoint root causes, errors, and supply exact, minimal code or commands.
- For general questions: provide accurate, direct answers without unnecessary fluff.
- Format output with clean, readable Markdown for terminal display.`

type Config struct {
	Provider     string `yaml:"provider,omitempty" mapstructure:"provider"`
	Model        string `yaml:"model,omitempty" mapstructure:"model"`
	APIKey       string `yaml:"api_key,omitempty" mapstructure:"api_key"`
	BaseURL      string `yaml:"base_url,omitempty" mapstructure:"base_url"`
	SystemPrompt string `yaml:"system_prompt,omitempty" mapstructure:"system_prompt"`
}

type Profile struct {
	Provider     string `yaml:"provider" mapstructure:"provider"`
	Model        string `yaml:"model" mapstructure:"model"`
	APIKey       string `yaml:"api_key,omitempty" mapstructure:"api_key"`
	BaseURL      string `yaml:"base_url" mapstructure:"base_url"`
	SystemPrompt string `yaml:"system_prompt" mapstructure:"system_prompt"`
}

type File struct {
	ActiveProfile string             `yaml:"active_profile,omitempty"`
	Profiles      map[string]Profile `yaml:"profiles,omitempty"`
	Config        `yaml:",inline"`
}

func (p Profile) Config() Config {
	return Config{
		Provider:     p.Provider,
		Model:        p.Model,
		APIKey:       p.APIKey,
		BaseURL:      p.BaseURL,
		SystemPrompt: p.SystemPrompt,
	}
}

func ProfileFromConfig(cfg Config) Profile {
	return Profile{
		Provider:     cfg.Provider,
		Model:        cfg.Model,
		APIKey:       cfg.APIKey,
		BaseURL:      cfg.BaseURL,
		SystemPrompt: cfg.SystemPrompt,
	}
}

func ConfigPath(override ...string) (string, error) {
	if len(override) > 0 && override[0] != "" {
		return override[0], nil
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "whaeh", "config.yaml"), nil
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "whaeh", "config.yaml"), nil
}

func Load(override ...string) (Config, error) {
	doc, err := LoadFile(override...)
	if err != nil {
		return Config{}, err
	}
	cfg := doc.Config
	if len(doc.Profiles) > 0 {
		name := doc.ActiveProfile
		if _, ok := doc.Profiles[name]; !ok {
			names := ProfileNames(doc)
			name = names[0]
		}
		cfg = doc.Profiles[name].Config()
	}
	applyDefaults(&cfg)
	return cfg, nil
}

func LoadFile(override ...string) (File, error) {
	target, err := ConfigPath(override...)
	if err != nil {
		return File{}, err
	}
	data, err := os.ReadFile(target)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return File{Profiles: map[string]Profile{}}, nil
		}
		return File{}, fmt.Errorf("read config: %w", err)
	}
	var doc File
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return File{}, fmt.Errorf("decode config: %w", err)
	}
	if doc.Profiles == nil {
		doc.Profiles = map[string]Profile{}
	}
	return doc, nil
}

func SaveFile(doc File, override ...string) (string, error) {
	target, err := ConfigPath(override...)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return "", err
	}
	data, err := yaml.Marshal(doc)
	if err != nil {
		return "", fmt.Errorf("encode config: %w", err)
	}
	if err := os.WriteFile(target, data, 0o600); err != nil {
		return "", err
	}
	return target, nil
}

func ProfileNames(doc File) []string {
	names := make([]string, 0, len(doc.Profiles))
	for name := range doc.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func Save(cfg Config, override ...string) (string, error) {
	doc, err := LoadFile(override...)
	if err != nil {
		return "", err
	}
	if len(doc.Profiles) > 0 {
		name := doc.ActiveProfile
		if _, ok := doc.Profiles[name]; !ok {
			name = ProfileNames(doc)[0]
			doc.ActiveProfile = name
		}
		doc.Profiles[name] = ProfileFromConfig(cfg)
		return SaveFile(doc, override...)
	}
	doc.Config = cfg
	return SaveFile(doc, override...)
}

func applyDefaults(cfg *Config) {
	if cfg.Provider == "" {
		cfg.Provider = "ollama"
	}
	if cfg.Model == "" {
		cfg.Model = "llama3.2"
	}
	if cfg.SystemPrompt == "" {
		cfg.SystemPrompt = DefaultSystemPrompt
	}
}
