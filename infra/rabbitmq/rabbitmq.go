// Package rabbitmq provides a driver for RabbitMQ message broker.
package rabbitmq

import (
	"context"

	logger "github.com/a-castellano/go-services/infra/logger"
	rabbitmqconfig "github.com/a-castellano/go-types/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// AMQPConnection abstracts the subset of *amqp.Connection methods this package
// relies on. Because the underlying library exposes a concrete struct
// (*amqp.Connection) rather than an interface, we cannot substitute it in tests.
// Defining our own interface lets us inject a real connection in production and a
// fake one in unit tests.
//
// Note that Channel() returns AMQPChannel (our interface) instead of the concrete
// *amqp.Channel. This is what keeps the whole chain mockable: a fake connection
// can return a fake channel.
type AMQPConnection interface {
	Channel() (AMQPChannel, error)
	Close() error
}

// AMQPChannel abstracts the subset of *amqp.Channel methods this package relies on.
// Each method mirrors the exact signature of its counterpart in amqp091-go so that
// the concrete channel can satisfy this interface through a thin wrapper, and so
// that the business code does not need to change when switching from concrete
// types to interfaces.
type AMQPChannel interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	Close() error
	Qos(prefetchCount, prefetchSize int, global bool) error
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
}

// DialFunc is the type of the function used to establish a connection. It is the
// injection point of the whole design: production code passes realDial, while
// tests pass a fake dialer that returns a fake AMQPConnection without ever
// touching a real RabbitMQ server.
type DialFunc func(url string) (AMQPConnection, error)

// RabbitmqClient implements a type for RabbitMQ operations.
// It manages the connection to a RabbitMQ server and provides methods for sending/receiving messages.
type RabbitmqClient struct {
	config *rabbitmqconfig.Config // RabbitMQ configuration (connection string, etc.)
	dial   DialFunc               // Injected connection factory; realDial in production, a fake in tests
}

// realConnection wraps a concrete *amqp.Connection so it satisfies the
// AMQPConnection interface. The wrapper holds no logic of its own: each method
// simply delegates to the embedded connection.
type realConnection struct {
	conn *amqp.Connection
}

// Close delegates to the underlying connection.
func (r *realConnection) Close() error {
	return r.conn.Close()
}

// realChannel wraps a concrete *amqp.Channel so it satisfies the AMQPChannel
// interface. As with realConnection, every method just forwards the call.
type realChannel struct {
	ch *amqp.Channel
}

// Close delegates to the underlying channel.
func (r *realChannel) Close() error {
	return r.ch.Close()
}

// QueueDeclare delegates to the underlying channel.
func (r *realChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	return r.ch.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)
}

// Publish delegates to the underlying channel.
func (r *realChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return r.ch.Publish(exchange, key, mandatory, immediate, msg)
}

// Qos delegates to the underlying channel.
func (r *realChannel) Qos(prefetchCount, prefetchSize int, global bool) error {
	return r.ch.Qos(prefetchCount, prefetchSize, global)
}

// Consume delegates to the underlying channel.
func (r *realChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	return r.ch.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
}

// Channel opens a real AMQP channel and wraps it in a realChannel so the caller
// receives an AMQPChannel. Returning the interface (not the concrete type) is what
// allows the rest of the package to stay decoupled from amqp091-go.
func (r *realConnection) Channel() (AMQPChannel, error) {
	ch, err := r.conn.Channel()
	if err != nil {
		return nil, err
	}
	realCh := realChannel{ch: ch}
	return &realCh, nil
}

// realDial is the production implementation of DialFunc. It opens a genuine
// connection to RabbitMQ and wraps it so the caller receives an AMQPConnection.
func realDial(url string) (AMQPConnection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	realConn := realConnection{conn: conn}
	return &realConn, nil
}

// NewRabbitmqClient creates a new RabbitmqClient instance with the provided
// configuration. It wires realDial as the connection factory, so production code
// talks to a real broker. Tests can build a RabbitmqClient directly with a fake
// dial function instead of using this constructor.
func NewRabbitmqClient(rabbitmqConfig *rabbitmqconfig.Config) RabbitmqClient {
	rabbitmqClient := RabbitmqClient{config: rabbitmqConfig, dial: realDial}
	return rabbitmqClient
}

// SendMessage sends a message to the specified queue in RabbitMQ.
// The message is sent as persistent to ensure it survives server restarts.
// Returns an error if the connection fails or if the message cannot be published.
func (client RabbitmqClient) SendMessage(ctx context.Context, queueName string, message []byte) error {

	log := logger.FromContext(ctx)
	// Establish connection to RabbitMQ server
	log.DebugContext(ctx, "dialing connection to RabbitMQ before sending a message", "rabbitmqConfig", client.config, "operation", "send")
	conn, errDial := client.dial(client.config.ConnectionString)
	if errDial != nil {
		log.ErrorContext(ctx, "RabbitMQ dial before sending a message has failed", "operation", "send", "error", errDial.Error())
		return errDial
	}

	log.DebugContext(ctx, "dialing connection to RabbitMQ has succeeded", "operation", "send")
	defer conn.Close()

	// Create a channel for communication
	log.DebugContext(ctx, "opening RabbitMQ channel", "operation", "send")
	channel, errChannel := conn.Channel()

	if errChannel != nil {
		log.ErrorContext(ctx, "an error has occurred during RabbitMQ channel opening", "operation", "send", "error", errChannel.Error())
		return errChannel
	}
	log.DebugContext(ctx, "opening RabbitMQ channel has succeeded", "operation", "send")

	defer channel.Close()

	// Declare the queue to ensure it exists
	// Parameters: name, durable, delete when unused, exclusive, no-wait, arguments
	log.DebugContext(ctx, "declaring RabbitMQ queue", "operation", "send", "queueName", queueName)
	queue, errQueue := channel.QueueDeclare(

		queueName, // name
		true,      // durable - queue survives server restarts
		false,     // delete when unused
		false,     // exclusive - only accessible by the connection that created it
		false,     // no-wait - don't wait for server confirmation
		nil,       // arguments

	)

	if errQueue != nil {
		log.ErrorContext(ctx, "error declaring RabbitMQ queue", "operation", "send", "queueName", queueName, "error", errQueue.Error())
		return errQueue
	}

	log.DebugContext(ctx, "declaring RabbitMQ queue has succeeded", "operation", "send", "queueName", queueName)
	// Publish the message to the queue
	log.DebugContext(ctx, "publishing RabbitMQ message", "operation", "send", "queueName", queueName)
	err := channel.Publish(

		"",         // exchange - empty string for default exchange
		queue.Name, // routing key - queue name
		false,      // mandatory - don't require the message to be routable to a queue
		false,      // immediate - don't require immediate delivery to a consumer
		amqp.Publishing{
			DeliveryMode: amqp.Persistent, // Make message persistent
			ContentType:  "text/plain",    // Set content type
			Body:         message,         // Message body
		})

	if err != nil {
		log.ErrorContext(ctx, "error publishing RabbitMQ message", "operation", "send", "queueName", queueName, "error", err.Error())
		return err
	}

	log.DebugContext(ctx, "publishing RabbitMQ message has succeeded", "operation", "send", "queueName", queueName)
	return nil
}

// ReceiveMessages continuously receives messages from the specified queue.
// Messages are sent to the messages channel, and errors to the errors channel.
// The operation can be stopped by canceling the provided context.
// This method runs until the context is canceled or an error occurs.
func (client RabbitmqClient) ReceiveMessages(ctx context.Context, queueName string, messages chan<- []byte, errors chan<- error) {

	log := logger.FromContext(ctx)

	// Establish connection to RabbitMQ server
	log.DebugContext(ctx, "dialing connection to RabbitMQ before receiving messages", "rabbitmqConfig", client.config, "operation", "receive")
	conn, errDial := client.dial(client.config.ConnectionString)

	if errDial != nil {
		errors <- errDial
		log.ErrorContext(ctx, "RabbitMQ dial before receiving messages has failed", "operation", "receive", "error", errDial.Error())
		return
	}

	log.DebugContext(ctx, "dialing connection to RabbitMQ before receiving messages has succeeded", "operation", "receive")
	defer conn.Close()

	log.DebugContext(ctx, "opening RabbitMQ channel", "operation", "receive")
	// Create a channel for communication
	channel, errChannel := conn.Channel()

	if errChannel != nil {
		errors <- errChannel
		log.ErrorContext(ctx, "an error has occurred during RabbitMQ channel opening", "operation", "receive", "error", errChannel.Error())
		return
	}

	log.DebugContext(ctx, "opening RabbitMQ channel has succeeded", "operation", "receive")

	// Declare the queue to ensure it exists
	log.DebugContext(ctx, "declaring RabbitMQ queue", "operation", "receive", "queueName", queueName)
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
		log.ErrorContext(ctx, "error declaring RabbitMQ queue", "operation", "receive", "queueName", queueName, "error", errQueue.Error())
		return
	}

	log.DebugContext(ctx, "declaring RabbitMQ queue has succeeded", "operation", "receive", "queueName", queueName)
	// Set quality of service - process one message at a time
	log.DebugContext(ctx, "setting RabbitMQ QoS", "operation", "receive")
	errChannelQos := channel.Qos(
		1,     // prefetch count - number of messages to prefetch
		0,     // prefetch size - size of messages to prefetch (0 = unlimited)
		false, // global - apply to all channels on this connection
	)

	if errChannelQos != nil {
		errors <- errChannelQos
		log.ErrorContext(ctx, "error setting channel QoS", "operation", "receive", "error", errChannelQos.Error())
		return
	}

	// Start consuming messages from the queue
	log.DebugContext(ctx, "start consuming messages from queue", "operation", "receive", "queueName", queueName)
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
		log.ErrorContext(ctx, "error receiving RabbitMQ message", "operation", "receive", "queueName", queueName, "error", errMessageReceived.Error())
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
			log.DebugContext(ctx, "exiting from message consuming function", "operation", "receive")
			return
		default:
			continue
		}
	}

}
