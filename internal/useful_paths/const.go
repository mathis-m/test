package useful_paths

const (
	RelativePathBaseDir           = "slurm-k8s"
	RelativePathScriptsDir        = "scripts"
	RelativePathSystemdUserDir    = ".config/systemd/user"
	RelativePathKubeadmConfigFile = "kubeadm-config.yaml"
)

const (
	ScriptsRootlessctl      = "rootlesskit.sh"
	ScriptsContainerd       = "containerd.sh"
	ScriptsBootstrapCluster = "bootstrap-cluster.sh"
	ScriptsNsenter          = "nsenter.sh"
	ScriptsKubelet          = "kubelet.sh"
)

const (
	ServicesRootlesskit = "slurm-k8s-rootlesskit.service"
	ServicesKubelet     = "slurm-k8s-kubelet.service"
)
