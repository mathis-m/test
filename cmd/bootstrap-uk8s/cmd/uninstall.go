package cmd

import (
	"github.com/s-bauer/slurm-k8s/internal/installer"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// uninstallCmd represents the start command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Best effort tool to remove installed components",
	Run: func(cmd *cobra.Command, args []string) {
		if err := installer.Uninstall(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
