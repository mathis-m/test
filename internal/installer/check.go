package installer

import (
	"github.com/s-bauer/slurm-k8s/internal/useful_paths"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
)

func CheckIsInstalled() (bool, error) {
	validInstallation := true

	usefulPaths, err := useful_paths.ConstructUsefulPaths()
	if err != nil {
		return false, err
	}

	foldersToCheck := []string{
		usefulPaths.BaseDir,
		usefulPaths.ScriptDir,
	}

	filesToCheck := []string{
		usefulPaths.Services.Rootlesskit,
		usefulPaths.Services.Kubelet,
		usefulPaths.KubeadmAdminConfig,
	}

	scriptsToCheck, err := getAllScriptPaths()
	if err != nil {
		return false, err
	}

	for _, item := range scriptsToCheck {
		filesToCheck = append(filesToCheck, path.Join(usefulPaths.ScriptDir, item))
	}

	for _, folder := range foldersToCheck {
		folderExists, err := checkFolderExists(folder)
		if err != nil {
			return false, err
		}

		if !folderExists {
			log.Warnf("check installation: folder %q does not exist", folder)
		}

		validInstallation = validInstallation && folderExists
	}

	for _, file := range filesToCheck {
		fileExists, err := checkFileExists(file)
		if err != nil {
			return false, err
		}

		if !fileExists {
			log.Warnf("check installation: file %q does not exist", file)
		}

		validInstallation = validInstallation && fileExists
	}

	return validInstallation, nil
}

func getAllScriptPaths() ([]string, error) {
	items, err := scriptFS.ReadDir(scriptsFSBasePath)
	if err != nil {
		return nil, err
	}

	var scriptPaths []string
	for _, item := range items {
		scriptPaths = append(scriptPaths, item.Name())
	}

	return scriptPaths, nil
}

func checkFolderExists(path string) (bool, error) {
	stats, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return stats.IsDir(), nil
}

func checkFileExists(path string) (bool, error) {
	stats, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return !stats.IsDir(), nil
}
