package cluster_join

import (
	"github.com/s-bauer/slurm-k8s/internal/util"
)

// Initialize uses viper options:
//    - restart
func Initialize() error {
	isChild := util.IsInNamespace()

	if !isChild {
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
