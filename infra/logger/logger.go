// Package logger provides a driver for slog.
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"

	slogconfig "github.com/a-castellano/go-types/slog"
)

// Client, the interface contract

type Client interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	// With returns a derived Client that adds the given attributes to every
	// record. It returns Client (not *slog.Logger) so derived loggers keep
	// satisfying this contract and can be stored back into the context.
	With(args ...any) Client
}

// We define a private context ctxKey
// It is not a good idea to use a string as a key in context.WithValue, because it can lead to collisions. Instead, we define a private type ctxKey and use it as the key in context.WithValue. This way, we can be sure that our key is unique and won't collide with other keys.
type ctxKey struct{}

// WithLogger stores the logger in the context. Called once, at startup.
func WithLogger(ctx context.Context, l Client) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// defaultLogger is the fallback returned by FromContext when the context
// carries no logger. It wraps slog.Default() in a *SlogLogger so it satisfies
// the Client interface (including With, which slog.Default() alone does not).
var defaultLogger Client = &SlogLogger{
	Logger:       slog.Default(),
	level:        new(slog.LevelVar),
	defaultLevel: slog.LevelInfo,
}

// FromContext retrieves the logger. It MUST always return a usable value,
// never nil: if the context has no logger, it returns a default fallback.
// This single guard is what prevents nil-pointer panics and is the critical
// part of the "logger in context" approach.
func FromContext(ctx context.Context) Client {
	if l, ok := ctx.Value(ctxKey{}).(Client); ok {
		return l
	}
	return defaultLogger // This cannot be nil. This is why it is defined.
}

type SlogLogger struct {
	*slog.Logger
	level        *slog.LevelVar
	defaultLevel slog.Level
}

func (s *SlogLogger) SetLevel(l slog.Level) {
	s.level.Set(l)
}

func (s *SlogLogger) SetDefaultLevel() {
	s.level.Set(s.defaultLevel)
}

// With returns a new Client that includes the given attributes in each log.
// We need to define it explicitly: the With method promoted from the embedded
// *slog.Logger returns *slog.Logger, which does not satisfy the Client
// interface. This wrapper preserves the level state and returns a Client.
func (s *SlogLogger) With(args ...any) Client {
	return &SlogLogger{
		Logger:       s.Logger.With(args...),
		level:        s.level,
		defaultLevel: s.defaultLevel,
	}
}

// We need to test logging, that is why we inject an io.Writer.
// The public NewLogger calls this with os.Stdout.
func newSlogLogger(writer io.Writer, config *slogconfig.Config) *SlogLogger {
	// Create a new logger with the provided configuration
	var logger *slog.Logger
	var programLevel = new(slog.LevelVar)
	if config.Format == "JSON" {
		logger = slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{AddSource: config.AddSource, Level: programLevel})).With(slog.String("app", config.AppName))
	} else {
		// format is plain text
		logger = slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{AddSource: config.AddSource, Level: programLevel})).With(slog.String("app", config.AppName))
	}
	programLevel.Set(config.DefaultLevel)

	return &SlogLogger{
		Logger:       logger,
		level:        programLevel,
		defaultLevel: config.DefaultLevel,
	}
}

func NewLogger(config *slogconfig.Config) *SlogLogger {
	return newSlogLogger(os.Stdout, config)
}
