#!/bin/bash

export BASE_DIR=$(realpath $(dirname $0)/..)
source $BASE_DIR/scripts/common.inc.sh

# Kubeadm prepare
$BASE_DIR/scripts/kubeadm-prepare.sh

# Start kubelet in background
$BASE_DIR/scripts/kubelet.sh &
KUBELET_PID=$!

# Kubeadm finish
$BASE_DIR/scripts/kubeadm-finish.sh

# Attach to kubelet
wait $KUBELET_PID
