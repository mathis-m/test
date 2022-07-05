package installer

import (
	"embed"
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/useful_paths"
	"github.com/s-bauer/slurm-k8s/internal/util"
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

func createBaseDir(paths *useful_paths.UsefulPaths, force bool) error {
	logger := log.WithFields(log.Fields{
		"baseDir": paths.BaseDir,
	})

	if _, err := os.Stat(paths.BaseDir); !os.IsNotExist(err) {
		if force {
			return nil
		}

		return fmt.Errorf("folder '%v' already exists", paths.BaseDir)
	}

	if err := os.Mkdir(paths.BaseDir, 0700); err != nil {
		return fmt.Errorf("unable to create folder '%v': %w", paths.BaseDir, err)
	}

	logger.Info("created base directory")
	return nil
}

func copyScripts(paths *useful_paths.UsefulPaths, force bool) error {
	logger := log.WithFields(log.Fields{
		"scriptDir": paths.ScriptDir,
	})

	if _, err := os.Stat(paths.ScriptDir); os.IsNotExist(err) {
		if err := os.Mkdir(paths.ScriptDir, 0700); err != nil {
			return fmt.Errorf("unable to create folder '%v': %w", paths.ScriptDir, err)
		}
	} else {
		if !force {
			return fmt.Errorf("folder %q already exists", paths.ScriptDir)
		}
	}

	scripts, err := scriptFS.ReadDir(scriptsFSBasePath)
	if err != nil {
		return fmt.Errorf("cannot read embedded scripts: %w", err)
	}

	for _, script := range scripts {
		if err = writeScript(script, paths.ScriptDir, force); err != nil {
			return fmt.Errorf("unable to copy script %q: %w", script.Name(), err)
		}
	}

	logger.Info("copied scripts")

	return nil
}

func createSystemdServices(paths *useful_paths.UsefulPaths, force bool) error {
	if err := createSystemdService(
		paths,
		useful_paths.ServicesRootlesskit,
		paths.Services.Rootlesskit,
		serviceRootlesskitTemplateContent,
		force,
	); err != nil {
		return err
	}

	if err := createSystemdService(
		paths,
		useful_paths.ServicesKubelet,
		paths.Services.Kubelet,
		serviceKubeletTemplateContent,
		force,
	); err != nil {
		return err
	}

	if err := util.ReloadSystemdDaemon(); err != nil {
		return err
	}

	log.Info("reloaded systemd daemon")

	return nil
}

func createSystemdService(
	paths *useful_paths.UsefulPaths,
	service string,
	servicePath string,
	templateContent string,
	force bool,
) error {
	logger := log.WithFields(log.Fields{
		"servicePath": servicePath,
	})

	targetPath := servicePath

	logger.Info("trying to create service file")
	logger.Info("Service: ", service)
	logger.Info("TemplateContent: ", templateContent)

	template1 := template.New(service)
	logger.Info("template created")

	serviceTemplate, err := template1.Parse(templateContent)
	logger.Info("parsed template")

	if err != nil {
		return fmt.Errorf("unable to parse systemd service template: %w", err)
	}

	if _, err = os.Stat(paths.SystemdUserDir); os.IsNotExist(err) {
		logger.Info("os.IsNotExist")
		if err = os.MkdirAll(paths.SystemdUserDir, 0755); err != nil {
			return fmt.Errorf("unable to create folder '%v': %w", paths.SystemdUserDir, err)
		}
		logger.Info("mkdir")
	}

	logger.Info("dir exists")

	var targetFile *os.File
	if _, err = os.Stat(targetPath); os.IsNotExist(err) {
		targetFile, err = os.Create(targetPath)
		defer targetFile.Close()
		if err != nil {
			return fmt.Errorf("unable to create file %q: %w", targetPath, err)
		}
	} else {
		if !force {
			return fmt.Errorf("file %q already exists", targetPath)
		}

		targetFile, err = os.OpenFile(targetPath, os.O_WRONLY|os.O_TRUNC, 0666)
		defer targetFile.Close()
		if err != nil {
			return fmt.Errorf("unable to open file %q: %w", targetPath, err)
		}
	}

	logger.Info("target file exists")

	if err := serviceTemplate.Execute(targetFile, paths); err != nil {
		return fmt.Errorf("executing service template failed: %w", err)
	}

	logger.Info("created systemd service")

	return nil
}

func createKubeadmConfig(paths *useful_paths.UsefulPaths, force bool) error {
	targetPath := paths.KubeadmAdminConfig

	logger := log.WithFields(log.Fields{
		"targetPath": targetPath,
	})

	kubeadmTemplate, err := template.New(useful_paths.RelativePathKubeadmAdminConfigFile).Parse(kubeadmTemplateContent)
	if err != nil {
		return fmt.Errorf("unable to parse kubeadm template: %w", err)
	}

	var targetFile *os.File
	if _, err = os.Stat(targetPath); os.IsNotExist(err) {
		targetFile, err = os.Create(targetPath)
		defer targetFile.Close()
		if err != nil {
			return fmt.Errorf("unable to create file %q: %w", targetPath, err)
		}
	} else {
		if !force {
			return fmt.Errorf("file '%v' already exists", targetPath)
		}

		targetFile, err = os.OpenFile(targetPath, os.O_WRONLY|os.O_TRUNC, 0666)
		defer targetFile.Close()
		if err != nil {
			return fmt.Errorf("unable to open file %q: %w", targetPath, err)
		}
	}

	if err := kubeadmTemplate.Execute(targetFile, nil); err != nil {
		return fmt.Errorf("executing service template failed: %w", err)
	}

	logger.Info("created kubeadm template")
	return nil
}

func writeScript(script fs.DirEntry, targetDir string, force bool) error {
	scriptPath := path.Join(scriptsFSBasePath, script.Name())
	targetPath := path.Join(targetDir, script.Name())

	logger := log.WithFields(log.Fields{
		"sourcePath": scriptPath,
		"targetPath": targetPath,
	})

	if err := writeFile(scriptPath, targetPath, force); err != nil {
		return err
	}

	if err := os.Chmod(targetPath, 0755); err != nil {
		return fmt.Errorf("unable to change permissions for file '%v': %w", targetPath, err)
	}

	logger.Debug("copied embedded script")

	return nil
}

func writeFile(sourcePath string, targetPath string, force bool) error {
	sourceFile, err := scriptFS.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("unable to read embedded script %q: %w", sourcePath, err)
	}
	defer sourceFile.Close()

	var targetFile *os.File
	if _, err = os.Stat(targetPath); os.IsNotExist(err) {
		targetFile, err = os.Create(targetPath)
		defer targetFile.Close()
		if err != nil {
			return fmt.Errorf("unable to create file %q: %w", targetPath, err)
		}
	} else {
		if !force {
			return fmt.Errorf("file %q already exists: %w", targetPath, err)
		}
		targetFile, err = os.OpenFile(targetPath, os.O_WRONLY|os.O_TRUNC, 0666)
		defer targetFile.Close()
		if err != nil {
			return fmt.Errorf("unable to open file %q", targetPath)
		}
	}

	if _, err = io.Copy(targetFile, sourceFile); err != nil {
		return fmt.Errorf("unable to copy to destination %q: %w", targetPath, err)
	}

	return nil
}
