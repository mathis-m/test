package util

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os/exec"
)

func readStream(pipe io.ReadCloser) {
	scanner := bufio.NewScanner(pipe)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		text := scanner.Text()
		log.Info(text)
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

	go readStream(stdout)
	go readStream(stderr)

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
