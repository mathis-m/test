package installer

import (
	"embed"
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/useful_paths"
	log "github.com/sirupsen/logrus"
	"io"
	"io/fs"
	"os"
	"path"
	"text/template"
)

var (
	//go:embed assets/scripts
	scriptFS          embed.FS
	scriptsFSBasePath = "assets/scripts"

	//go:embed assets/slurm-k8s-rootlesskit.service
	serviceRootlesskitTemplateContent string

	//go:embed assets/slurm-k8s-kubelet.service
	serviceKubeletTemplateContent string

	//go:embed assets/kubeadm-config.yaml
	kubeadmTemplateContent string
)

func createBaseDir(paths *useful_paths.UsefulPaths) error {
	logger := log.WithFields(log.Fields{
		"baseDir": paths.BaseDir,
	})

	if _, err := os.Stat(paths.BaseDir); !os.IsNotExist(err) {
		return fmt.Errorf("folder '%v' already exists", paths.BaseDir)
	}

	if err := os.Mkdir(paths.BaseDir, 0700); err != nil {
		return fmt.Errorf("unable to create folder '%v': %w", paths.BaseDir, err)
	}

	logger.Info("created base directory")
	return nil
}

func copyScripts(paths *useful_paths.UsefulPaths) error {
	logger := log.WithFields(log.Fields{
		"scriptDir": paths.ScriptDir,
	})

	if _, err := os.Stat(paths.ScriptDir); !os.IsNotExist(err) {
		return fmt.Errorf("folder '%v' already exists", paths.ScriptDir)
	}

	if err := os.Mkdir(paths.ScriptDir, 0700); err != nil {
		return fmt.Errorf("unable to create folder '%v': %w", paths.ScriptDir, err)
	}

	scripts, err := scriptFS.ReadDir(scriptsFSBasePath)
	if err != nil {
		return fmt.Errorf("cannot read embedded scripts: %w", err)
	}

	for _, script := range scripts {
		if err = writeScript(script, paths.ScriptDir); err != nil {
			return fmt.Errorf("unable to copy script '%v': %w", script.Name(), err)
		}
	}

	logger.Info("copied scripts")

	return nil
}

func createSystemdServices(paths *useful_paths.UsefulPaths) error {
	if err := createSystemdService(
		paths,
		useful_paths.ServicesRootlesskit,
		paths.Services.Rootlesskit,
		serviceRootlesskitTemplateContent,
	); err != nil {
		return err
	}

	if err := createSystemdService(
		paths,
		useful_paths.ServicesKubelet,
		paths.Services.Kubelet,
		serviceKubeletTemplateContent,
	); err != nil {
		return err
	}

	return nil
}

func createSystemdService(paths *useful_paths.UsefulPaths, service string, servicePath string, templateContent string) error {
	logger := log.WithFields(log.Fields{
		"servicePath": servicePath,
	})

	targetPath := servicePath

	serviceTemplate, err := template.New(service).Parse(templateContent)
	if err != nil {
		return fmt.Errorf("unable to parse systemd service template: %w", err)
	}

	if _, err = os.Stat(paths.SystemdUserDir); os.IsNotExist(err) {
		if err = os.MkdirAll(paths.SystemdUserDir, 0755); err != nil {
			return fmt.Errorf("unable to create folder '%v': %w", paths.SystemdUserDir, err)
		}
	}

	if _, err = os.Stat(targetPath); !os.IsNotExist(err) {
		return fmt.Errorf("file '%v' already exists", targetPath)
	}

	targetFile, err := os.Create(targetPath)
	defer targetFile.Close()

	if err := serviceTemplate.Execute(targetFile, paths); err != nil {
		return fmt.Errorf("executing service template failed: %w", err)
	}

	logger.Info("created systemd service")

	return nil
}

func createKubeadmConfig(paths *useful_paths.UsefulPaths) error {
	targetPath := paths.KubeadmConfig

	logger := log.WithFields(log.Fields{
		"targetPath": targetPath,
	})

	kubeadmTemplate, err := template.New(useful_paths.RelativePathKubeadmConfigFile).Parse(kubeadmTemplateContent)
	if err != nil {
		return fmt.Errorf("unable to parse kubeadm template: %w", err)
	}

	if _, err = os.Stat(targetPath); !os.IsNotExist(err) {
		return fmt.Errorf("file '%v' already exists", targetPath)
	}

	targetFile, err := os.Create(targetPath)
	defer targetFile.Close()

	if err := kubeadmTemplate.Execute(targetFile, nil); err != nil {
		return fmt.Errorf("executing service template failed: %w", err)
	}

	logger.Info("created kubeadm template")
	return nil
}

func writeScript(script fs.DirEntry, targetDir string) error {
	scriptPath := path.Join(scriptsFSBasePath, script.Name())
	targetPath := path.Join(targetDir, script.Name())

	logger := log.WithFields(log.Fields{
		"sourcePath": scriptPath,
		"targetPath": targetPath,
	})

	if err := writeFile(scriptPath, targetPath); err != nil {
		return err
	}

	if err := os.Chmod(targetPath, 0755); err != nil {
		return fmt.Errorf("unable to change permissions for file '%v': %w", targetPath, err)
	}

	logger.Debug("copied embedded script")

	return nil
}

func writeFile(sourcePath string, targetPath string) error {
	sourceFile, err := scriptFS.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("unable to read embedded script '%v': %w", sourcePath, err)
	}
	defer sourceFile.Close()

	if _, err = os.Stat(targetPath); !os.IsNotExist(err) {
		return fmt.Errorf("file '%v' already exists: %w", targetPath, err)
	}

	targetFile, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("unable to create file '%v': %w", targetPath, err)
	}
	defer targetFile.Close()

	if _, err = io.Copy(targetFile, sourceFile); err != nil {
		return fmt.Errorf("uanble to copy to destination '%v': %w", targetFile, err)
	}

	return nil
}
