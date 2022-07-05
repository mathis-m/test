/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	publicIpOverwriteAnnotation = "flannel.alpha.coreos.com/public-ip-overwrite"
	restartedAtAnnotation       = "kubectl.kubernetes.io/restartedAt"
	needsFlannelRestart         = "slurm-k8s-manager.bauer.link/needs-flannel-restart"
)

var (
	flannelDsNamespacedName = types.NamespacedName{Namespace: "kube-system", Name: "kube-flannel-ds"}
)

//+kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=nodes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=nodes/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;update

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// get node object
	var node corev1.Node
	if err := r.Get(ctx, req.NamespacedName, &node); err != nil {
		logger.Error(err, "unable to get node by name", "name", req.NamespacedName)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// get internal ip
	internalIp := ""
	for _, address := range node.Status.Addresses {
		if address.Type != corev1.NodeInternalIP {
			break
		}

		internalIp = address.Address
		break
	}

	if internalIp == "" {
		return ctrl.Result{}, fmt.Errorf("unable to find internal ip of node")
	}

	// get existing annotation
	existingValue, ok := node.Annotations[publicIpOverwriteAnnotation]
	if !ok || existingValue != internalIp {
		node.Annotations[publicIpOverwriteAnnotation] = internalIp
		node.Annotations[needsFlannelRestart] = "true"

		if err := r.Update(ctx, &node, &client.UpdateOptions{}); err != nil {
			return ctrl.Result{}, err
		}

		logger.Info("updated flannel public ip", "annotation", publicIpOverwriteAnnotation, "public-ip", internalIp)
	}

	_, ok = node.Annotations[needsFlannelRestart]
	if ok {
		// restart flanneld for it to use the new annotation
		if err := r.restartFlannelDs(ctx); err != nil {
			return ctrl.Result{}, err
		}
		logger.Info("restarted flannel-ds", "name", flannelDsNamespacedName)

		// remove needs restart annotation
		if err := r.Get(ctx, req.NamespacedName, &node); err != nil {
			return ctrl.Result{}, err
		}
		delete(node.Annotations, needsFlannelRestart)
		if err := r.Update(ctx, &node, &client.UpdateOptions{}); err != nil {
			if errors.IsConflict(err) {
				return ctrl.Result{RequeueAfter: 1 * time.Second}, nil
			}

			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *NodeReconciler) restartFlannelDs(ctx context.Context) error {
	logger := log.FromContext(ctx)

	// restart flannel-ds
	var flannelDs appsv1.DaemonSet
	if err := r.Get(ctx, flannelDsNamespacedName, &flannelDs); err != nil {
		if errors.IsNotFound(err); err != nil {
			logger.Info("flannel daemonset not found")
			return nil
		}

		return err
	}

	if flannelDs.Spec.Template.ObjectMeta.Annotations == nil {
		flannelDs.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	flannelDs.Spec.Template.ObjectMeta.Annotations[restartedAtAnnotation] = time.Now().Format(time.RFC3339)
	if err := r.Update(ctx, &flannelDs, &client.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}
