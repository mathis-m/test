package cluster_initialize

import (
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/installer"
	"github.com/s-bauer/slurm-k8s/internal/util"
	"os"
	"strconv"
)

// Initialize uses viper options:
//   - restart
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
		phase, err := strconv.Atoi(os.Getenv("SLURM_K8S_CHILD_PHASE"))
		if err != nil {
			return fmt.Errorf("unable to determine child phase")
		}

		switch phase {
		case 1:
			if err := childPhase1(); err != nil {
				return err
			}
		case 2:
			if err := childPhase2(); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown child phase: %v", phase)
		}

	}

	return nil
}
