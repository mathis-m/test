package useful_paths

import (
	"fmt"
	"os"
	"path"
)

type UsefulPaths struct {
	HomeDir        string
	BaseDir        string
	ScriptDir      string
	SystemdUserDir string

	KubeadmConfig string

	Scripts struct {
		Rootlesskit      string
		Containerd       string
		BootstrapCluster string
		KubeadmPrepare   string
		KubeadmFinish    string
		Nsenter          string
		Kubelet          string
	}
	Services struct {
		Rootlesskit string
		Kubelet     string
	}
}

func ConstructUsefulPaths() (*UsefulPaths, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("unable to get users home directory: %w", err)
	}

	baseDir := path.Join(homeDir, RelativePathBaseDir)
	scriptDir := path.Join(baseDir, RelativePathScriptsDir)
	systemdUserDir := path.Join(homeDir, RelativePathSystemdUserDir)

	paths := &UsefulPaths{
		HomeDir:        homeDir,
		BaseDir:        baseDir,
		ScriptDir:      scriptDir,
		SystemdUserDir: systemdUserDir,
	}

	paths.KubeadmConfig = path.Join(baseDir, RelativePathKubeadmConfigFile)

	paths.Scripts.Rootlesskit = path.Join(scriptDir, ScriptsRootlessctl)
	paths.Scripts.Containerd = path.Join(scriptDir, ScriptsContainerd)
	paths.Scripts.BootstrapCluster = path.Join(scriptDir, ScriptsBootstrapCluster)
	paths.Scripts.Nsenter = path.Join(scriptDir, ScriptsNsenter)
	paths.Scripts.Kubelet = path.Join(scriptDir, ScriptsKubelet)
	paths.Scripts.KubeadmPrepare = path.Join(scriptDir, ScriptsKubeadmPrepare)
	paths.Scripts.KubeadmFinish = path.Join(scriptDir, ScriptsKubeadmFinish)

	paths.Services.Rootlesskit = path.Join(systemdUserDir, ServicesRootlesskit)
	paths.Services.Kubelet = path.Join(systemdUserDir, ServicesKubelet)

	return paths, nil
}
