package cluster_join

import (
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func childInitialize() error {
	cmdResult, err := util.RunCommand(
		"kubeadm",
		"join",
		viper.GetString("api-server-endpoint"),
		"--token",
		viper.GetString("token"),
		"--discovery-token-ca-cert-hash",
		viper.GetString("discovery-token-ca-cert-hash"),
	)
	if err != nil {
		return fmt.Errorf("unable to execute kubeadm join: %w", err)
	}
	if cmdResult.ExitCode != 0 {
		return fmt.Errorf("kubeadm join failed with exit code %v", cmdResult.ExitCode)
	}
	log.Infof("kubeadm join succeeded")

	if err := util.WriteResult(util.ChildResult{}); err != nil {
		return fmt.Errorf("unable to write result: %w", err)
	}

	return nil
}
