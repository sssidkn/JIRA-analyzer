package logger

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
)

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

func Interceptor(log Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		start := time.Now()

		log.Info(fmt.Sprintf("gRPC method %s called with request: %+v", info.FullMethod, req))

		resp, err = handler(ctx, req)

		duration := time.Since(start)
		if err != nil {
			log.Error(fmt.Sprintf("gRPC method %s failed: %v (took %v)", info.FullMethod, err, duration))
		} else {
			log.Info(fmt.Sprintf("gRPC method %s completed (took %v)", info.FullMethod, duration))
		}

		return resp, err
	}
}
