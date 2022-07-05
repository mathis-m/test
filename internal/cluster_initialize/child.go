package cluster_initialize

import (
	_ "embed"
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/useful_paths"
	"github.com/s-bauer/slurm-k8s/internal/util"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
)

var (
	//go:embed assets/slurm-k8s-manager.yaml
	slurmK8sManager string
)

func childInitialize() error {
	if err := os.Setenv("KUBECONFIG", "/etc/kubernetes/admin.conf"); err != nil {
		return fmt.Errorf("unable to set KUBECONFIG env var: %w", err)
	}

	usefulPaths, err := useful_paths.ConstructUsefulPaths()
	if err != nil {
		return err
	}

	// Copy kubeconfig outside of namespace
	if err := util.EnsureFolderExistsWithPermissions(path.Join(usefulPaths.HomeDir, ".kube"), 0755); err != nil {
		return err
	}
	if err := util.CopyFile(usefulPaths.KubernetesAdminConfig, usefulPaths.KubernetesUserConfig, true, 0700); err != nil {
		return err
	}
	log.Infof("copied kubernetes config to %q", usefulPaths.KubernetesUserConfig)

	// Install flannel
	cmdResult, err := util.RunCommand("kubectl", "apply", "-f", "https://raw.githubusercontent.com/flannel-io/flannel/master/Documentation/kube-flannel.yml")
	if err != nil {
		return fmt.Errorf("unable to execute kubectl: %w", err)
	}
	if cmdResult.ExitCode != 0 {
		return fmt.Errorf("kubectl failed with exit code %v", cmdResult.ExitCode)
	}
	log.Infof("installed flannel")

	// Install manager / controller / operator
	if err := deploySlurmK8sManager(); err != nil {
		return err
	}
	log.Info("installed flannel-annotator")

	// Get Token and Cert
	token, err := getJoinToken()
	if err != nil {
		return err
	}

	certHash, err := getCertThumbprint()
	if err != nil {
		return err
	}

	if err := util.WriteResult(util.ChildResult{
		"token":    token,
		"certHash": certHash,
	}); err != nil {
		return fmt.Errorf("unable to write result: %w", err)
	}

	log.Info("wrote result. exiting...")

	return nil
}

func deploySlurmK8sManager() error {
	cmd := "cat <<EOF | kubectl apply -f - \n"
	cmd += slurmK8sManager
	cmd += "\nEOF"

	cmdResult, err := util.RunCommand("bash", "-c", cmd)
	if err != nil {
		return fmt.Errorf("slurm-k8s-manager: kubectl apply failed: %w", err)
	}
	if cmdResult.ExitCode != 0 {
		return fmt.Errorf("slurm-k8s-manager: process exited with exit code: %v", cmdResult.ExitCode)
	}

	return nil
}

func getCertThumbprint() (string, error) {
	cmd := "openssl x509 -pubkey -in /etc/kubernetes/pki/ca.crt | "
	cmd += "openssl rsa -pubin -outform der 2>/dev/null | "
	cmd += "openssl dgst -sha256 -hex | "
	cmd += "sed 's/^.* //'"

	cmdResult, err := util.RunCommand("bash", "-c", cmd)
	if err != nil {
		return "", fmt.Errorf("get cert thumbprint: %w", err)
	}
	if cmdResult.ExitCode != 0 {
		return "", fmt.Errorf("get cert thumbprint: process exited with exit code: %v", cmdResult.ExitCode)
	}

	return cmdResult.Stdout, nil
}

func getJoinToken() (string, error) {
	cmdResult, err := util.RunCommand("kubeadm", "token", "create")
	if err != nil {
		return "", fmt.Errorf("get join token: %w", err)
	}
	if cmdResult.ExitCode != 0 {
		return "", fmt.Errorf("get join token: process existed with exit code: %v", cmdResult.ExitCode)
	}

	return cmdResult.Stdout, nil
}
