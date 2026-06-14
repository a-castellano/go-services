// Package logger provides a driver for slog.
package logger

import (
	"context"
	"log/slog"
)

// Client, the interface contract

type Client interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
}

// We define a private context ctxKey
// It is not a good idea to use a string as a key in context.WithValue, because it can lead to collisions. Instead, we define a private type ctxKey and use it as the key in context.WithValue. This way, we can be sure that our key is unique and won't collide with other keys.
type ctxKey struct{}

// WithLogger stores the logger in the context. Called once, at startup.
func WithLogger(ctx context.Context, l Client) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

var defaultLogger Client = slog.Default()

// FromContext retrieves the logger. It MUST always return a usable value,
// never nil: if the context has no logger, it returns a default fallback.
// This single guard is what prevents nil-pointer panics and is the critical
// part of the "logger in context" approach.
func FromContext(ctx context.Context) Client {
	if l, ok := ctx.Value(ctxKey{}).(Client); ok {
		return l
	}
	return defaultLogger // This cannot be nil. This is why is defined.
}
