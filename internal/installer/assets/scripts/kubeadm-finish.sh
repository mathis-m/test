#!/bin/bash

export BASE_DIR=$(realpath $(dirname $0)/..)
source $BASE_DIR/scripts/common.inc.sh

CONFIG_PATH=$BASE_DIR/kubeadm-config.yaml

log::info "Phase: control-plane"
bash -c "$(dirname $0)/nsenter.sh kubeadm init phase control-plane all --config $CONFIG_PATH"

log::info "Phase: etcd"
bash -c "$(dirname $0)/nsenter.sh kubeadm init phase etcd local --config $CONFIG_PATH"

log::info "Phase: upload-config"
bash -c "$(dirname $0)/nsenter.sh kubeadm init phase upload-config all --config $CONFIG_PATH"

log::info "Phase: upload-certs"
bash -c "$(dirname $0)/nsenter.sh kubeadm init phase upload-certs --config $CONFIG_PATH --upload-certs"

log::info "Phase: bootstrap-token"
bash -c "$(dirname $0)/nsenter.sh kubeadm init phase bootstrap-token --config $CONFIG_PATH"

log::info "Phase: kubelet-finalize"
bash -c "$(dirname $0)/nsenter.sh kubeadm init phase kubelet-finalize all --config $CONFIG_PATH"

log::info "Phase: addon"
bash -c "$(dirname $0)/nsenter.sh kubeadm init phase addon all --config $CONFIG_PATH"
