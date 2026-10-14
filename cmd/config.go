package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/veenoise/whaeh/pkg/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage whaeh configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set [provider|model|api_key|base_url] [value]",
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

func init() {
	configCmd.AddCommand(configSetCmd)
}
