//go:build integration_tests || unit_tests || logger_tests || logger_unit_tests

// Unit tests for the logger package.
package logger

import (
	"bytes"
	slogconfig "github.com/a-castellano/go-types/slog"
	"log/slog"
	"testing"
)

func TestLogSomething(t *testing.T) {
	var buf bytes.Buffer
	config := slogconfig.Config{DefaultLevel: slog.LevelDebug, Format: "plain", AddSource: false, AppName: "LogSomething"}

	testLogger := newSlogLogger(&buf, &config)

	testLogger.Info("hello")

	bufferLen := buf.Len()

	if bufferLen <= 0 {
		t.Errorf("TestLogSomething fail, buffer is empty")
	}
}
