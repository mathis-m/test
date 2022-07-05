package spank_remote

import "C"
import (
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/slurm"
	"github.com/s-bauer/slurm-k8s/internal/useful_paths"
	"github.com/s-bauer/slurm-k8s/internal/util"
	"unsafe"
)

func Exit(spank unsafe.Pointer) error {
	if err := slurm.FixPathEnvironmentVariable(spank); err != nil {
		return fmt.Errorf("util.FixPathEnvironmentVariable: %w", err)
	}

	// stopServices()

	return nil
}

func stopServices() {
	rootlessKit := &util.Service{Name: useful_paths.ServicesRootlesskit}
	kubelet := &util.Service{Name: useful_paths.ServicesKubelet}

	_ = kubelet.Stop()
	_ = rootlessKit.Stop()
}
