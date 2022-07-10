package util

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"
)

const (
	PipeEnvKey = "_SLURM_K8S_PIPEFD"
)

type ChildResult = map[any]any

func IsInNamespace() bool {
	return os.Getenv(PipeEnvKey) != ""
}

func WriteResult(msg ChildResult) error {
	pipeFDStr := os.Getenv(PipeEnvKey)
	if pipeFDStr == "" {
		return fmt.Errorf("%s is not set", PipeEnvKey)
	}

	pipeFD, err := strconv.Atoi(pipeFDStr)
	if err != nil {
		return fmt.Errorf("unexpected fd value: %s: %w", pipeFDStr, err)
	}

	pipeW := os.NewFile(uintptr(pipeFD), "")
	encoder := gob.NewEncoder(pipeW)

	if err := encoder.Encode(msg); err != nil {
		return err
	}

	return nil
}

func ReexecuteInNamespace(additionalEnvVars []string) (ChildResult, error) {
	pipeR, pipeW, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("unable to create pipe: %w", err)
	}
	defer pipeR.Close()
	defer pipeW.Close()

	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	childPidPath := path.Join(runtimeDir, "slurm-k8s", "rootlesskit", "child_pid")

	maxLoops := 10
	for i := 0; i < maxLoops; i++ {
		if _, err := os.Stat(childPidPath); os.IsNotExist(err) {
			if i == maxLoops-1 {
				return nil, fmt.Errorf("child_pid file doesn't exist: %w", err)
			}

			time.Sleep(1 * time.Second)
		}
	}

	childPid, err := ioutil.ReadFile(childPidPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read child_pid file: %w", err)
	}

	executable, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("unable to get executable path: %w", err)
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

	log.Debugf("child command: %v", append([]string{baseCmd}, args...))

	cmd := exec.Command(baseCmd, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = []*os.File{pipeW}
	cmd.Env = append(os.Environ(), fmt.Sprintf("%v=%v", PipeEnvKey, "3"))
	cmd.Env = append(cmd.Env, additionalEnvVars...)

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("unable to start command: %w", err)
	}

	resultChannel := make(chan ChildResult)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go readResult(ctx, pipeR, resultChannel)

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("child failed: %w", err)
	}

	select {
	case <-time.After(500 * time.Millisecond):
		return nil, fmt.Errorf("timeout reading result from child process")
	case result := <-resultChannel:
		return result, nil
	}

}

func readResult(ctx context.Context, pipe *os.File, result chan<- ChildResult) {
	decoder := gob.NewDecoder(pipe)
	childResult := ChildResult{}

	for {
		if err := pipe.SetDeadline(time.Now().Add(250 * time.Millisecond)); err != nil {
			log.Errorf("unable to set deadline: %v", err)
		}

		if err := decoder.Decode(&childResult); err != nil {
			if !errors.Is(err, os.ErrDeadlineExceeded) && !errors.Is(err, io.EOF) && !errors.Is(err, os.ErrClosed) {
				log.Errorf("unable to read result from pipe: %v", err)
			}
		} else {
			result <- childResult
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}
