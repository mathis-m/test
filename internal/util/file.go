package util

import (
	"errors"
	"fmt"
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
