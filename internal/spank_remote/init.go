package spank_remote

import "C"
import (
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/slurm"
	"github.com/s-bauer/slurm-k8s/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"unsafe"
)

var (
	importantEnvVars = []string{
		"LANG",
		"PATH",
		"HOME",
		"LOGNAME",
		"USER",
		"SHELL",
		"TERM",
		"MAIL",
		"XDG_SESSION_ID",
		"XDG_RUNTIME_DIR",
		"DBUS_SESSION_BUS_ADDRESS",
		"XDG_SESSION_TYPE",
		"XDG_SESSION_CLASS",
		"PWD",
	}
)

func Init(spank unsafe.Pointer) error {
	//if err := slurm.FixPathEnvironmentVariable(spank); err != nil {
	//	return fmt.Errorf("util.FixPathEnvironmentVariable: %w", err)
	//}
	//
	//for _, env := range os.Environ() {
	//	log.Infof("Init: [%v]", env)
	//}

	//
	//jobUser, err := slurm.GetJobUser(spank)
	//if err != nil {
	//	return fmt.Errorf("slurm.GetJobUser: %w", err)
	//}
	//
	//kubeCluster := kube.NewKubernetesCluster(jobUser)
	//if err = kubeCluster.Initialize(); err != nil {
	//	return fmt.Errorf("kubeCluster.Initialize: %w", err)
	//}
	//
	//if err = kubeCluster.InitializeAdminUser(); err != nil {
	//	return fmt.Errorf("kubeCluster.InitializeAdminUser: %w", err)
	//}

	return nil
}

func UserInit(spank unsafe.Pointer) error {
	initCluster := viper.GetBool("k8s-init-cluster")
	joinCluster := viper.GetBool("k8s-join-cluster")

	if !initCluster && !joinCluster {
		return nil
	}

	// Copy env vars
	if err := slurm.FixEnvironmentVariables(spank, importantEnvVars); err != nil {
		return fmt.Errorf("util.FixPathEnvironmentVariable: %w", err)
	}

	// Get SLURM_K8S env variables
	bootstrapToken, err := slurm.GetSlurmEnvVar(spank, "SLURM_K8S_BOOTSTRAP_TOKEN")
	if err != nil {
		return fmt.Errorf("unable to retrieve SLURM_K8S_BOOTSTRAP_TOKEN env var: %w", err)
	}

	caCertB64, err := slurm.GetSlurmEnvVar(spank, "SLURM_K8S_CA_CERT")
	if err != nil {
		return fmt.Errorf("unable to retrieve SLURM_K8S_CA_CERT env var: %w", err)
	}

	caKeyB64, err := slurm.GetSlurmEnvVar(spank, "SLURM_K8S_CA_KEY")
	if err != nil {
		return fmt.Errorf("unable to retrieve SLURM_K8S_CA_KEY env var: %w", err)
	}

	log.Infof(
		"SLURM_K8S_BOOTSTRAP_TOKEN: %v, SLURM_K8S_CA_CERT: %v, SLURM_K8S_CA_KEY: %v",
		bootstrapToken,
		caCertB64,
		caKeyB64,
	)

	// prepare
	jobUser, err := slurm.GetJobUser(spank)
	if err != nil {
		return fmt.Errorf("slurm.GetJobUser: %w", err)
	}

	// run install
	cmdResult, err := util.RunCommand(
		"/home/simon/spank-go/bin/bootstrap-uk8s",
		"--verbose",
		"--simple-log",
		fmt.Sprintf("--drop-uid=%v", jobUser.Uid),
		fmt.Sprintf("--drop-gid=%v", jobUser.Gid),
		"install",
		"--force",
	)
	if err != nil {
		return fmt.Errorf("install failed: %w", err)
	}
	if cmdResult.ExitCode != 0 {
		return fmt.Errorf("install failed")
	}

	// run init
	cmdResult, err = util.RunCommand(
		"/home/simon/spank-go/bin/bootstrap-uk8s",
		"--verbose",
		"--simple-log",
		fmt.Sprintf("--drop-uid=%v", jobUser.Uid),
		fmt.Sprintf("--drop-gid=%v", jobUser.Gid),
		"init",
		fmt.Sprintf("--token=%v", bootstrapToken),
		fmt.Sprintf("--ca-cert-b64=%v", caCertB64),
		fmt.Sprintf("--ca-key-b64=%v", caKeyB64),
	)
	if err != nil {
		return fmt.Errorf("init failed: %w", err)
	}
	if cmdResult.ExitCode != 0 {
		return fmt.Errorf("init failed")
	}

	return nil
}
