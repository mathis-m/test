package cmd

import (
	"github.com/s-bauer/slurm-k8s/internal/cluster_initialize"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// initCmd represents the start command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Start the kubernetes cluster",
	PreRun: func(cmd *cobra.Command, args []string) {
		_ = viper.BindPFlags(cmd.Flags())
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := cluster_initialize.Initialize(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().String("token", "", "the bootstrap token")
}
