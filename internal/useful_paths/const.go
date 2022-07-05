package useful_paths

const (
	RelativePathBaseDir                  = "slurm-k8s"
	RelativePathScriptsDir               = "scripts"
	RelativePathSystemdUserDir           = ".config/systemd/user"
	RelativePathKubeadmAdminConfigFile   = "kubeadm-config.yaml"
	RelativePathKubernetesUserConfigFile = ".kube/config"

	PathKubernetesAdminConfigFile = "/etc/kubernetes/admin.conf"
)

const (
	ScriptsRootlessctl    = "rootlesskit.sh"
	ScriptsContainerd     = "containerd.sh"
	ScriptsNsenter        = "nsenter.sh"
	ScriptsKubelet        = "kubelet.sh"
	ScriptsKubeadmPrepare = "kubeadm-prepare.sh"
	ScriptsKubeadmFinish  = "kubeadm-finish.sh"
)

const (
	ServicesRootlesskit = "slurm-k8s-rootlesskit.service"
	ServicesKubelet     = "slurm-k8s-kubelet.service"
)
