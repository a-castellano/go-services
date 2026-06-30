//go:build integration_tests || unit_tests || opentelemetry_tests || opentelemetry_unit_tests

// Unit tests for the opentelemetry package.
package opentelemetry

import (
	"context"
	"testing"

	opentelemetryconfig "github.com/a-castellano/go-types/types/opentelemetry"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func TestTelemetryDisabled(t *testing.T) {

	config := opentelemetryconfig.Config{Enabled: false}
	ctx := context.Background()

	shutdown, err := SetupOpenTelemetry(ctx, &config)

	if err != nil {
		t.Errorf("TestTelemetryDisabled should not fail, error was '%s'.", err.Error())
	} else {

		if shutdown == nil {
			t.Errorf("TestTelemetryDisabled shutdown function should not be nil as telemetry is disabled")
		}

		shutdownReturn := shutdown(ctx)
		if shutdownReturn != nil {
			t.Errorf("TestTelemetryDisabled shutdown should be nil as telemetry is disabled")
		}
	}

}

func TestTraceProvides(t *testing.T) {

	config := opentelemetryconfig.Config{Enabled: true, AppName: "Test"}

	// The recorder is itself a SpanProcessor. We inject it through the seam in
	// newTracerProvider so the spans land in memory as structs we can inspect,
	// instead of being exported as JSON to stdout the way the production path
	// (SetupOpenTelemetry) does.
	spanRecorder := tracetest.NewSpanRecorder()
	tracerProvider := newTracerProvider(&config, spanRecorder)

	_, span := tracerProvider.Tracer("test").Start(context.Background(), "test-span")
	span.End()

	spans := spanRecorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected exactly 1 recorded span, got %d", len(spans))
	}

	recorded := spans[0]
	if recorded.Name() != "test-span" {
		t.Errorf("expected span name %q, got %q", "test-span", recorded.Name())
	}

	serviceName, ok := recorded.Resource().Set().Value(semconv.ServiceNameKey)
	if !ok {
		t.Fatalf("recorded span resource has no %q attribute", semconv.ServiceNameKey)
	}
	if serviceName.AsString() != config.AppName {
		t.Errorf("expected service.name %q, got %q", config.AppName, serviceName.AsString())
	}
}

func TestPropagatorRoundTrip(t *testing.T) {

	propagator := newPropagator()

	// Build a context carrying a known, valid span context so the propagator
	// has a real traceparent to inject. The ids are arbitrary but valid hex.
	traceID, _ := oteltrace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	spanID, _ := oteltrace.SpanIDFromHex("0102030405060708")
	spanContext := oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: oteltrace.FlagsSampled,
		Remote:     true,
	})
	ctx := oteltrace.ContextWithSpanContext(context.Background(), spanContext)

	// Inject: the composite propagator writes the W3C keys into the carrier map.
	carrier := propagation.MapCarrier{}
	propagator.Inject(ctx, carrier)

	if carrier["traceparent"] == "" {
		t.Fatalf("expected propagator to inject a traceparent, carrier was %v", carrier)
	}

	// Extract: round-trips the carrier back into a context whose span context
	// carries the same ids. If the W3C TraceContext propagator were not wired,
	// the extracted span context would be empty and these would not match.
	extracted := oteltrace.SpanContextFromContext(propagator.Extract(context.Background(), carrier))
	if extracted.TraceID() != traceID {
		t.Errorf("expected extracted trace id %s, got %s", traceID, extracted.TraceID())
	}
	if extracted.SpanID() != spanID {
		t.Errorf("expected extracted span id %s, got %s", spanID, extracted.SpanID())
	}
}

func TestTelemetryEnabled(t *testing.T) {

	config := opentelemetryconfig.Config{Enabled: true, AppName: "Test"}
	ctx := context.Background()

	shutdown, err := SetupOpenTelemetry(ctx, &config)

	if err != nil {
		t.Errorf("TestTelemetryEnabled should not fail, error was '%s'.", err.Error())
	} else {
		shutdownReturn := shutdown(ctx)
		if shutdownReturn != nil {
			t.Errorf("TestTelemetryEnabled shutdown should be nil as telemetry is disabled")
		}
	}

}
