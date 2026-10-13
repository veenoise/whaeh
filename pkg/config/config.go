package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Provider string `mapstructure:"provider"`
	Model    string `mapstructure:"model"`
	APIKey   string `mapstructure:"api_key"`
	BaseURL  string `mapstructure:"base_url"`
}

func Load() (Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("WHAEH")
	viper.AutomaticEnv()
	viper.SetDefault("provider", "ollama")
	viper.SetDefault("model", "llama3.2")
	viper.SetDefault("base_url", "")
	viper.AddConfigPath(filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "whaeh"))
	if home, err := os.UserHomeDir(); err == nil {
		viper.AddConfigPath(filepath.Join(home, ".whaeh"))
	}
	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !isConfigNotFound(err, &notFound) {
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

func Save(cfg Config) error {
	viper.Set("provider", cfg.Provider)
	viper.Set("model", cfg.Model)
	viper.Set("api_key", cfg.APIKey)
	viper.Set("base_url", cfg.BaseURL)
	path := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "whaeh")
	if path == filepath.Join("", "whaeh") {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path = filepath.Join(home, ".whaeh")
	}
	if err := os.MkdirAll(path, 0o700); err != nil {
		return err
	}
	return viper.WriteConfigAs(filepath.Join(path, "config.yaml"))
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
