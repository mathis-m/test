package cluster_join

import (
	"github.com/s-bauer/slurm-k8s/internal/useful_paths"
	"github.com/s-bauer/slurm-k8s/internal/util"
	log "github.com/sirupsen/logrus"
)

func parentInitialize() error {
	kubeletService := util.Service{Name: useful_paths.ServicesKubelet}
	rootlessService := util.Service{Name: useful_paths.ServicesRootlesskit}

	_ = rootlessService.Stop()
	_ = kubeletService.Stop()

	_ = rootlessService.ResetFailed()
	_ = kubeletService.ResetFailed()

	if err := rootlessService.Start(); err != nil {
		return err
	}
	log.Infof("started rootlesskit service")

	if err := kubeletService.Start(); err != nil {
		return err
	}
	log.Infof("started kubelet service")

	_, err := util.ReexecuteInNamespace(nil)
	if err != nil {
		return err
	}

	return nil
}
