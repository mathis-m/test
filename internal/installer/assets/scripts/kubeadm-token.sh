#!/bin/bash

export BASE_DIR=$(realpath $(dirname $0)/..)
source $BASE_DIR/scripts/common.inc.sh

CONFIG_PATH=$BASE_DIR/kubeadm-config.yaml

log::info "create token"
bash -c "$(dirname $0)/nsenter.sh kubeadm token create $@"

