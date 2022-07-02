package installer

import (
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/useful_paths"
)

func Install() error {
	paths, err := useful_paths.ConstructUsefulPaths()
	if err != nil {
		return fmt.Errorf("unable to construct paths: %w", err)
	}

	if err := createBaseDir(paths); err != nil {
		return err
	}

	if err := copyScripts(paths); err != nil {
		return err
	}

	if err := createSystemdServices(paths); err != nil {
		return err
	}

	if err := createKubeadmConfig(paths); err != nil {
		return err
	}

	return nil
}
