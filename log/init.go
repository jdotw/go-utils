package log

import (
	"math/rand"
	"time"

	"github.com/uber/jaeger-lib/metrics"
	"github.com/uber/jaeger-lib/metrics/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Init(service string) (Factory, metrics.Factory) {

	rand.Seed(int64(time.Now().Nanosecond()))
	rootLogger, _ := zap.NewDevelopment(
		zap.AddStacktrace(zapcore.FatalLevel),
		zap.AddCallerSkip(1),
	)

	serviceLogger := rootLogger.With(zap.String("service", service))

	metricsFactory := prometheus.New().Namespace(metrics.NSOptions{Name: service, Tags: nil})

	return NewFactory(serviceLogger), metricsFactory
}
