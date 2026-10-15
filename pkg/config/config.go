package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const DefaultSystemPrompt = `You are a Unix CLI assistant embedded in a command-line pipeline.
- Be direct and concise. Skip pleasantries, introductions, and filler.
- When analyzing code, diffs, or logs: pinpoint root causes, errors, and supply exact, minimal code or commands.
- For general questions: provide accurate, direct answers without unnecessary fluff.
- Format output with clean, readable Markdown for terminal display.`

type Config struct {
	Provider     string `mapstructure:"provider"`
	Model        string `mapstructure:"model"`
	APIKey       string `mapstructure:"api_key"`
	BaseURL      string `mapstructure:"base_url"`
	SystemPrompt string `mapstructure:"system_prompt"`
}

func ConfigPath(override ...string) (string, error) {
	if len(override) > 0 && override[0] != "" {
		return override[0], nil
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "whaeh", "config.yaml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".whaeh", "config.yaml"), nil
}

func Load(override ...string) (Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("WHAEH")
	viper.AutomaticEnv()
	viper.SetDefault("provider", "ollama")
	viper.SetDefault("model", "llama3.2")
	viper.SetDefault("base_url", "")
	viper.SetDefault("system_prompt", DefaultSystemPrompt)
	if len(override) > 0 && override[0] != "" {
		viper.SetConfigFile(override[0])
	} else {
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			viper.AddConfigPath(filepath.Join(xdg, "whaeh"))
		}
		if home, err := os.UserHomeDir(); err == nil {
			viper.AddConfigPath(filepath.Join(home, ".whaeh"))
		}
	}
	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !isConfigNotFound(err, &notFound) && !errors.Is(err, os.ErrNotExist) {
			return Config{}, fmt.Errorf("read config: %w", err)
		}
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv(providerEnv(cfg.Provider))
	}
	return cfg, nil
}

func Save(cfg Config, override ...string) (string, error) {
	target, err := ConfigPath(override...)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return "", err
	}
	viper.Set("provider", cfg.Provider)
	viper.Set("model", cfg.Model)
	viper.Set("api_key", cfg.APIKey)
	viper.Set("base_url", cfg.BaseURL)
	viper.Set("system_prompt", cfg.SystemPrompt)
	return target, viper.WriteConfigAs(target)
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

func isConfigNotFound(err error, target *viper.ConfigFileNotFoundError) bool {
	return errors.As(err, target)
}
