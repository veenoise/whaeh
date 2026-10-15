package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/veenoise/whaeh/pkg/ai"
	"github.com/veenoise/whaeh/pkg/config"
	"github.com/veenoise/whaeh/pkg/input"
	"github.com/veenoise/whaeh/pkg/output"
	"golang.org/x/term"
)

var (
	filePath     string
	model        string
	provider     string
	systemPrompt string
	raw          bool
)

var rootCmd = &cobra.Command{
	Use:   "whaeh [query]",
	Short: "Ask an AI about stdin or a file",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if provider != "" {
			cfg.Provider = provider
		}
		if model != "" {
			cfg.Model = model
		}
		if systemPrompt != "" {
			cfg.SystemPrompt = systemPrompt
		}
		if cfg.Provider == "" {
			return fmt.Errorf("no provider configured; use --provider or `whaeh config set provider <name>`")
		}
		contextText, err := input.ReadContext(filePath)
		if err != nil {
			return fmt.Errorf("read context: %w", err)
		}
		p, err := ai.NewProvider(cfg.Provider, cfg.APIKey, cfg.BaseURL)
		if err != nil {
			return err
		}
		var response bytes.Buffer
		var processing *loader
		if term.IsTerminal(int(os.Stderr.Fd())) {
			processing = startLoader(os.Stderr)
		}
		err = p.StreamCompletion(context.Background(), ai.ProviderRequest{
			Prompt:       argsString(args),
			Context:      contextText,
			Model:        cfg.Model,
			SystemPrompt: cfg.SystemPrompt,
		}, &response)
		if err != nil {
			if processing != nil {
				processing.stopLoader()
			}
			return fmt.Errorf("completion failed: %w", err)
		}
		if processing != nil {
			processing.stopLoader()
		}
		return output.NewRenderer(raw).Render(response.String(), os.Stdout)
	},
}

func init() {
	rootCmd.Flags().StringVarP(&filePath, "file", "f", "", "read context from a file")
	rootCmd.Flags().StringVarP(&model, "model", "m", "", "model override")
	rootCmd.Flags().StringVarP(&provider, "provider", "p", "", "provider override (gemini, openai, ollama)")
	rootCmd.Flags().StringVarP(&systemPrompt, "system", "s", "", "system prompt override")
	rootCmd.Flags().BoolVar(&raw, "raw", false, "disable terminal markdown rendering")
	rootCmd.AddCommand(configCmd, versionCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "whaeh:", err)
		os.Exit(1)
	}
}

func argsString(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		result += arg
	}
	return result
}
