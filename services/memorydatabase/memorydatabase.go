// Package memorydatabase provides a service for managing interactions with memory databases.
// Currently supports Redis as the primary memory database implementation.
package memorydatabase

import (
	"context"
	"errors"
	logger "github.com/a-castellano/go-services/infra/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Client interface defines the contract for memory database operations.
// Implementations must provide methods for reading and writing string values,
// as well as checking if the client has been properly initialized.
type Client interface {
	// WriteString stores a string value with the given key and TTL (time-to-live) in seconds.
	// Returns an error if the operation fails.
	WriteString(context.Context, string, string, int) error

	// ReadString retrieves a string value by key.
	// Returns the value, a boolean indicating if the key was found, and any error.
	ReadString(context.Context, string) (string, bool, error)

	// IsClientInitiated returns true if the client has been successfully initialized.
	IsClientInitiated() bool
}

// MemoryDatabase provides a high-level interface for memory database operations.
// It uses a Client implementation to perform actual database operations.
type MemoryDatabase struct {
	client Client
}

// NewMemoryDatabase creates a new MemoryDatabase instance with the provided client.
// The client must implement the Client interface.
func NewMemoryDatabase(client Client) MemoryDatabase {
	memorydatabase := MemoryDatabase{client: client}

	return memorydatabase
}

// WriteString writes a string value to the memory database using the underlying client.
// This is a wrapper method that checks if the client is initialized before performing the operation.
// Returns an error if the client is not initialized or if the write operation fails.
func (memorydatabase *MemoryDatabase) WriteString(ctx context.Context, key string, value string, ttl int) error {

	// Start span
	ctx, span := otel.Tracer("github.com/a-castellano/go-services/services/memorydatabase").Start(ctx, "WriteString")
	defer span.End()

	span.SetAttributes(
		attribute.String("key", key),
	)

	log := logger.FromContext(ctx).With("operation", "WriteString")
	log.DebugContext(ctx, "checking if memorydatabase client is initiated")

	if memorydatabase.client.IsClientInitiated() {
		log.DebugContext(ctx, "writing into memorydatabase", "key", key, "value", value)
		return memorydatabase.client.WriteString(ctx, key, value, ttl)
	} else {
		errorString := "memorydatabase client is not initiated, cannot perform WriteString operation"
		errNotInitiated := errors.New(errorString)
		span.RecordError(errNotInitiated)
		span.SetStatus(codes.Error, errorString)

		log.ErrorContext(ctx, errorString)
		return errNotInitiated
	}
}

// ReadString reads a string value from the memory database using the underlying client.
// This is a wrapper method that checks if the client is initialized before performing the operation.
// Returns the value, a boolean indicating if the key was found, and any error.
// Returns an error if the client is not initialized.
func (memorydatabase *MemoryDatabase) ReadString(ctx context.Context, key string) (string, bool, error) {

	// Start span
	ctx, span := otel.Tracer("github.com/a-castellano/go-services/services/memorydatabase").Start(ctx, "WriteString")
	defer span.End()

	span.SetAttributes(
		attribute.String("key", key),
	)

	log := logger.FromContext(ctx).With("operation", "ReadString")
	log.DebugContext(ctx, "checking if memorydatabase client is initiated")

	if memorydatabase.client.IsClientInitiated() {
		log.DebugContext(ctx, "reading from memorydatabase", "key", key)
		return memorydatabase.client.ReadString(ctx, key)
	} else {
		errorString := "memorydatabase client is not initiated, cannot perform ReadString operation"
		errNotInitiated := errors.New(errorString)
		span.RecordError(errNotInitiated)
		span.SetStatus(codes.Error, errorString)

		log.ErrorContext(ctx, errorString)
		return "", false, errNotInitiated
	}
}
