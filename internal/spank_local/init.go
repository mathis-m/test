package spank_local

import "C"
import (
	"fmt"
	"github.com/spf13/viper"
	bootstraputil "k8s.io/cluster-bootstrap/token/util"
	"os"
	"unsafe"
)

func Init(spank unsafe.Pointer) error {
	initCluster := viper.GetBool("k8s-init-cluster")

	if initCluster {
		if err := runInitCluster(); err != nil {
			return err
		}
	}

	return nil
}

func runInitCluster() error {
	bootstrapToken, err := bootstraputil.GenerateBootstrapToken()
	if err != nil {
		return fmt.Errorf("unable to generate bootstrap token: %w", err)
	}

	if err := os.Setenv("SLURM_K8S_BOOTSTRAP_TOKEN", bootstrapToken); err != nil {
		return fmt.Errorf("unable to set environment variable: %w", err)
	}

	return nil
}
