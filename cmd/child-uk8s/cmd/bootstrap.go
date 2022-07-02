package cmd

import (
	"github.com/spf13/cobra"
)

// bootstrapCmd represents the start command
var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstraps the k8s cluster in this namespace",
	Run: func(cmd *cobra.Command, args []string) {
		
	},
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)
}
