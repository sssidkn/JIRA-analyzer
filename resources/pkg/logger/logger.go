package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	logger *logrus.Logger
}

func New(logDir string) (*Logger, error) {
	var logInstance = logrus.New()
	logInstance.SetOutput(io.Discard)

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

	return &Logger{logInstance}, nil
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
	line, err := entry.String()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(os.Stdout, line)
	return err
}

func Middleware(l *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		endTime := time.Now()
		latency := endTime.Sub(startTime)
		statusCode := c.Writer.Status()

		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path

		query := c.Request.URL.RawQuery
		if query != "" {
			path = path + "?" + query
		}

		entry := l.logger.WithFields(logrus.Fields{
			"status":     statusCode,
			"method":     method,
			"path":       path,
			"ip":         clientIP,
			"latency":    latency,
			"user-agent": c.Request.UserAgent(),
		})

		if statusCode >= 400 {
			entry.Error("HTTP Request Error")
		} else {
			entry.Info("HTTP Request")
		}
	}
}

func (l *Logger) Error(args ...interface{}) {
	l.logger.Error(args)
}
func (l *Logger) Info(args ...interface{}) {
	l.logger.Info(args)
}
func (l *Logger) Debug(args ...interface{}) {
	l.logger.Debug(args)
}
func (l *Logger) Warn(args ...interface{}) {
	l.logger.Warn(args)
}
func (l *Logger) Fatal(args ...interface{}) {
	l.logger.Fatal(args)
}
