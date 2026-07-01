// Package opentelemetry provides a driver for opentelemetry.
package opentelemetry

import (
	"context"
	"errors"

	opentelemetryconfig "github.com/a-castellano/go-types/types/opentelemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
)

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// newTracerProvider builds a TracerProvider whose Resource carries the
// service.name taken from config.AppName, registering the given span
// processors. Taking the processors as a parameter is the seam that lets a
// test inject an in-memory recorder (tracetest) instead of the stdout
// exporter the production path uses.
func newTracerProvider(config *opentelemetryconfig.Config, processors ...trace.SpanProcessor) *trace.TracerProvider {
	resourceAttributes := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(config.AppName),
	)

	opts := []trace.TracerProviderOption{trace.WithResource(resourceAttributes)}
	for _, processor := range processors {
		opts = append(opts, trace.WithSpanProcessor(processor))
	}

	return trace.NewTracerProvider(opts...)
}

func SetupOpenTelemetry(ctx context.Context, config *opentelemetryconfig.Config) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error
	var err error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	if !config.Enabled {
		return shutdown, err
	}
	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider. The stdout exporter is the part that can fail, so
	// it is built here; on failure we return the no-op shutdown plus the error
	// so main can log it and keep running.
	traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return shutdown, err
	}
	tracerProvider := newTracerProvider(config, trace.NewBatchSpanProcessor(traceExporter))

	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	return shutdown, err
}
