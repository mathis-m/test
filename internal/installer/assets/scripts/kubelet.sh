#!/bin/bash

export BASE_DIR=$(realpath $(dirname $0)/..)
source $BASE_DIR/scripts/common.inc.sh

parent_ip=$(cat $XDG_RUNTIME_DIR/slurm-k8s/parent_ip)

exec $(dirname $0)/nsenter.sh /usr/local/bin/kubelet \
	--bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf \
	--kubeconfig=/etc/kubernetes/kubelet.conf \
	--config=/var/lib/kubelet/config.yaml \
	--container-runtime=remote \
	--container-runtime-endpoint=unix:///run/containerd/containerd.sock \
	--pod-infra-container-image=k8s.gcr.io/pause:3.7 \
  --node-ip=$parent_ip $EXTRA_ARGS
