package spank_remote

import "C"
import (
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/util"
	"unsafe"
)

func Exit(spank unsafe.Pointer) error {
	if err := util.FixPathEnvironmentVariable(spank); err != nil {
		return fmt.Errorf("util.FixPathEnvironmentVariable: %w", err)
	}

	_, err := util.RunProcess("kubeadm reset", "kubeadm reset --force")
	if err != nil {
		return fmt.Errorf("util.RunProcess(kubeadm reset): %w", err)
	}

	return nil
}
