package cluster_initialize

import (
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/useful_paths"
	"github.com/s-bauer/slurm-k8s/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func parentInitialize() error {
	rootlessService := &util.Service{Name: useful_paths.ServicesRootlesskit}
	kubeletService := &util.Service{Name: useful_paths.ServicesKubelet}

	usefulPaths, err := useful_paths.ConstructUsefulPaths()
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("reloading systemd user daemon")
	if err := util.ReloadSystemdDaemon(); err != nil {
		return fmt.Errorf("unable to reload systemd daemon: %w", err)
	}
	log.Info("reloaded systemd user daemon")

	if err := util.DeleteFileIfExists(usefulPaths.KubernetesUserConfig); err != nil {
		return fmt.Errorf("unable to delete kubernetes config at %q: %w", usefulPaths.KubernetesUserConfig, err)
	}

	stopServices()

	if err := startService(rootlessService); err != nil {
		return err
	}

	log.Info("child phase #1")

	if _, err := util.ReexecuteInNamespace([]string{"SLURM_K8S_CHILD_PHASE=1"}); err != nil {
		return err
	}

	if err := kubeadmPrepare(usefulPaths); err != nil {
		return err
	}

	if err := startService(kubeletService); err != nil {
		return err
	}

	if err := kubeadmFinish(usefulPaths); err != nil {
		return err
	}

	if err := kubeletService.Restart(); err != nil {
		return err
	}

	bootstrapToken := viper.GetString("token")
	log.Infof("user provided token: %v", bootstrapToken)
	if bootstrapToken != "" {
		if err := kubeadmToken(usefulPaths, bootstrapToken); err != nil {
			return err
		}
	}

	log.Info("child phase #2")

	childResult, err := util.ReexecuteInNamespace([]string{"SLURM_K8S_CHILD_PHASE=2"})
	if err != nil {
		return err
	}

	log.Info("initialize succeeded")

	localIp := util.GetLocalIP()
	log.Infof("Use the following command to join a second computer:")
	log.Infof(
		"\t./bootstrap-uk8s join --api-server-endpoint %v:6443 --token %v --discovery-token-ca-cert-hash sha256:%v --verbose",
		localIp,
		childResult["token"],
		childResult["certHash"],
	)

	return nil
}

func kubeadmPrepare(usefulPaths *useful_paths.UsefulPaths) error {
	log.Info("executing kubeadm-prepare")
	cmdResult, err := util.RunCommand(usefulPaths.Scripts.KubeadmPrepare)
	if err != nil || cmdResult.ExitCode != 0 {
		return fmt.Errorf("failed to kubeadm-prepare: %w", err)
	}

	log.Info("completed kubeadm-prepare")
	return nil
}

func kubeadmFinish(usefulPaths *useful_paths.UsefulPaths) error {
	log.Info("executing kubeadm-finish")
	cmdResult, err := util.RunCommand(usefulPaths.Scripts.KubeadmFinish)
	if err != nil || cmdResult.ExitCode != 0 {
		return fmt.Errorf("failed to kubeadm-finish: %w", err)
	}

	log.Info("completed kubeadm-finish")
	return nil
}

func kubeadmToken(usefulPaths *useful_paths.UsefulPaths, token string) error {
	log.Info("executing kubeadm-token")
	cmdResult, err := util.RunCommand(usefulPaths.Scripts.KubeadmToken, token)
	if err != nil || cmdResult.ExitCode != 0 {
		return fmt.Errorf("failed to kubeadm-token: %w", err)
	}

	log.Info("completed kubeadm-token")
	return nil
}

func startService(service *util.Service) error {
	logger := log.WithFields(log.Fields{"serviceName": service.Name})

	logger.Debug("querying service status")

	status, err := service.Status()
	if err != nil {
		return err
	}

	logger.Debugf("service status is: %v", status)

	if status == util.Active {
		if viper.GetBool("restart") {
			logger.Info("service is running, stopping first...")
			if err := service.Stop(); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("service %q is running, stop first or include --restart flag", service.Name)
		}
	}

	logger.Debug("(re-)starting service")
	if err := service.Start(); err != nil {
		return err
	}

	logger.Info("service started")
	return nil
}

func stopServices() {
	rootlessKit := &util.Service{Name: useful_paths.ServicesRootlesskit}
	kubelet := &util.Service{Name: useful_paths.ServicesKubelet}

	_ = kubelet.Stop()
	_ = rootlessKit.Stop()

	_ = kubelet.ResetFailed()
	_ = rootlessKit.ResetFailed()
}
