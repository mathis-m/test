package spank_remote

import "C"
import (
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/slurm"
	"github.com/s-bauer/slurm-k8s/internal/util"
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

	bootstrapToken, err := slurm.GetSlurmEnvVar(spank, "SLURM_K8S_BOOTSTRAP_TOKEN")
	if err != nil {
		return fmt.Errorf("unable to retrieve SLURM_K8S_BOOTSTRAP_TOKEN env var: %w", err)
	}

	if !initCluster && !joinCluster {
		return nil
	}

	if err := slurm.FixEnvironmentVariables(spank, importantEnvVars); err != nil {
		return fmt.Errorf("util.FixPathEnvironmentVariable: %w", err)
	}

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
	)
	if err != nil {
		return fmt.Errorf("init failed: %w", err)
	}
	if cmdResult.ExitCode != 0 {
		return fmt.Errorf("init failed")
	}

	return nil
}
