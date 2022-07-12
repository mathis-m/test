package spank_remote

import "C"
import (
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/slurm"
	"github.com/s-bauer/slurm-k8s/internal/util"
	"unsafe"
)

func Exit(spank unsafe.Pointer) error {
	if err := slurm.FixPathEnvironmentVariable(spank); err != nil {
		return fmt.Errorf("util.FixPathEnvironmentVariable: %w", err)
	}

	// prepare
	jobUser, err := slurm.GetJobUser(spank)
	if err != nil {
		return fmt.Errorf("slurm.GetJobUser: %w", err)
	}

	// run uninstall
	cmdResult, err := util.RunCommand(
		"/home/simon/spank-go/bin/bootstrap-uk8s",
		"--verbose",
		"--simple-log",
		fmt.Sprintf("--drop-uid=%v", jobUser.Uid),
		fmt.Sprintf("--drop-gid=%v", jobUser.Gid),
		"uninstall",
	)
	if err != nil {
		return fmt.Errorf("uninstall failed: %w", err)
	}
	if cmdResult.ExitCode != 0 {
		return fmt.Errorf("uninstall failed")
	}

	return nil
}
