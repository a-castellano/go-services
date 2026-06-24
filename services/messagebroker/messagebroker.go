// Package messagebroker provides a service for managing interactions with message broker services.
package messagebroker

import (
	"context"
	logger "github.com/a-castellano/go-services/infra/logger"
)

// Client interface defines the contract for message broker operations.
// Implementations must provide methods for sending and receiving messages.
type Client interface {
	// SendMessage sends a message to the specified queue.
	// Returns an error if the operation fails.
	SendMessage(context.Context, string, []byte) error

	// ReceiveMessages continuously receives messages from the specified queue.
	// Messages are sent to the messages channel, and errors to the errors channel.
	// The operation can be stopped by canceling the context.
	ReceiveMessages(context.Context, string, chan<- []byte, chan<- error)
}

// MessageBroker provides a high-level interface for message broker operations.
// It uses a Client implementation to perform actual message broker operations.
type MessageBroker struct {
	Client Client
}

// SendMessage sends a message to the specified queue using the underlying client.
// This is a wrapper method that delegates to the client's SendMessage method.
func (messageBroker MessageBroker) SendMessage(ctx context.Context, queueName string, message []byte) error {

	log := logger.FromContext(ctx).With("operation", "SendMessage")
	log.DebugContext(ctx, "sending message through messageBroker", "queueName", queueName, "message", message)
	return messageBroker.Client.SendMessage(ctx, queueName, message)
}

// ReceiveMessages receives messages from the specified queue using the underlying client.
// This is a wrapper method that delegates to the client's ReceiveMessages method.
// The operation can be stopped by canceling the provided context.
func (messageBroker MessageBroker) ReceiveMessages(ctx context.Context, queueName string, messages chan<- []byte, errors chan<- error) {
	log := logger.FromContext(ctx).With("operation", "ReceiveMessages")
	log.DebugContext(ctx, "receiving messages from messageBroker", "queueName", queueName)

	messageBroker.Client.ReceiveMessages(ctx, queueName, messages, errors)
}
