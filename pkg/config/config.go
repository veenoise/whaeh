package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Provider string `yaml:"provider"`
	Model    string `yaml:"model"`
	APIKey   string `yaml:"api_key"`
	BaseURL  string `yaml:"base_url"`
}

func Load() (Config, error) {
	cfg := Config{
		Provider: "ollama",
		Model:    "llama3.2",
	}

	for _, p := range configPaths() {
		data, err := os.ReadFile(p)
		if err == nil {
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return Config{}, fmt.Errorf("decode config %s: %w", p, err)
			}
			break
		} else if !errors.Is(err, os.ErrNotExist) {
			return Config{}, fmt.Errorf("read config %s: %w", p, err)
		}
	}

	if v := os.Getenv("WHAEH_PROVIDER"); v != "" {
		cfg.Provider = v
	}
	if v := os.Getenv("WHAEH_MODEL"); v != "" {
		cfg.Model = v
	}
	if v := os.Getenv("WHAEH_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("WHAEH_BASE_URL"); v != "" {
		cfg.BaseURL = v
	}

	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv(providerEnv(cfg.Provider))
	}
	return cfg, nil
}

func Save(cfg Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "config.yaml"), data, 0o600)
}

func configDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "whaeh"), nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "whaeh"), nil
}

func configPaths() []string {
	var paths []string
	if dir, err := configDir(); err == nil {
		paths = append(paths, filepath.Join(dir, "config.yaml"))
	}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".whaeh", "config.yaml"))
	}
	return paths
}

func providerEnv(provider string) string {
	switch provider {
	case "gemini":
		return "GEMINI_API_KEY"
	case "openai":
		return "OPENAI_API_KEY"
	default:
		return ""
	}
}
