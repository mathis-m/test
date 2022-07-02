package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().BoolP("restart", "r", false, "restarts the service if it's already running")

	_ = viper.BindPFlags(startCmd.Flags())
}
