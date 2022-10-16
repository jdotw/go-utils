package tracing

import (
	"context"
	"fmt"
	"time"

	"github.com/jdotw/go-utils/log"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/rpcmetrics"
	"github.com/uber/jaeger-lib/metrics"
	"go.uber.org/zap"
)

func Init(serviceName string, metricsFactory metrics.Factory, logger log.Factory) opentracing.Tracer {
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{},
	}
	cfg.ServiceName = serviceName
	cfg.Sampler.Type = "const"
	cfg.Sampler.Param = 1

	_, err := cfg.FromEnv()
	if err != nil {
		logger.Bg().Fatal("Failed to parse tracer env vars", zap.Error(err))
	}

	time.Sleep(100 * time.Millisecond)
	jaegerLogger := jaegerLoggerAdapter{logger.Bg()}

	metricsFactory = metricsFactory.Namespace(metrics.NSOptions{Name: serviceName, Tags: nil})
	tracer, _, err := cfg.NewTracer(
		config.Logger(jaegerLogger),
		config.Metrics(metricsFactory),
		config.Observer(rpcmetrics.NewObserver(metricsFactory, rpcmetrics.DefaultNameNormalizer)),
	)
	if err != nil {
		logger.Bg().Fatal("Failed to initialize tracer", zap.Error(err))
	}
	return tracer
}

type jaegerLoggerAdapter struct {
	logger log.Logger
}

func (l jaegerLoggerAdapter) Error(msg string) {
	l.logger.Error(msg)
}

func (l jaegerLoggerAdapter) Infof(msg string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(msg, args...))
}

func NewChildSpanAndContext(ctx context.Context, tracer opentracing.Tracer, name string) (context.Context, opentracing.Span) {
	var span opentracing.Span
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		span = tracer.StartSpan(
			name,
			opentracing.ChildOf(parentSpan.Context()),
		)
	} else {
		span = tracer.StartSpan(name)
	}
	ctx = opentracing.ContextWithSpan(ctx, span)
	return ctx, span
}
