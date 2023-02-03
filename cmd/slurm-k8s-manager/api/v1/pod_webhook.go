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

package v1

import (
	"context"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	impersonationRequestKey = "impersonateFor"
	userNodeKey             = "userNodeFor"
)

//+kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=fail,sideEffects=None,groups=*,resources=pods,verbs=create;update,versions=v1;v1beta1,name=mpod.kb.io,admissionReviewVersions=v1

type podAnnotator struct {
	Client  client.Client
	decoder *admission.Decoder
}

func NewPodValidator(c client.Client) admission.Handler {
	return &podAnnotator{Client: c}
}

func (a *podAnnotator) Handle(ctx context.Context, req admission.Request) admission.Response {
	logger := log.FromContext(ctx)
	pod := &corev1.Pod{}
	err := a.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}

	uid, needsImpersonation := pod.Annotations[impersonationRequestKey]

	logger.Info("Checking Pod for annotation")

	if needsImpersonation {
		logger.Info("annotation exists for user", "uid", uid)

		createOrUpdateNodeSelector(pod, logger, uid)
		createOrUpdateTolerationForUid(pod, uid, logger)
	}

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

func createOrUpdateNodeSelector(pod *corev1.Pod, logger logr.Logger, uid string) {
	if pod.Spec.NodeSelector == nil {
		pod.Spec.NodeSelector = map[string]string{}
	}

	logger.Info("Adding new nodeSelector to pod")
	pod.Spec.NodeSelector[userNodeKey] = uid
}

func createOrUpdateTolerationForUid(pod *corev1.Pod, uid string, logger logr.Logger) {
	exists := false
	if pod.Spec.Tolerations != nil {
		for _, toleration := range pod.Spec.Tolerations {
			if toleration.Key != userNodeKey {
				continue
			}

			exists = true

			if toleration.Value != userNodeKey {
				toleration.Value = uid
			}

			if toleration.Effect != corev1.TaintEffectNoSchedule {
				toleration.Effect = corev1.TaintEffectNoSchedule
			}
		}
	}

	if !exists {
		logger.Info("Adding new toleration to pod")
		newToleration := corev1.Toleration{
			Key:      userNodeKey,
			Operator: corev1.TolerationOpEqual,
			Value:    uid,
			Effect:   corev1.TaintEffectNoSchedule,
		}
		pod.Spec.Tolerations = append(pod.Spec.Tolerations, newToleration)
	}
}

func (a *podAnnotator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}
