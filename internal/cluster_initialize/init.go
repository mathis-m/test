package cluster_initialize

import (
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/installer"
	"github.com/s-bauer/slurm-k8s/internal/util"
)

// Initialize uses viper options:
//    - restart
func Initialize() error {
	isChild := util.IsInNamespace()

	if !isChild {
		isInstalled, err := installer.CheckIsInstalled()
		if err != nil {
			return fmt.Errorf("check is installed: %w", err)
		}

		if !isInstalled {
			return fmt.Errorf("installation not found. please run %q before", "bootstrap-uk8s install")
		}

		if err := parentInitialize(); err != nil {
			return err
		}
	} else {
		if err := childInitialize(); err != nil {
			return err
		}
	}

	return nil
}