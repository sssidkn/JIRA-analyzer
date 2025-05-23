package logger

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	logger *logrus.Logger
}

func New(logDir string) (*logrus.Logger, error) {
	var logInstance = logrus.New()

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	logInstance.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := filepath.Base(f.File)
			return "", filename
		},
	})
	logInstance.SetReportCaller(true)

	allLogsFile, err := os.OpenFile(
		filepath.Join(logDir, "logs.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)
	if err != nil {
		return nil, err
	}

	errLogsFile, err := os.OpenFile(
		filepath.Join(logDir, "err_logs.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)
	if err != nil {
		return nil, err
	}

	logInstance.AddHook(&fileHook{
		allLogsFile: allLogsFile,
		errLogsFile: errLogsFile,
	})

	logInstance.AddHook(&consoleHook{})

	return logInstance, nil
}

type fileHook struct {
	allLogsFile *os.File
	errLogsFile *os.File
}

func (hook *fileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *fileHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}

	if _, err = hook.allLogsFile.WriteString(line); err != nil {
		return err
	}

	if entry.Level <= logrus.WarnLevel {
		if _, err = hook.errLogsFile.WriteString(line); err != nil {
			return err
		}
	}

	return nil
}

type consoleHook struct{}

func (hook *consoleHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
	}
}

func (hook *consoleHook) Fire(entry *logrus.Entry) error {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logrus.SetOutput(os.Stdout)

	switch entry.Level {
	case logrus.PanicLevel:
		logrus.Panic(entry.Message)
	case logrus.FatalLevel:
		logrus.Fatal(entry.Message)
	case logrus.ErrorLevel:
		logrus.Error(entry.Message)
	case logrus.WarnLevel:
		logrus.Warn(entry.Message)
	}

	return nil
}
