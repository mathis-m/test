package slurm

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

type LogHook struct{}

func (hook *LogHook) Fire(entry *logrus.Entry) error {
	msg, err := entry.String()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to read log entry: %v", err)
		return err
	}

	switch entry.Level {
	case logrus.ErrorLevel:
	case logrus.FatalLevel:
		WriteError(msg)
		break
	case logrus.InfoLevel:
	default:
		WriteInfo(msg)
		break
	}

	return nil
}

func (hook *LogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
