#!/bin/bash

# Taken from github.com/rootless-containers/usernetes and modified
# License: https://github.com/rootless-containers/usernetes/blob/master/LICENSE


# needs to be called inside the namespaces
export BASE_DIR=$(realpath $(dirname $0)/..)
source $BASE_DIR/scripts/common.inc.sh

mkdir -p $XDG_RUNTIME_DIR/slurm-k8s
cat >$XDG_RUNTIME_DIR/slurm-k8s/containerd.toml <<EOF
version = 2
[plugins]
  [plugins."io.containerd.grpc.v1.cri"]
    disable_cgroup = false
    disable_apparmor = true
    restrict_oom_score_adj = true
    disable_hugetlb_controller = true
    [plugins."io.containerd.grpc.v1.cri".cni]
      bin_dir = "/opt/cni/bin"
      conf_dir = "/etc/cni/net.d"
EOF

exec containerd -c $XDG_RUNTIME_DIR/slurm-k8s/containerd.toml $@