package spank_remote

import "C"
import (
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/kube"
	"github.com/s-bauer/slurm-k8s/internal/slurm"
	"github.com/s-bauer/slurm-k8s/internal/util"
	"unsafe"
)

func Init(spank unsafe.Pointer) error {
	if err := util.FixPathEnvironmentVariable(spank); err != nil {
		return fmt.Errorf("util.FixPathEnvironmentVariable: %w", err)
	}

	jobUser, err := slurm.GetJobUser(spank)
	if err != nil {
		return fmt.Errorf("slurm.GetJobUser: %w", err)
	}

	kubeCluster := kube.NewKubernetesCluster(jobUser)
	if err = kubeCluster.Initialize(); err != nil {
		return fmt.Errorf("kubeCluster.Initialize: %w", err)
	}

	if err = kubeCluster.InitializeAdminUser(); err != nil {
		return fmt.Errorf("kubeCluster.InitializeAdminUser: %w", err)
	}

	return nil
}
