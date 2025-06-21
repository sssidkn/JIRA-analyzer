package logger

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type LogrusLogger struct {
	logger *logrus.Logger
}

func NewLogrusLogger() *LogrusLogger {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.DateTime,
		DisableSorting:  false,
		PadLevelText:    true,
	})

	log.SetOutput(os.Stdout)

	fileAllLogs, err := os.OpenFile("logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Can't open logs.log:", err)
	}

	fileErrLogs, err := os.OpenFile("err_logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Can't open err_logs.log:", err)
	}

	log.AddHook(&fileHook{
		fileAll:  fileAllLogs,
		fileWarn: fileErrLogs,
	})

	return &LogrusLogger{logger: log}
}

type fileHook struct {
	fileAll  *os.File
	fileWarn *os.File
}

func (h *fileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *fileHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	_, err = h.fileAll.WriteString(line)
	if err != nil {
		return err
	}

	if entry.Level <= logrus.WarnLevel {
		_, err = h.fileWarn.WriteString(line)
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *LogrusLogger) Info(msg string, fields ...Field) {
	l.logger.WithFields(toLogrusFields(fields)).Info(msg)
}

func (l *LogrusLogger) Debug(msg string, fields ...Field) {
	l.logger.WithFields(toLogrusFields(fields)).Debug(msg)
}

func (l *LogrusLogger) Warn(msg string, fields ...Field) {
	l.logger.WithFields(toLogrusFields(fields)).Warn(msg)
}

func (l *LogrusLogger) Error(msg string, fields ...Field) {
	l.logger.WithFields(toLogrusFields(fields)).Error(msg)
}

func (l *LogrusLogger) With(fields ...Field) Logger {
	l.logger.WithFields(toLogrusFields(fields))
	return l
}

func (l *LogrusLogger) SetLevel(level Level) {
	switch level {
	case LevelDebug:
		l.logger.SetLevel(logrus.DebugLevel)
	case LevelWarn:
		l.logger.SetLevel(logrus.WarnLevel)
	case LevelError:
		l.logger.SetLevel(logrus.ErrorLevel)
	default:
		l.logger.SetLevel(logrus.InfoLevel)
	}
}

func toLogrusFields(fields []Field) logrus.Fields {
	logrusFields := make(logrus.Fields)
	for _, field := range fields {
		logrusFields[field.Key] = field.Value
	}
	return logrusFields
}
