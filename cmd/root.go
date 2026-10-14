package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"
	"github.com/veenoise/whaeh/pkg/ai"
	"github.com/veenoise/whaeh/pkg/config"
	"github.com/veenoise/whaeh/pkg/input"
	"github.com/veenoise/whaeh/pkg/output"
	"golang.org/x/term"
)

var (
	filePath string
	model    string
	provider string
	raw      bool
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

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		var processing *loader
		if term.IsTerminal(int(os.Stderr.Fd())) {
			processing = startLoader(os.Stderr)
		}
		defer func() {
			if processing != nil {
				processing.stopLoader()
			}
		}()

		isTTY := term.IsTerminal(int(os.Stdout.Fd()))
		var response bytes.Buffer
		var out io.Writer = &response

		if raw || !isTTY {
			out = output.NewStreamWriter(os.Stdout, func() {
				if processing != nil {
					processing.stopLoader()
				}
			})
		}

		err = p.StreamCompletion(ctx, ai.ProviderRequest{
			Prompt:  strings.Join(args, " "),
			Context: contextText,
			Model:   cfg.Model,
		}, out)
		if err != nil {
			return fmt.Errorf("completion failed: %w", err)
		}

		if processing != nil {
			processing.stopLoader()
		}

		if !raw && isTTY {
			return output.Render(response.String(), os.Stdout)
		}
		return nil
	},
}

func init() {
	rootCmd.Flags().StringVarP(&filePath, "file", "f", "", "read context from a file")
	rootCmd.Flags().StringVarP(&model, "model", "m", "", "model override")
	rootCmd.Flags().StringVarP(&provider, "provider", "p", "", "provider override (gemini, openai, ollama)")
	rootCmd.Flags().BoolVar(&raw, "raw", false, "disable terminal markdown rendering")
	rootCmd.AddCommand(configCmd, versionCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "whaeh:", err)
		os.Exit(1)
	}
}
