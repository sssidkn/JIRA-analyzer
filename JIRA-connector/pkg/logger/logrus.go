package logger

import (
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
	return &LogrusLogger{logger: log}
}

func (l *LogrusLogger) Info(msg string, fileds ...Field) {
	l.logger.WithFields(toLogrusFields(fileds)).Info(msg)
}

func (l *LogrusLogger) Debug(msg string, fileds ...Field) {
	l.logger.WithFields(toLogrusFields(fileds)).Debug(msg)
}

func (l *LogrusLogger) Warn(msg string, fileds ...Field) {
	l.logger.WithFields(toLogrusFields(fileds)).Warn(msg)
}

func (l *LogrusLogger) Error(msg string, fileds ...Field) {
	l.logger.WithFields(toLogrusFields(fileds)).Error(msg)
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
