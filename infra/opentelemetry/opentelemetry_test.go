//go:build integration_tests || unit_tests || opentelemetry_tests || opentelemetry_unit_tests

// Unit tests for the opentelemetry package.
package opentelemetry

import (
	"context"
	opentelemetryconfig "github.com/a-castellano/go-types/types/opentelemetry"
	"testing"
)

func TestTelemetryDisabled(t *testing.T) {

	config := opentelemetryconfig.Config{Enabled: false}
	ctx := context.Background()

	shutdown, err := SetupOpenTelemetry(ctx, &config)

	if err != nil {
		t.Errorf("TestTelemetryDisabled should not fail, error was '%s'.", err.Error())
	} else {
		shutdownReturn := shutdown(ctx)
		if shutdownReturn != nil {
			t.Errorf("TestTelemetryDisabled shutdown should be nil as telemetry is disabled")
		}
	}

}
