package slurm

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"unsafe"
)

func FixPathEnvironmentVariable(spank unsafe.Pointer) error {
	pathVar, err := GetSlurmEnvVar(spank, "PATH")
	if err != nil {
		return errors.New(fmt.Sprint("Unable to get PATH env var from slurm:", err))
	}

	log.Info("PATH from slurm is:", pathVar, ", from os:", os.Getenv("PATH"))

	if err = os.Setenv("PATH", pathVar); err != nil {
		return errors.New(fmt.Sprint("Unable to set PATH os env:", err))
	}

	return nil
}

func FixEnvironmentVariables(spank unsafe.Pointer, variables []string) error {
	for _, name := range variables {
		slurmValue, err := GetSlurmEnvVar(spank, name)
		if err != nil {
			log.Warnf("unable to get env var %q from slurm: %v", name, err)
			continue
		}

		osValue := os.Getenv(name)

		log.Infof("variable: %q, from slurm: %q, from os: %q", name, slurmValue, osValue)

		if err := os.Setenv(name, slurmValue); err != nil {
			return err
		}
	}

	return nil
}
