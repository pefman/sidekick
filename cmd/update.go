package cmd

import (
	"fmt"

	"github.com/pefman/sidekick/internal/updater"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update sidekick to the latest version",
	Long:  `Check for updates and install the latest version of sidekick from GitHub releases.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("🔍 Current version: %s\n", updater.Version)
		fmt.Println("🔍 Checking for updates...")

		return updater.Update()
	},
}
