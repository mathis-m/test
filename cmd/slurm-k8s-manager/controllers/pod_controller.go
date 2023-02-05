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
	"github.com/go-logr/logr"
	v1 "github.com/s-bauer/slurm-k8s/cmd/slurm-k8s-manager/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	finalizerName           = "pod.mathis.me/finalizer"
	impersonationRequestKey = "impersonateFor"
)

// PodReconciler reconciles a Node object
type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core,resources=pods,verbs=create;delete
//+kubebuilder:rbac:groups=core,resources=pods/status,verbs=update;patch
//+kubebuilder:rbac:groups=core,resources=pods/finalizers,verbs=update

func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		logger.Error(err, "unable to fetch pod")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	logger.Info("Checking annotation")
	uid, needsImpersonation := pod.Annotations[impersonationRequestKey]
	if needsImpersonation {
		hasFinalizer := contains(pod.Finalizers, finalizerName)
		if pod.ObjectMeta.DeletionTimestamp.IsZero() {

			logger.Info("Ensuring user node exists")
			err := r.createUserNodeIfNeeded(ctx, uid, logger, req.NamespacedName)

			if err != nil {
				logger.Error(err, "unable to list user node resources", "namespace", req.NamespacedName)
				return ctrl.Result{}, err
			}
			// The object is not being deleted, so if it does not have our finalizer,
			// then lets add the finalizer and update the object. This is equivalent
			// registering our finalizer.
			if !hasFinalizer {
				pod.Finalizers = append(pod.Finalizers, finalizerName)
				if err := r.Update(ctx, &pod); err != nil {
					return ctrl.Result{}, err
				}
			}
		} else {
			logger.Info("Pod deletion")
			// The object is being deleted
			if hasFinalizer {
				logger.Info("Checking if user node can be removed")
				var allPodsOnNode corev1.PodList
				err := r.List(ctx, &allPodsOnNode, client.MatchingFields{"spec.nodeName": pod.Spec.NodeName})
				if err != nil {
					logger.Info("unable to get actual node", pod.Spec.NodeName, err)
					removeFinalizer(pod)
					return ctrl.Result{}, err
				}

				hasMorePodsOnNode := false
				for _, item := range allPodsOnNode.Items {
					_, isImpersonatedPod := item.Annotations[impersonationRequestKey]
					if item.UID != pod.UID && isImpersonatedPod {
						hasMorePodsOnNode = true
						break
					}
				}

				if !hasMorePodsOnNode {
					var userNodeList v1.UserNodeList
					err := r.List(ctx, &userNodeList)
					if err != nil {
						logger.Error(err, "unable to list user node resources")
						removeFinalizer(pod)
						return ctrl.Result{}, err
					}

					userNodeOfUID := userNodeList.GetUserNodeByUID(uid)
					if userNodeOfUID != nil {
						err = r.Delete(ctx, userNodeOfUID)
						if err != nil || userNodeOfUID == nil {
							logger.Error(err, "unable to delete user node", "node", userNodeOfUID)
							removeFinalizer(pod)
							return ctrl.Result{}, err
						}
					} else {
						logger.Info("user node does not exist anymore no need to finalize")
					}
				}

				removeFinalizer(pod)
				if err := r.Update(ctx, &pod); err != nil {
					return ctrl.Result{}, err
				}
			}

			// Stop reconciliation as the item is being deleted
			return ctrl.Result{}, nil
		}
	}
	return ctrl.Result{}, nil
}

func (r *PodReconciler) createUserNodeIfNeeded(ctx context.Context, uid string, logger logr.Logger, nameSpaceName types.NamespacedName) error {
	var userNodeList v1.UserNodeList
	err := r.List(ctx, &userNodeList)
	if err != nil {
		logger.Error(err, "unable to list user node resources", "namespace", nameSpaceName)
		return err
	}

	userNodeOfUID := userNodeList.GetUserNodeByUID(uid)

	if userNodeOfUID != nil {
		logger.Info("user node found no need to create one", "namespace", nameSpaceName)
		return nil
	}

	logger.Error(err, "creating user node", "uid", uid)

	userNodeOfUID = &v1.UserNode{
		Spec: v1.UserNodeSpec{
			UserId: uid,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      fmt.Sprintf("user-node-%s", uid),
		},
	}

	return r.Create(ctx, userNodeOfUID)
}

func removeFinalizer(pod corev1.Pod) {
	for i, name := range pod.Finalizers {
		if name == finalizerName {
			pod.Finalizers = append(pod.Finalizers[:i], pod.Finalizers[i+1:]...)
			break
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	_ = mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Pod{}, "spec.nodeName", func(o client.Object) []string {
		return []string{o.(*corev1.Pod).Spec.NodeName}
	})
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
