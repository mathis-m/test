package cluster_initialize

import (
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/useful_paths"
	"github.com/s-bauer/slurm-k8s/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Initialize uses viper options:
//    - restart
func Initialize() error {
	rootlessService := &util.Service{Name: useful_paths.ServicesRootlesskit}
	kubeletService := &util.Service{Name: useful_paths.ServicesKubelet}

	usefulPaths, err := useful_paths.ConstructUsefulPaths()
	if err != nil {
		log.Fatal(err)
	}

	if err := util.ReloadSystemdDaemon(); err != nil {
		return fmt.Errorf("unable to reload systemd daemon: %w", err)
	}

	if err := startService(rootlessService); err != nil {
		return err
	}

	if err := kubeadmPrepare(usefulPaths); err != nil {
		return err
	}

	if err := startService(kubeletService); err != nil {
		return err
	}

	if err := kubeadmFinish(usefulPaths); err != nil {
		return err
	}

	if err := kubeletService.Restart(); err != nil {
		return err
	}

	log.Info("finished bootstrapping")
	return nil
}

func kubeadmPrepare(usefulPaths *useful_paths.UsefulPaths) error {
	log.Info("executing kubeadm-prepare")
	cmdResult, err := util.RunCommand(usefulPaths.Scripts.KubeadmPrepare)
	if err != nil || cmdResult.ExitCode != 0 {
		return fmt.Errorf("failed to kubeadm-prepare: %w", err)
	}

	return nil
}

func kubeadmFinish(usefulPaths *useful_paths.UsefulPaths) error {
	log.Info("executing kubeadm-prepare")
	cmdResult, err := util.RunCommand(usefulPaths.Scripts.KubeadmFinish)
	if err != nil || cmdResult.ExitCode != 0 {
		return fmt.Errorf("failed to kubeadm-finish: %w", err)
	}

	return nil
}

func startService(service *util.Service) error {
	status, err := service.Status()
	if err != nil {
		return err
	}

	if status == util.Active {
		if viper.GetBool("restart") {
			log.Infof("service %q is running, stopping first...", service.Name)
			if err := service.Stop(); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("service %q is running, stop first or include --restart flag", service.Name)
		}
	}

	if err := service.Start(); err != nil {
		return err
	}

	log.Infof("started service %q", service.Name)
	return nil
}
