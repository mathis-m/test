package util

import "C"
import (
	"errors"
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/slurm"
	log "github.com/sirupsen/logrus"
	"os"
	"unsafe"
)

func FixPathEnvironmentVariable(spank unsafe.Pointer) error {
	pathVar, err := slurm.GetSlurmEnvVar(spank, "PATH")
	if err != nil {
		return errors.New(fmt.Sprint("Unable to get PATH env var from slurm:", err))
	}

	log.Info("PATH from slurm is:", pathVar, ", from os:", os.Getenv("PATH"))

	if err = os.Setenv("PATH", pathVar); err != nil {
		return errors.New(fmt.Sprint("Unable to set PATH os env:", err))
	}

	return nil
}
