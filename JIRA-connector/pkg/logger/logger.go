package logger

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	With(fields ...Field) Logger
	SetLevel(level Level)
}

type TestLogger struct {
}

func NewTestLogger() Logger {
	return TestLogger{}
}
func (TestLogger) Debug(msg string, fields ...Field) {

}

func (TestLogger) Info(msg string, fields ...Field) {

}

func (TestLogger) Warn(msg string, fields ...Field) {

}

func (TestLogger) Error(msg string, fields ...Field) {
}

func (TestLogger) With(fields ...Field) Logger {
	return TestLogger{}
}

func (TestLogger) SetLevel(level Level) {
}

type Field struct {
	Key   string
	Value interface{}
}
