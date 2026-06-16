//go:build integration_tests || unit_tests || logger_tests || logger_unit_tests

// Unit tests for the logger package.
package logger

import (
	"bytes"
	"context"
	"encoding/json"
	slogconfig "github.com/a-castellano/go-types/slog"
	"log/slog"
	"strings"
	"testing"
)

func TestLogSomething(t *testing.T) {
	var buf bytes.Buffer
	config := slogconfig.Config{DefaultLevel: slog.LevelDebug, Format: "plain", AddSource: true, AppName: "LogSomething"}

	testLogger := newSlogLogger(&buf, &config)

	testLogger.Info("hello")

	bufferLen := buf.Len()

	if bufferLen <= 0 {
		t.Errorf("TestLogSomething has failed, buffer is empty")
	} else {
		emitedLog := buf.String()
		if !strings.Contains(emitedLog, "msg=hello") {
			t.Errorf("expected log to contain msg=hello, got: %q", emitedLog)
		}
		if !strings.Contains(emitedLog, "level=INFO") {
			t.Errorf("expected log to contain level=INFO, got: %q", emitedLog)
		}
		if !strings.Contains(emitedLog, "source=/app/infra/logger") {
			t.Errorf("expected log to contain at least source=/app/infra/logger, got: %q", emitedLog)
		}

	}
}

func TestLogSomethingJSON(t *testing.T) {
	var buf bytes.Buffer
	config := slogconfig.Config{DefaultLevel: slog.LevelDebug, Format: "JSON", AddSource: true, AppName: "LogSomething"}

	testLogger := newSlogLogger(&buf, &config)

	testLogger.Info("hello")

	bufferLen := buf.Len()

	if bufferLen <= 0 {
		t.Errorf("TestLogSomethingJSON has failed, buffer is empty")
	} else {
		var loggedData map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &loggedData); err != nil {
			t.Errorf("TestLogSomethingJSON has failed, cannot unmarshal json log")
		} else {
			message := loggedData["msg"].(string)
			if message != "hello" {
				t.Errorf("TestLogSomethingJSON has failed, message shoulf be \"hello\" but it was \"%s\"", message)
			}
		}

	}
}

func TestChangeLogLevel(t *testing.T) {
	var buf bytes.Buffer

	config := slogconfig.Config{DefaultLevel: slog.LevelInfo, Format: "plain", AddSource: false, AppName: "LogSomething"}

	testLogger := newSlogLogger(&buf, &config)

	testLogger.SetLevel(slog.LevelDebug)

	testLogger.Debug("hello_debug")

	bufferLen := buf.Len()

	if bufferLen <= 0 {
		t.Errorf("TestLogSomething has failed, buffer is empty")
	} else {
		emitedLog := buf.String()
		if !strings.Contains(emitedLog, "msg=hello_debug ") {
			t.Errorf("expected log to contain msg=hello, got: %q", emitedLog)
		}
		if !strings.Contains(emitedLog, "level=DEBUG") {
			t.Errorf("expected log to contain level=DEBUG, got: %q", emitedLog)
		}

	}
}

func TestChangeLogLevelToDefault(t *testing.T) {
	var buf bytes.Buffer

	config := slogconfig.Config{DefaultLevel: slog.LevelInfo, Format: "plain", AddSource: false, AppName: "LogSomething"}

	testLogger := newSlogLogger(&buf, &config)

	testLogger.SetLevel(slog.LevelDebug)

	// back to default level
	testLogger.SetDefaultLevel()

	testLogger.Debug("hello_debug")

	bufferLen := buf.Len()

	if bufferLen > 0 {
		t.Errorf("TestChangeLogLevelToDefault has failed, buffer should be empty as defult level is Info at this moment")
	} else {

		testLogger.Info("hello_info")
		newBufferLen := buf.Len()
		if newBufferLen <= 0 {
			t.Errorf("TestChangeLogLevelToDefault has failed, after logging back to INFO buffer shouldn't be empty as defult")
		} else {
			emitedLog := buf.String()
			if !strings.Contains(emitedLog, "msg=hello_info ") {
				t.Errorf("expected log to contain msg=hello_info, got: %q", emitedLog)
			}
			if !strings.Contains(emitedLog, "level=INFO ") {
				t.Errorf("expected log to contain level=INFO, got: %q", emitedLog)
			}
		}

	}
}

func TestNewLogger(t *testing.T) {

	config := slogconfig.Config{DefaultLevel: slog.LevelInfo, Format: "plain", AddSource: false, AppName: "LogSomething"}

	NewLogger(&config)
}

func TestWithLogger(t *testing.T) {
	var buf bytes.Buffer
	cfg := slogconfig.Config{DefaultLevel: slog.LevelInfo, Format: "plain", AppName: "x"}
	contextLogger := newSlogLogger(&buf, &cfg)

	// Add logger to context
	ctx := WithLogger(context.Background(), contextLogger)

	retrievedLogger := FromContext(ctx)

	if retrievedLogger != contextLogger {
		t.Errorf("FromContext returned a different logger than the one stored")
	}
}

func TestFromContextDefaultLogger(t *testing.T) {

	ctx := context.Background()
	// don't define logger, so a defult slog should be returned
	FromContext(ctx)

}
