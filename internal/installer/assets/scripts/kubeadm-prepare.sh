#!/bin/bash

export BASE_DIR=$(realpath $(dirname $0)/..)
source $BASE_DIR/scripts/common.inc.sh

CONFIG_PATH=$BASE_DIR/kubeadm-config.yaml

log::info "Phase: preflight"
bash -c "$(dirname $0)/nsenter.sh kubeadm init phase preflight --config $CONFIG_PATH"

log::info "Phase: certs"
bash -c "$(dirname $0)/nsenter.sh kubeadm init phase certs all --config $CONFIG_PATH"

log::info "Phase: kubeconfig"
bash -c "$(dirname $0)/nsenter.sh kubeadm init phase kubeconfig all --config $CONFIG_PATH"

log::info "Phase: kubelet-start"
bash -c "$(dirname $0)/nsenter.sh kubeadm init phase kubelet-start --config $CONFIG_PATH"
