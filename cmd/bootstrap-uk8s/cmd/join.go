package cmd

import (
	"github.com/s-bauer/slurm-k8s/internal/cluster_join"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// joinCmd represents the start command
var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "Joins a node to the kubernetes cluster",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cluster_join.Initialize(); err != nil {
			log.Fatalf("failed to join cluster: %v", err)
		}
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		_ = viper.BindPFlags(cmd.Flags())
	},
}

func init() {
	rootCmd.AddCommand(joinCmd)

	joinCmd.Flags().String("token", "", "The token to join the cluster")
	joinCmd.Flags().String("discovery-token-ca-cert-hash", "", "The certificate thumbprint for validation")
	joinCmd.Flags().String("api-server-endpoint", "", "The endpoint of the API server to connect to")

	_ = joinCmd.MarkFlagRequired("token")
	_ = joinCmd.MarkFlagRequired("discovery-token-ca-cert-hash")
	_ = joinCmd.MarkFlagRequired("api-server-endpoint")
}
