package installer

import (
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/useful_paths"
	"github.com/s-bauer/slurm-k8s/internal/util"
	"github.com/spf13/viper"
	"strings"
)

func Install() error {
	force := viper.GetBool("force")
	taints := viper.GetString("taints")
	labels := viper.GetString("labels")

	var args []string
	if labels != "" {
		args = append(args, fmt.Sprintf("--node-labels=userNodeFor=%s", labels))
	}
	if taints != "" {
		args = append(args, fmt.Sprintf("--register-with-taints=%s", taints))
	}

	extraArgs := ""
	if len(args) > 0 {
		extraArgs = fmt.Sprintf("Environment=\"EXTRA_ARGS=%s\"", strings.Join(args[:], " "))
	}

	if force {
		stopServices()
	}

	paths, err := useful_paths.ConstructUsefulPaths()
	if err != nil {
		return fmt.Errorf("unable to construct paths: %w", err)
	}

	if err := createBaseDir(paths, force); err != nil {
		return err
	}

	if err := copyScripts(paths, force); err != nil {
		return err
	}

	if err := createSystemdServices(paths, force, extraArgs); err != nil {
		return err
	}

	if err := createKubeadmConfig(paths, force); err != nil {
		return err
	}

	return nil
}

func stopServices() {
	rootlessKit := &util.Service{Name: useful_paths.ServicesRootlesskit}
	kubelet := &util.Service{Name: useful_paths.ServicesKubelet}

	_ = kubelet.Stop()
	_ = rootlessKit.Stop()
}
