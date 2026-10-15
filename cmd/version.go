package cmd

import "github.com/spf13/cobra"

const version = "0.2.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(version)
	},
}
