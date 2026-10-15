package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/veenoise/whaeh/pkg/config"
	"golang.org/x/term"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage whaeh configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return cmd.Help()
		}
		return runInteractiveConfig(cmd)
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [provider|model|api_key|base_url|system_prompt] [value]",
	Args:  cobra.ExactArgs(2),
	Short: "Set a configuration value",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		switch args[0] {
		case "provider":
			cfg.Provider = args[1]
		case "model":
			cfg.Model = args[1]
		case "api_key":
			cfg.APIKey = args[1]
		case "base_url":
			cfg.BaseURL = args[1]
		case "system_prompt":
			cfg.SystemPrompt = args[1]
		default:
			return fmt.Errorf("unknown setting %q", args[0])
		}
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "configuration saved")
		return nil
	},
}

func runInteractiveConfig(cmd *cobra.Command) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Fprintln(cmd.OutOrStdout(), "\nConfigure whaeh:")
	fmt.Fprintln(cmd.OutOrStdout(), "1) ollama")
	fmt.Fprintln(cmd.OutOrStdout(), "2) openai")
	fmt.Fprintln(cmd.OutOrStdout(), "3) gemini")

	currentProviderIdx := "1"
	switch cfg.Provider {
	case "openai":
		currentProviderIdx = "2"
	case "gemini":
		currentProviderIdx = "3"
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Select provider [%s]: ", currentProviderIdx)
	providerChoice, _ := reader.ReadString('\n')
	providerChoice = strings.TrimSpace(providerChoice)
	if providerChoice == "" {
		providerChoice = currentProviderIdx
	}

	oldProvider := cfg.Provider
	switch providerChoice {
	case "1", "ollama":
		cfg.Provider = "ollama"
	case "2", "openai":
		cfg.Provider = "openai"
	case "3", "gemini":
		cfg.Provider = "gemini"
	default:
		return fmt.Errorf("invalid provider choice: %s", providerChoice)
	}

	defaultModel := cfg.Model
	if cfg.Provider != oldProvider {
		switch cfg.Provider {
		case "ollama":
			defaultModel = "llama3.2"
		case "openai":
			defaultModel = "gpt-4o-mini"
		case "gemini":
			defaultModel = "gemini-2.5-flash"
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Model [%s]: ", defaultModel)
	modelInput, _ := reader.ReadString('\n')
	modelInput = strings.TrimSpace(modelInput)
	if modelInput != "" {
		cfg.Model = modelInput
	} else {
		cfg.Model = defaultModel
	}

	basePrompt := "Base URL (optional, press Enter to skip)"
	if cfg.BaseURL != "" {
		basePrompt += fmt.Sprintf(" [%s]", cfg.BaseURL)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s: ", basePrompt)
	baseInput, _ := reader.ReadString('\n')
	baseInput = strings.TrimSpace(baseInput)
	if baseInput != "" {
		cfg.BaseURL = baseInput
	}

	if cfg.Provider == "openai" || cfg.Provider == "gemini" {
		keyPrompt := "API Key"
		if cfg.APIKey != "" {
			masked := cfg.APIKey
			if len(masked) > 4 {
				masked = "••••" + masked[len(masked)-4:]
			} else {
				masked = "••••"
			}
			keyPrompt += fmt.Sprintf(" [keep existing: %s]", masked)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s: ", keyPrompt)
		keyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(cmd.OutOrStdout())
		if err == nil {
			keyInput := strings.TrimSpace(string(keyBytes))
			if keyInput != "" {
				cfg.APIKey = keyInput
			}
		}
	}

	fmt.Fprint(cmd.OutOrStdout(), "Modify system prompt? [y/N]: ")
	modPrompt, _ := reader.ReadString('\n')
	modPrompt = strings.TrimSpace(strings.ToLower(modPrompt))
	if modPrompt == "y" || modPrompt == "yes" {
		updatedPrompt, err := editInEditor(cfg.SystemPrompt, reader, cmd)
		if err != nil {
			return fmt.Errorf("edit system prompt: %w", err)
		}
		cfg.SystemPrompt = updatedPrompt
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "\n✓ Configuration saved successfully.")
	return nil
}

func editInEditor(currentPrompt string, reader *bufio.Reader, cmd *cobra.Command) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		if _, err := exec.LookPath("nano"); err == nil {
			editor = "nano"
		} else if _, err := exec.LookPath("vi"); err == nil {
			editor = "vi"
		} else {
			return currentPrompt, fmt.Errorf("no editor found; set $EDITOR")
		}
	}

	tmpFile, err := os.CreateTemp("", "whaeh-system-prompt-*.txt")
	if err != nil {
		return "", err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.WriteString(currentPrompt); err != nil {
		_ = tmpFile.Close()
		return "", err
	}
	_ = tmpFile.Close()

	c := exec.Command(editor, tmpPath)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return "", err
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", err
	}

	result := strings.TrimSpace(string(data))
	if result == "" {
		fmt.Fprint(cmd.OutOrStdout(), "File is empty. Reset to default prompt or leave empty? [D/e]: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToLower(choice))
		if choice == "e" || choice == "empty" {
			return "", nil
		}
		return config.DefaultSystemPrompt, nil
	}

	return result, nil
}

func init() {
	configCmd.AddCommand(configSetCmd)
}
