package util

import (
	"fmt"
	"strings"
)

type Service struct {
	Name string
}

type ServiceStatus string

const (
	Active   ServiceStatus = "active"
	Inactive ServiceStatus = "inactive"
)

func ReloadSystemdDaemon() error {
	cmdResult, err := RunCommand("systemctl", "--user", "daemon-reload")
	if err != nil || cmdResult.ExitCode != 0 {
		return fmt.Errorf("unable to do reload systemctl daemon %w", err)
	}

	return nil
}

func (service *Service) Status() (ServiceStatus, error) {
	cmdResult, err := RunCommand("systemctl", "--user", "is-active", service.Name)
	if err != nil {
		return "", fmt.Errorf("unable to query service status for %q: %w", service.Name, err)
	}

	if strings.HasPrefix(cmdResult.CombinedOutput, "active") {
		return Active, nil
	}

	return Inactive, nil
}

func (service *Service) Start() error {
	cmdResult, err := RunCommand("systemctl", "--user", "start", service.Name)
	if err != nil || cmdResult.ExitCode != 0 {
		return fmt.Errorf("unable to start service %q: %w", service.Name, err)
	}

	return nil
}

func (service *Service) Stop() error {
	cmdResult, err := RunCommand("systemctl", "--user", "stop", service.Name)
	if err != nil || cmdResult.ExitCode != 0 {
		return fmt.Errorf("unable to stop service %q: %w", service.Name, err)
	}

	return nil
}

func (service *Service) Restart() error {
	cmdResult, err := RunCommand("systemctl", "--user", "restart", service.Name)
	if err != nil || cmdResult.ExitCode != 0 {
		return fmt.Errorf("unable to restart service %q: %w", service.Name, err)
	}

	return nil
}

func (service *Service) Reload() error {
	cmdResult, err := RunCommand("systemctl", "--user", "reload", service.Name)
	if err != nil || cmdResult.ExitCode != 0 {
		return fmt.Errorf("unable to reload service %q: %w", service.Name, err)
	}

	return nil
}
