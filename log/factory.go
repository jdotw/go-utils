package log

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Factory interface {
	Bg() Logger
	For(ctx context.Context) Logger
	With(fields ...zapcore.Field) Factory
}

func NewFactory(logger *zap.Logger) Factory {
	return factory{logger: logger}
}

type factory struct {
	logger *zap.Logger
}

// Bg creates a context-unaware logger.
func (b factory) Bg() Logger {
	return logger(b)
}

// For returns a context-aware Logger. If the context
// contains an OpenTracing span, all logging calls are also
// echo-ed into the span.
func (b factory) For(ctx context.Context) Logger {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		logger := spanLogger{span: span, logger: b.logger}

		if jaegerCtx, ok := span.Context().(jaeger.SpanContext); ok {
			logger.spanFields = []zapcore.Field{
				zap.String("trace_id", jaegerCtx.TraceID().String()),
				zap.String("span_id", jaegerCtx.SpanID().String()),
			}
		}

		return logger
	}
	return b.Bg()
}

// With creates a child logger, and optionally adds some context fields to that logger.
func (b factory) With(fields ...zapcore.Field) Factory {
	return factory{logger: b.logger.With(fields...)}
}
