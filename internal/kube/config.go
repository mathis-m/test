package kube

import (
	"errors"
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/util"
	log "github.com/sirupsen/logrus"
	"io"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"os/user"
)

const (
	AdminConfigPath        = "/etc/kubernetes/admin.conf"
	UserConfigRelativePath = ".kube/config"
	KubeadmConfigPath      = "/tmp/kubeadm-config"
)

func writeKubeadmConfig() error {
	cmd := fmt.Sprintf(
		"kubectl --kubeconfig %v get cm kubeadm-config -n kube-system -o=jsonpath=\"{.data.ClusterConfiguration}\" > %v",
		AdminConfigPath, KubeadmConfigPath)

	_, err := util.RunProcess("kubectl get cm kubeadm-config", cmd)
	return err
}

func kubeadmCreateUser(jobUser *user.User) (string, error) {
	cmd := fmt.Sprintf(
		"kubeadm kubeconfig user --client-name %v --config %v",
		jobUser.Username, KubeadmConfigPath)

	kubeConfig, err := util.RunProcessGetStdout("kubeadm kubeconfig user", cmd)
	if err != nil {
		return "", err
	}

	return kubeConfig, nil
}

func CreateKubeClient(configString string) (*kubernetes.Clientset, error) {
	clientConfig, err := clientcmd.NewClientConfigFromBytes([]byte(configString))
	if err != nil {
		return nil, err
	}

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return clientSet, nil
}

// deprecated
func copyKubernetesConfig(user *user.User) error {
	// Get Source File
	adminConf, err := os.Open(AdminConfigPath)
	if err != nil {
		return errors.New(fmt.Sprint("Unable to open", AdminConfigPath, ":", err))
	}
	defer func() {
		err := adminConf.Close()
		if err != nil {
			log.Error("Unable to close /etc/kubernetes/admin.conf", err)
		}
	}()

	userKubeDirPath := fmt.Sprintf("%s/.kube", user.HomeDir)
	if _, err := os.Stat(userKubeDirPath); os.IsNotExist(err) {
		if err := os.Mkdir(userKubeDirPath, os.ModePerm); err != nil {
			return errors.New(fmt.Sprint("Unable to create .kube directory:", err))
		}
	}

	userConfigPath := fmt.Sprintf("%s/config", userKubeDirPath)
	if _, err := os.Stat(userConfigPath); !os.IsNotExist(err) {
		if err := os.Remove(userConfigPath); err != nil {
			return errors.New(fmt.Sprint("Unable to delete ~/.kube/config file:", err))
		}
	}

	userConfig, err := os.Create(userConfigPath)
	if err != nil {
		return errors.New(fmt.Sprint("Unable to create ~/.kube/config file:", err))
	}

	if _, err := io.Copy(userConfig, adminConf); err != nil {
		return errors.New(fmt.Sprint("Unable to copy", AdminConfigPath, "to ~/.kube/config:", err))
	}

	return nil
}
