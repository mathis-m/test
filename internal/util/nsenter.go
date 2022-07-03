package util

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

const (
	PipeEnvKey = "_SLURM_K8S_PIPEFD"
)

func IsInNamespace() bool {
	return os.Getenv(PipeEnvKey) != ""
}

func ReexecuteInNamespace() error {
	pipeR, _, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("unable to create pipe: %w", err)
	}

	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	childPidPath := path.Join(runtimeDir, "usernetes", "rootlesskit", "child_pid")
	if _, err := os.Stat(childPidPath); os.IsNotExist(err) {
		return fmt.Errorf("child_pid file doesn't exist: %w", err)
	}

	childPid, err := ioutil.ReadFile(childPidPath)
	if err != nil {
		return fmt.Errorf("unable to read child_pid file: %w", err)
	}

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("unable to get executable path: %w", err)
	}

	baseCmd := "nsenter"
	args := []string{
		"--user",
		"--preserve-credential",
		"--mount",
		"--net",
		"--cgroup",
		"--pid",
		"--ipc",
		"--uts",
		fmt.Sprintf("-t %v", string(childPid)),
		fmt.Sprintf("--wd=%v", os.Getenv("PWD")),
		"--",
		executable,
	}
	args = append(args, os.Args[1:]...)

	log.Debugf("child command: %v %v", baseCmd, args)

	cmd := exec.Command(baseCmd, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = []*os.File{pipeR}
	cmd.Env = append(os.Environ(), fmt.Sprintf("%v=%v", PipeEnvKey, "3"))

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to start command: %w", err)
	}

	// no messages needed at the moment
	// encoder := gob.NewEncoder(pipeW)

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}
