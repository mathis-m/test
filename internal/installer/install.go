package installer

import (
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/useful_paths"
	"github.com/s-bauer/slurm-k8s/internal/util"
	"github.com/spf13/viper"
)

func Install() error {
	force := viper.GetBool("force")
	if force {
		stopServices()
	}

	paths, err := useful_paths.ConstructUsefulPaths()
	if err != nil {
		return fmt.Errorf("unable to construct paths: %w", err)
	}

	if err := createBaseDir(paths, force); err != nil {
		return err
	}

	if err := copyScripts(paths, force); err != nil {
		return err
	}

	if err := createSystemdServices(paths, force); err != nil {
		return err
	}

	if err := createKubeadmConfig(paths, force); err != nil {
		return err
	}

	return nil
}

func stopServices() {
	rootlessKit := &util.Service{Name: useful_paths.ServicesRootlesskit}
	kubelet := &util.Service{Name: useful_paths.ServicesKubelet}

	_ = kubelet.Stop()
	_ = rootlessKit.Stop()
}