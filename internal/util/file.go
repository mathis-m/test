package util

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func WriteStringToFile(content string, path string) error {
	targetDir := filepath.Dir(path)

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
			return errors.New(fmt.Sprint("Unable to create \"", targetDir, "\" directory:", err))
		}
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		if err := os.Remove(path); err != nil {
			return errors.New(fmt.Sprint("Unable to delete \"", path, "\" file", err))
		}
	}

	userConfig, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = userConfig.Close() }()

	if _, err = userConfig.WriteString(content); err != nil {
		return err
	}

	return nil
}

func DeleteFileIfExists(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("unexpected error: %w", err)
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("unable to remove file: %w", err)
	}

	return nil
}

func EnsureFolderExistsWithPermissions(path string, perm os.FileMode) error {
	needToCreateDir := false

	stats, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			needToCreateDir = true
		} else {
			return fmt.Errorf("ensure folder: %w", err)
		}
	}

	if stats != nil && !stats.IsDir() {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("ensure folder: unable to delete file %q: %w", path, err)
		}
		needToCreateDir = true
	}

	if needToCreateDir {
		if err := os.MkdirAll(path, perm); err != nil {
			return fmt.Errorf("ensure folder: unable to create directory at %q: %w", path, err)
		}
	} else {
		if stats.Mode() != perm {
			if err := os.Chmod(path, perm); err != nil {
				return fmt.Errorf("ensure folder: unable to chmod dir %q: %q", path, err)
			}
		}
	}

	return nil
}

func CopyFileSafe(sourcePath string, targetPath string) error {
	return CopyFile(sourcePath, targetPath, false, 0666)
}

func CopyFileOverwrite(sourcePath string, targetPath string) error {
	return CopyFile(sourcePath, targetPath, true, 0666)
}

func CopyFile(sourcePath string, targetPath string, overwrite bool, perm os.FileMode) error {
	if _, err := os.Stat(sourcePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("unable to copy: source file at %q does not exist", sourcePath)
		}

		return fmt.Errorf("unable to copy: %w", err)
	}

	if _, err := os.Stat(targetPath); err == nil {
		if overwrite {
			if err := os.Remove(targetPath); err != nil {
				return fmt.Errorf("unable to copy: failed to already existing delete target file: %w", err)
			}
		} else {
			return fmt.Errorf("unable to copy: target file at %q already exists", targetPath)
		}
	}

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("unable to copy: unable to open source file %q: %w", sourcePath, err)
	}
	defer sourceFile.Close()

	targetFile, err := os.OpenFile(targetPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("unable to copy: unable to create target file %q: %w", targetPath, err)
	}
	defer targetFile.Close()

	if _, err := io.Copy(targetFile, sourceFile); err != nil {
		return fmt.Errorf("unable to copy: %w", err)
	}

	return nil
}
