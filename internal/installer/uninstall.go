package installer

import (
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/useful_paths"
	"github.com/s-bauer/slurm-k8s/internal/util"
	log "github.com/sirupsen/logrus"
	"os"
)

func Uninstall() error {
	paths, err := useful_paths.ConstructUsefulPaths()
	if err != nil {
		return fmt.Errorf("unable to construct paths: %w", err)
	}

	// Make sure services are stopped
	_ = util.ReloadSystemdDaemon()

	log.WithFields(log.Fields{"service": useful_paths.ServicesRootlesskit}).Info("Stopping service")
	service := util.Service{Name: useful_paths.ServicesRootlesskit}
	_ = service.Stop()

	log.WithFields(log.Fields{"service": useful_paths.ServicesKubelet}).Info("Stopping service")
	service = util.Service{Name: useful_paths.ServicesKubelet}
	_ = service.Stop()

	removeIgnoreError(paths.Services.Rootlesskit)
	removeIgnoreError(paths.Services.Kubelet)
	removeIgnoreError(paths.BaseDir)

	return nil
}

func removeIgnoreError(name string) {
	logger := log.WithFields(log.Fields{
		"path": name,
	})

	info, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warn("no such file or directory")
			return
		}

		logger.Warn("unable to stat")
		return
	}

	if info.IsDir() {
		if err := os.RemoveAll(name); err != nil {
			logger.Warn("unable to delete directory")
			return
		}

		logger.Info("deleted successfully")
		return
	}

	if err := os.Remove(name); err != nil {
		logger.Warn("unable to delete file")
		return
	}

	logger.Info("deleted successfully")
}
