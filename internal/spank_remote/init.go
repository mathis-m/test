package spank_remote

import "C"
import (
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/slurm"
	"github.com/s-bauer/slurm-k8s/internal/util"
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
	if err := slurm.FixEnvironmentVariables(spank, importantEnvVars); err != nil {
		return fmt.Errorf("util.FixPathEnvironmentVariable: %w", err)
	}

	jobUser, err := slurm.GetJobUser(spank)
	if err != nil {
		return fmt.Errorf("slurm.GetJobUser: %w", err)
	}

	// save existing ids
	//savedUid := syscall.Getuid()
	//log.Infof("getuid: %v", savedUid)
	//
	//savedEuid := syscall.Geteuid()
	//log.Infof("geteuid: %v", savedEuid)
	//
	//savedGid := syscall.Getgid()
	//log.Infof("getgid: %v", savedGid)
	//
	//savedEgid := syscall.Getegid()
	//log.Infof("getegid: %v", savedEgid)
	//
	//savedGroups, err := syscall.Getgroups()
	//if err != nil {
	//	log.Infof("getgroups: error: %v", err)
	//}
	//log.Infof("getgroups: %v", savedGroups)
	//
	//// drop privileges
	//newUid, err := strconv.Atoi(jobUser.Uid)
	//if err != nil {
	//	return fmt.Errorf("unable to convert uid %q to int", newUid)
	//}
	//
	//newGid, err := strconv.Atoi(jobUser.Gid)
	//if err != nil {
	//	return fmt.Errorf("unable to convert gid %q to int", newGid)
	//}
	//
	//if err := syscall.Setregid(newUid, newUid); err != nil {
	//	log.Warnf("Setegid failed: %v", err)
	//}
	//
	//if err := syscall.Setreuid(newGid, newGid); err != nil {
	//	log.Warnf("Seteuid failed: %v", err)
	//}

	// Run test command
	//cmd := exec.Command("bash", "-c", "systemctl --user is-active containerd")
	//output, err := cmd.CombinedOutput()
	//if err != nil {
	//	log.Infof("command failed with: %v", err)
	//}
	//log.Infof("Output: %v", string(output))

	// regain privileges - DOESN'T WORK!
	//if err := syscall.Setregid(savedGid, savedGid); err != nil {
	//	log.Warnf("Setregid failed: %v", err)
	//}
	//
	//if err := syscall.Setreuid(savedUid, savedUid); err != nil {
	//	log.Warnf("Setregid failed: %v", err)
	//}

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
	)
	if err != nil {
		return fmt.Errorf("init failed: %w", err)
	}

	if cmdResult.ExitCode != 0 {
		return fmt.Errorf("init failed")
	}

	return nil
}
