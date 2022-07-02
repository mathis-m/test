package kube

import "C"
import (
	"bytes"
	"context"
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/util"
	log "github.com/sirupsen/logrus"
	"io"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"os/user"
	"path/filepath"
)

type KubernetesCluster struct {
	Socket         string
	PodNetworkCidr string
	Token          string
	AdminUser      *user.User
}

func NewKubernetesCluster(adminUser *user.User) *KubernetesCluster {
	return &KubernetesCluster{
		Socket:         "unix:///var/run/containerd/containerd.sock",
		PodNetworkCidr: "10.244.0.0/16",
		Token:          "jd8j3l.jdkla8dasfj90sao",
		AdminUser:      adminUser,
	}
}

func (cluster *KubernetesCluster) createAdminUserConfig() (string, error) {
	if err := writeKubeadmConfig(); err != nil {
		return "", fmt.Errorf("writeKubeadmConfig: %w", err)
	}

	userConfig, err := kubeadmCreateUser(cluster.AdminUser)
	if err != nil {
		return "", fmt.Errorf("createUserConfig: %w", err)
	}

	if err = os.Remove(KubeadmConfigPath); err != nil {
		return "", fmt.Errorf("os.Remove: %w", err)
	}

	return userConfig, nil
}

func (cluster *KubernetesCluster) readSuperUserConfig() (string, error) {
	file, err := os.Open(AdminConfigPath)
	if err != nil {
		return "", fmt.Errorf("os.Open: %w", err)
	}

	buffer := bytes.NewBuffer(nil)

	if _, err = io.Copy(buffer, file); err != nil {
		return "", fmt.Errorf("io.Copy: %w", err)
	}

	return string(buffer.Bytes()), nil

}

func (cluster *KubernetesCluster) createSuperUserClient() (*kubernetes.Clientset, error) {
	configString, err := cluster.readSuperUserConfig()
	if err != nil {
		return nil, fmt.Errorf("cluster.readSuperUserConfig: %w", err)
	}

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

func (cluster *KubernetesCluster) Initialize() error {
	// kubeadm cluster_initialize
	cmd := fmt.Sprintf(
		"kubeadm cluster_initialize --cri-socket=\"%v\" --pod-network-cidr=\"%v\" --token=\"%v\"",
		cluster.Socket,
		cluster.PodNetworkCidr,
		cluster.Token,
	)

	_, err := util.RunProcess("kubeadm cluster_initialize", cmd)
	if err != nil {
		return fmt.Errorf("util.RunProcess(kubeadm cluster_initialize): %w", err)
	}

	// Remove control-plane taint
	cmd = fmt.Sprintf(
		"kubectl --kubeconfig \"%v\" taint nodes --all node-role.kubernetes.io/control-plane- node-role.kubernetes.io/master-",
		AdminConfigPath,
	)

	_, err = util.RunProcess("kubectl taint", cmd)
	if err != nil {
		return fmt.Errorf("util.RunProcess(kubectl taint): %w", err)
	}

	log.Info("kubeadm cluster_initialize finished")
	return nil
}

func (cluster *KubernetesCluster) InitializeAdminUser() error {
	userConfig, err := cluster.createAdminUserConfig()
	if err != nil {
		return fmt.Errorf("createUserConfig: %w", err)
	}

	if err = util.WriteStringToFile(userConfig, filepath.Join(cluster.AdminUser.HomeDir, UserConfigRelativePath)); err != nil {
		return fmt.Errorf("util.WriteStringToFile: %w", err)
	}

	adminClient, err := cluster.createSuperUserClient()
	if err != nil {
		return fmt.Errorf("cluster.createSuperUserClient: %w", err)
	}

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "slurm-job-user-admin",
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup: rbacv1.GroupName,
				Kind:     rbacv1.UserKind,
				Name:     cluster.AdminUser.Username,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
	}

	clusterRoleBindingClient := adminClient.RbacV1().ClusterRoleBindings()
	if _, err := clusterRoleBindingClient.Create(context.TODO(), clusterRoleBinding, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("ClusterRoleBindings.Create: %w", err)
	}

	return nil
}
