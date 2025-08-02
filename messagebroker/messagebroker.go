// Package messagebroker provides a service for managing interactions with message broker services.
// Currently supports RabbitMQ as the primary message broker implementation.
package messagebroker

import (
	"context"

	rabbitmqconfig "github.com/a-castellano/go-types/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Client interface defines the contract for message broker operations.
// Implementations must provide methods for sending and receiving messages.
type Client interface {
	// SendMessage sends a message to the specified queue.
	// Returns an error if the operation fails.
	SendMessage(string, []byte) error

	// ReceiveMessages continuously receives messages from the specified queue.
	// Messages are sent to the messages channel, and errors to the errors channel.
	// The operation can be stopped by canceling the context.
	ReceiveMessages(context.Context, string, chan<- []byte, chan<- error)
}

// RabbitmqClient implements the Client interface for RabbitMQ operations.
// It manages the connection to a RabbitMQ server and provides methods for sending/receiving messages.
type RabbitmqClient struct {
	config *rabbitmqconfig.Config // RabbitMQ configuration (connection string, etc.)
}

// NewRabbitmqClient creates a new RabbitmqClient instance with the provided configuration.
func NewRabbitmqClient(rabbitmqConfig *rabbitmqconfig.Config) RabbitmqClient {
	rabbitmqClient := RabbitmqClient{config: rabbitmqConfig}
	return rabbitmqClient
}

// SendMessage sends a message to the specified queue in RabbitMQ.
// The message is sent as persistent to ensure it survives server restarts.
// Returns an error if the connection fails or if the message cannot be published.
func (client RabbitmqClient) SendMessage(queueName string, message []byte) error {

	// Establish connection to RabbitMQ server
	conn, errDial := amqp.Dial(client.config.ConnectionString)
	if errDial != nil {
		return errDial
	}
	defer conn.Close()

	// Create a channel for communication
	channel, errChannel := conn.Channel()

	if errChannel != nil {
		return errChannel
	}

	defer channel.Close()

	// Declare the queue to ensure it exists
	// Parameters: name, durable, delete when unused, exclusive, no-wait, arguments
	queue, errQueue := channel.QueueDeclare(

		queueName, // name
		true,      // durable - queue survives server restarts
		false,     // delete when unused
		false,     // exclusive - only accessible by the connection that created it
		false,     // no-wait - don't wait for server confirmation
		nil,       // arguments

	)

	if errQueue != nil {
		return errQueue
	}

	// Publish the message to the queue
	err := channel.Publish(

		"",         // exchange - empty string for default exchange
		queue.Name, // routing key - queue name
		false,      // mandatory - don't require immediate delivery
		false,      // immediate - don't require immediate delivery
		amqp.Publishing{
			DeliveryMode: amqp.Persistent, // Make message persistent
			ContentType:  "text/plain",    // Set content type
			Body:         message,         // Message body
		})

	if err != nil {
		return err
	}

	return nil
}

// ReceiveMessages continuously receives messages from the specified queue.
// Messages are sent to the messages channel, and errors to the errors channel.
// The operation can be stopped by canceling the provided context.
// This method runs until the context is canceled or an error occurs.
func (client RabbitmqClient) ReceiveMessages(ctx context.Context, queueName string, messages chan<- []byte, errors chan<- error) {

	// Establish connection to RabbitMQ server
	conn, errDial := amqp.Dial(client.config.ConnectionString)

	if errDial != nil {
		errors <- errDial
		return
	}

	defer conn.Close()

	// Create a channel for communication
	channel, errChannel := conn.Channel()

	if errChannel != nil {
		errors <- errChannel
		return
	}

	// Declare the queue to ensure it exists
	// Parameters: name, durable, delete when unused, exclusive, no-wait, arguments
	_, errQueue := channel.QueueDeclare(
		queueName,
		true,  // Durable - queue survives server restarts
		false, // DeleteWhenUnused
		false, // Exclusive - only accessible by the connection that created it
		false, // NoWait - don't wait for server confirmation
		nil,   // arguments
	)

	if errQueue != nil {
		errors <- errQueue
		return
	}

	// Set quality of service - process one message at a time
	errChannelQos := channel.Qos(
		1,     // prefetch count - number of messages to prefetch
		0,     // prefetch size - size of messages to prefetch (0 = unlimited)
		false, // global - apply to all channels on this connection
	)

	if errChannelQos != nil {
		errors <- errChannelQos
		return
	}

	// Start consuming messages from the queue
	// Parameters: queue, consumer, auto-ack, exclusive, no-local, no-wait, args
	messagesToReceive, errMessageReceived := channel.Consume(
		queueName,
		"",    // consumer - empty string for auto-generated consumer name
		true,  // auto-ack - automatically acknowledge messages
		false, // exclusive - not exclusive to this connection
		false, // no-local - don't receive messages published by this connection
		false, // no-wait - don't wait for server confirmation
		nil,   // args
	)

	if errMessageReceived != nil {
		errors <- errMessageReceived
		return
	}

	// Process incoming messages
	for receivedMessage := range messagesToReceive {

		// Send the message body to the messages channel
		messages <- receivedMessage.Body

		// Check if context has been canceled
		select {
		case <-ctx.Done(): // Exit function if context is canceled
			errors <- nil
			return
		default:
			continue
		}
	}

	errors <- nil
}

// MessageBroker provides a high-level interface for message broker operations.
// It uses a Client implementation to perform actual message broker operations.
type MessageBroker struct {
	client Client
}

// SendMessage sends a message to the specified queue using the underlying client.
// This is a wrapper method that delegates to the client's SendMessage method.
func (messageBroker MessageBroker) SendMessage(queueName string, message []byte) error {
	return messageBroker.client.SendMessage(queueName, message)
}

// ReceiveMessages receives messages from the specified queue using the underlying client.
// This is a wrapper method that delegates to the client's ReceiveMessages method.
// The operation can be stopped by canceling the provided context.
func (messageBroker MessageBroker) ReceiveMessages(ctx context.Context, queueName string, messages chan<- []byte, errors chan<- error) {
	messageBroker.client.ReceiveMessages(ctx, queueName, messages, errors)
}
