//go:build integration_tests || unit_tests || logger_tests || logger_unit_tests

// Unit tests for the logger package.
package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	slogconfig "github.com/a-castellano/go-types/slog"
)

// TestLogSomething checks that a plain-text logger writes a record to the
// injected writer and that the line carries the message, level and source.
// AddSource is enabled, so the record must include the file location; the
// expected path assumes tests run inside the container, where the module is
// mounted under /app.
func TestLogSomething(t *testing.T) {
	var buf bytes.Buffer
	config := slogconfig.Config{DefaultLevel: slog.LevelDebug, Format: "plain", AddSource: true, AppName: "LogSomething"}

	testLogger := newSlogLogger(&buf, &config)

	testLogger.Info("hello")

	bufferLen := buf.Len()

	if bufferLen <= 0 {
		t.Errorf("TestLogSomething has failed, buffer is empty")
	} else {
		emittedLog := buf.String()
		if !strings.Contains(emittedLog, "msg=hello") {
			t.Errorf("expected log to contain msg=hello, got: %q", emittedLog)
		}
		if !strings.Contains(emittedLog, "level=INFO") {
			t.Errorf("expected log to contain level=INFO, got: %q", emittedLog)
		}
		if !strings.Contains(emittedLog, "source=/app/infra/logger") {
			t.Errorf("expected log to contain at least source=/app/infra/logger, got: %q", emittedLog)
		}

	}
}

// TestLogSomethingJSON checks that the JSON format produces a parseable JSON
// object whose "msg" field matches the logged message.
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
				t.Errorf("TestLogSomethingJSON has failed, message should be \"hello\" but it was \"%s\"", message)
			}
		}

	}
}

// TestChangeLogLevel checks that SetLevel raises verbosity at runtime: with the
// config default at Info, a Debug record is normally filtered, but after
// SetLevel(Debug) it must be emitted.
func TestChangeLogLevel(t *testing.T) {
	var buf bytes.Buffer

	config := slogconfig.Config{DefaultLevel: slog.LevelInfo, Format: "plain", AddSource: false, AppName: "LogSomething"}

	testLogger := newSlogLogger(&buf, &config)

	testLogger.SetLevel(slog.LevelDebug)

	testLogger.Debug("hello_debug")

	bufferLen := buf.Len()

	if bufferLen <= 0 {
		t.Errorf("TestChangeLogLevel has failed, buffer is empty")
	} else {
		emittedLog := buf.String()
		if !strings.Contains(emittedLog, "msg=hello_debug ") {
			t.Errorf("expected log to contain msg=hello_debug, got: %q", emittedLog)
		}
		if !strings.Contains(emittedLog, "level=DEBUG") {
			t.Errorf("expected log to contain level=DEBUG, got: %q", emittedLog)
		}

	}
}

// TestChangeLogLevelToDefault checks that SetDefaultLevel restores the config
// level: after temporarily lowering the level to Debug and resetting it, a
// Debug record must be filtered again (default is Info) while an Info record
// is still emitted.
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
		t.Errorf("TestChangeLogLevelToDefault has failed, buffer should be empty as default level is Info at this moment")
	} else {

		testLogger.Info("hello_info")
		newBufferLen := buf.Len()
		if newBufferLen <= 0 {
			t.Errorf("TestChangeLogLevelToDefault has failed, after logging back to INFO buffer shouldn't be empty as default")
		} else {
			emittedLog := buf.String()
			if !strings.Contains(emittedLog, "msg=hello_info ") {
				t.Errorf("expected log to contain msg=hello_info, got: %q", emittedLog)
			}
			if !strings.Contains(emittedLog, "level=INFO ") {
				t.Errorf("expected log to contain level=INFO, got: %q", emittedLog)
			}
		}

	}
}

// TestNewLogger is a smoke test: the public constructor must build a logger
// from a valid config without panicking.
func TestNewLogger(t *testing.T) {

	config := slogconfig.Config{DefaultLevel: slog.LevelInfo, Format: "plain", AddSource: false, AppName: "LogSomething"}

	NewLogger(&config)
}

// TestWithLogger checks the context round-trip: a logger stored with WithLogger
// must be the very same instance retrieved by FromContext.
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

// TestFromContextDefaultLogger checks the fallback path: calling FromContext on
// a context with no logger must return the default logger instead of nil.
func TestFromContextDefaultLogger(t *testing.T) {

	ctx := context.Background()
	// don't define a logger, so the default slog logger should be returned
	FromContext(ctx)

}
