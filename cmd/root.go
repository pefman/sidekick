package cmd

import (
	"github.com/pefman/sidekick/internal/interactive"
	"github.com/pefman/sidekick/internal/updater"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sidekick",
	Short: "Sidekick - AI-powered code assistant",
	Long: `Sidekick is a CLI tool that uses local LLM (via Ollama) to scan 
your codebase for security vulnerabilities and provide insights.

Run without arguments to launch interactive mode.`,
	Version: updater.Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no subcommand, run interactive mode
		im := interactive.New()
		return im.Run()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(updateCmd)
}
