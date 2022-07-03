package cluster_initialize

import (
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/kube"
	"github.com/s-bauer/slurm-k8s/internal/util"
	log "github.com/sirupsen/logrus"
	"os"
)

func childInitialize() error {
	if err := os.Setenv("KUBECONFIG", "/etc/kubernetes/admin.conf"); err != nil {
		return fmt.Errorf("unable to set KUBECONFIG env var: %w", err)
	}

	// Annotate flannel public ip
	if err := kube.AnnotateFlannelPublicIp(); err != nil {
		return err
	}

	// Install flannel
	cmdResult, err := util.RunCommand("kubectl", "apply", "-f", "https://raw.githubusercontent.com/flannel-io/flannel/master/Documentation/kube-flannel.yml")
	if err != nil {
		return fmt.Errorf("unable to execute kubectl: %w", err)
	}
	if cmdResult.ExitCode != 0 {
		return fmt.Errorf("kubectl failed with exit code %v", cmdResult.ExitCode)
	}
	log.Infof("installed flannel")

	return nil
}
