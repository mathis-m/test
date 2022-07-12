package spank_remote

import "C"
import (
	"fmt"
	"github.com/abrekhov/hostlist"
	"github.com/s-bauer/slurm-k8s/internal/slurm"
	"github.com/s-bauer/slurm-k8s/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
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
	// debug tmp
	slurmVars, err := slurm.GetSlurmEnvVars(spank)
	if err != nil {
		return err
	}
	for _, env := range slurmVars {
		log.Infof("SLURM: %v", env)
	}

	initCluster := viper.GetBool("k8s-init-cluster")
	joinCluster := viper.GetBool("k8s-join-cluster")

	if !initCluster && !joinCluster {
		return nil
	}

	// Copy env vars
	if err := slurm.FixEnvironmentVariables(spank, importantEnvVars); err != nil {
		return fmt.Errorf("util.FixPathEnvironmentVariable: %w", err)
	}

	firstNode, err := isFirstNode(spank)
	if err != nil {
		return fmt.Errorf("unable to determine isFirstNode: %w", err)
	}

	if initCluster && firstNode {
		if err := runInitCluster(spank); err != nil {
			return err
		}
	} else if joinCluster || (initCluster && !firstNode) {
		if err := runJoinCluster(spank); err != nil {
			return err
		}
	}

	return nil
}

func getFirstNode(spank unsafe.Pointer) (string, error) {
	jobHostListString, err := slurm.GetSlurmEnvVar(spank, "SLURM_JOB_NODELIST")
	if err != nil {
		return "", fmt.Errorf("unable to get SLURM_JOB_NODELIST: %w", err)
	}

	jobHostList := hostlist.ExpandNodeList(jobHostListString)

	if len(jobHostList) == 0 {
		return "", fmt.Errorf("host list is empty")
	}

	return jobHostList[0], nil
}

func isFirstNode(spank unsafe.Pointer) (bool, error) {
	firstNode, err := getFirstNode(spank)
	if err != nil {
		return false, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return false, fmt.Errorf("unable to get hostname: %w", err)
	}

	return firstNode == hostname, nil
}

func runJoinCluster(spank unsafe.Pointer) error {
	// Get SLURM_K8S env variables
	bootstrapToken, err := slurm.GetSlurmEnvVar(spank, "SLURM_K8S_BOOTSTRAP_TOKEN")
	if err != nil {
		return fmt.Errorf("unable to retrieve SLURM_K8S_BOOTSTRAP_TOKEN env var: %w", err)
	}

	certHash, err := slurm.GetSlurmEnvVar(spank, "SLURM_K8S_CA_CERT_HASH")
	if err != nil {
		return fmt.Errorf("unable to retrieve SLURM_K8S_CA_CERT_HASH env var: %w", err)
	}

	apiEndpoint, err := slurm.GetSlurmEnvVar(spank, "SLURM_K8S_API_ENDPOINT")
	if err != nil {
		apiEndpoint, err = getFirstNode(spank)
		if err != nil {
			return err
		}
	}

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

	// run join
	cmdResult, err = util.RunCommand(
		"/home/simon/spank-go/bin/bootstrap-uk8s",
		"--verbose",
		"--simple-log",
		fmt.Sprintf("--drop-uid=%v", jobUser.Uid),
		fmt.Sprintf("--drop-gid=%v", jobUser.Gid),
		"join",
		fmt.Sprintf("--token=%v", bootstrapToken),
		fmt.Sprintf("--api-server-endpoint=%v:6443", apiEndpoint),
		fmt.Sprintf("--discovery-token-ca-cert-hash=sha256:%v", certHash),
	)
	if err != nil {
		return fmt.Errorf("join failed: %w", err)
	}
	if cmdResult.ExitCode != 0 {
		return fmt.Errorf("join failed")
	}

	return nil
}

func runInitCluster(spank unsafe.Pointer) error {
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
