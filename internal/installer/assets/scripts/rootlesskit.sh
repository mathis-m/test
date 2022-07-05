#!/bin/bash

# Taken from github.com/rootless-containers/usernetes and modified
# License: https://github.com/rootless-containers/usernetes/blob/master/LICENSE

export BASE_DIR=$(realpath $(dirname $0)/..)
source $BASE_DIR/scripts/common.inc.sh

rk_state_dir=$XDG_RUNTIME_DIR/slurm-k8s/rootlesskit

: ${SLURM_K8S_CHILD=0}

if [[ $SLURM_K8S_CHILD == 0 ]]; then

	SLURM_K8S_CHILD=1
	if hostname -I &>/dev/null ; then
		: ${SLURM_K8S_PARENT_IP=$(hostname -I | sed -e 's/ .*//g')}
	else
		: ${SLURM_K8S_PARENT_IP=$(hostname -i | sed -e 's/ .*//g')}
	fi
	export SLURM_K8S_CHILD SLURM_K8S_PARENT_IP

	# Re-exec the script via RootlessKit, so as to create unprivileged namespaces.
	rootlesskit \
		--state-dir $rk_state_dir \
		--net=slirp4netns --mtu=65520 --disable-host-loopback --slirp4netns-sandbox=true --slirp4netns-seccomp=true \
		--cidr=10.24.0.0/16 \
		--port-driver=builtin \
		--copy-up=/etc --copy-up=/run --copy-up=/var/lib --copy-up=/opt \
		--cgroupns \
		--pidns \
		--ipcns \
		--utsns \
		--propagation=rslave \
		--evacuate-cgroup2="rootlesskit_evac" \
		$0 $@
else
	# save IP address
	echo $SLURM_K8S_PARENT_IP > $XDG_RUNTIME_DIR/slurm-k8s/parent_ip

	# Remove symlinks so that the child won't be confused by the parent configuration
	rm -f \
		/run/xtables.lock \
		/run/flannel \
		/run/netns \
		/run/runc \
		/run/crun \
		/run/containerd \
		/run/containers \
		/run/crio

	rm -f \
		/etc/cni \
		/etc/containerd \
		/etc/containers \
		/etc/crio \
		/etc/kubernetes

	rm -f \
		/var/lib/etcd \
		/var/lib/kubelet \
		/var/lib/containerd

	rm -f \
		/opt/cni

	# CNi
	mkdir -p /opt/cni/bin
	mount --bind /home/simon/tmp/bin/cni /opt/cni/bin

	# Copy CNI config to /etc/cni/net.d (Likely to be hardcoded in CNI installers)
	# mkdir -p /etc/cni/net.d
	# cp -f $BASE_DIR/config/cni_net.d/* /etc/cni/net.d

	# These bind-mounts are needed at the moment because the paths are hard-coded in Kube and CRI-O.
	binds=(/var/lib/cni /var/log /var/lib/containers /var/cache)
	for f in ${binds[@]}; do
		src=$XDG_DATA_HOME/slurm-k8s/$(echo $f | sed -e s@/@_@g)
		if [[ -L $f ]]; then
			# Remove link created by `rootlesskit --copy-up` if any
			rm -rf $f
		fi
		mkdir -p $src $f
		mount --bind $src $f
	done

	rk_pid=$(cat $rk_state_dir/child_pid)

	# workaround for https://github.com/rootless-containers/rootlesskit/issues/37
	# child_pid might be created before the child is ready
	echo $rk_pid >$rk_state_dir/_child_pid.u7s-ready
	log::info "RootlessKit ready, PID=${rk_pid}, state directory=$rk_state_dir ."
	log::info "Hint: You can enter RootlessKit namespaces by running \`nsenter -U --preserve-credential -n -m -t ${rk_pid}\`."
	
	# Add required ports
	# > Kubernetes API server (control-plane)
	rootlessctl --socket $rk_state_dir/api.sock add-ports 0.0.0.0:6443:6443/tcp
	# > kube-scheduler (control-plane)
	rootlessctl --socket $rk_state_dir/api.sock add-ports 0.0.0.0:10259:10259/tcp
	# > kube-controller-manager (control-plane)
	rootlessctl --socket $rk_state_dir/api.sock add-ports 0.0.0.0:10257:10257/tcp

	# > kubelet API
	rootlessctl --socket $rk_state_dir/api.sock add-ports 0.0.0.0:10250:10250/tcp
	# > flannel
	rootlessctl --socket $rk_state_dir/api.sock add-ports 0.0.0.0:8472:8472/udp

	# Add public IP to loopback interface
	ip addr add $SLURM_K8S_PARENT_IP dev lo

	# Execute command in namespace
	rc=0
	if [[ $# -eq 0 ]]; then
		sleep infinity || rc=$?
	else
		# $@ || rc=$?
		exec $@ || rc=$?
	fi

	# Finish
	log::info "RootlessKit exiting (status=$rc)"
	exit $rc
fi
