#!/bin/bash

# Taken from github.com/rootless-containers/usernetes and modified
# License: https://github.com/rootless-containers/usernetes/blob/master/LICENSE

export BASE_DIR=$(realpath $(dirname $0)/..)
source $BASE_DIR/scripts/common.inc.sh
nsenter::main $0 $@

export KUBECONFIG=/etc/kubernetes/admin.conf

if [[ $# -eq 0 ]]; then
	exec $SHELL $@
else
	exec $@
fi
