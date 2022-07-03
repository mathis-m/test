package util

import (
	"bufio"
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os/exec"
	"syscall"
)

type CommandResult struct {
	Stdout         string
	Stderr         string
	CombinedOutput string

	ExitCode int
}

func readStream(pipe io.Reader, outputs []io.StringWriter, logOutput bool, logPrefix string) {
	scanner := bufio.NewScanner(pipe)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		text := scanner.Text()

		if outputs != nil {
			for _, output := range outputs {
				_, err := output.WriteString(text)
				if err != nil {
					log.Warn("unable to write process std steams to output: ", err)
				}
			}
		}

		if logOutput {
			log.Debugf("%v: %v", logPrefix, text)
		}
	}
}

func RunProcess(name string, command string) (*exec.Cmd, error) {
	cmd := exec.Command("/bin/bash", "-c", command)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	log.Info("Starting ", name)
	if err := cmd.Start(); err != nil {
		log.Error("Unable to start ", name, " :", err)
		return nil, err
	}
	log.Info("Started ", name)

	go readStream(stdout, nil, true, name)
	go readStream(stderr, nil, true, name)

	log.Info("Waiting for ", name)
	if err := cmd.Wait(); err != nil {
		log.Error("Waiting for ", name, ":", err)
		return nil, err
	}

	return cmd, nil
}

func RunProcessGetStdout(name string, command string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", command)

	log.Info("Running", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("Process ", name, " failed:", err)
		return "", fmt.Errorf("failed to run '%v': %w", name, err)
	}

	return string(output), nil
}

// RunCommand taken from https://github.com/kardianos/service and modified
func RunCommand(command string, arguments ...string) (*CommandResult, error) {
	cmd := exec.Command(command, arguments...)

	log.Debugf("Executing command %v", append([]string{command}, arguments...))

	var outputBuffer bytes.Buffer
	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("unable to open stdout pipe: %w", err)
	}

	stderr, _ := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("unable to open stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("%q failed: %w", command, err)
	}

	go readStream(stdout, []io.StringWriter{&stdoutBuffer, &outputBuffer}, true, command)
	go readStream(stderr, []io.StringWriter{&stderrBuffer, &outputBuffer}, true, command)

	err = cmd.Wait()
	result := &CommandResult{
		Stdout:         stdoutBuffer.String(),
		Stderr:         stderrBuffer.String(),
		CombinedOutput: outputBuffer.String(),
		ExitCode:       0,
	}

	if err != nil {
		exitCode, ok := isExitError(err)
		if ok {
			result.ExitCode = exitCode
			return result, nil
		}

		// An error occurred and there is no exit status.
		return nil, fmt.Errorf("%q failed: %w", command, err)
	}

	return result, nil
}

// isExitError taken from https://github.com/kardianos/service and modified
func isExitError(err error) (int, bool) {
	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus(), true
		}
	}

	return 0, false
}
