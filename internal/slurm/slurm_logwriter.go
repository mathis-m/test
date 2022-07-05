package slurm

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

type LogHook struct{}

func (hook *LogHook) Fire(entry *logrus.Entry) error {
	msg, err := entry.String()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to read log entry: %v", err)
		return err
	}

	msg = strings.TrimRight(msg, "\r\n")

	switch entry.Level {
	case logrus.ErrorLevel, logrus.FatalLevel:
		WriteError(msg)
		break
	default:
		WriteInfo(msg)
		break
	}

	return nil
}

func (hook *LogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
