package kube

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func AnnotateFlannelPublicIp() error {
	// Create go client
	kubeconfig := "/etc/kubernetes/admin.conf"
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return fmt.Errorf("unable to build kube client config: %w", err)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("unable to build kube client: %w", err)
	}

	// Annotate flannel public ip
	nodes, err := clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("unable list nodes: %w", err)
	}

	for _, node := range nodes.Items {
		var internalIp string
		for _, address := range node.Status.Addresses {
			if address.Type == corev1.NodeInternalIP {
				internalIp = address.Address
			}
		}

		if internalIp == "" {
			log.Warnf("unable to find InternalIP for node %v", node.Name)
		}

		node.Annotations["flannel.alpha.coreos.com/public-ip-overwrite"] = internalIp
		if _, err := clientSet.CoreV1().Nodes().Update(context.TODO(), &node, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("unable to add public-ip-overwrite annotation to node %v: %w", node.Name, err)
		}

		log.Infof("added flannel public ip annotation (%v) to node %v", internalIp, node.Name)
	}

	return nil
}
