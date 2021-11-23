package log

import (
	"context"

	"go.uber.org/zap/zapcore"
)

type mockLogFactory struct {
}

func (f mockLogFactory) Bg() Logger {
	return mocklogger{}
}

func (b mockLogFactory) For(ctx context.Context) Logger {
	return b.Bg()
}

// With creates a child logger, and optionally adds some context fields to that logger.
func (b mockLogFactory) With(fields ...zapcore.Field) Factory {
	return mockLogFactory{}
}

func NewMockLogFactory() Factory {
	return mockLogFactory{}
}

type mocklogger struct {
}

func (l mocklogger) Info(msg string, fields ...zapcore.Field) {
}

func (l mocklogger) Error(msg string, fields ...zapcore.Field) {
}

func (l mocklogger) Fatal(msg string, fields ...zapcore.Field) {
}

func (l mocklogger) With(fields ...zapcore.Field) Logger {
	return mocklogger{}
}
