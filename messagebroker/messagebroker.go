package messagebroker

import (
	"context"

	rabbitmqconfig "github.com/a-castellano/go-types/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Client interface operates against memory database instance
type Client interface {
	SendMessage(string, []byte) error
	ReceiveMessages(context.Context, string, chan<- []byte, chan<- error)
}

// RabbitmqClient is the real Redis client, it has Client methods
type RabbitmqClient struct {
	config *rabbitmqconfig.Config
}

func NewRabbimqClient(rabbitmqConfig *rabbitmqconfig.Config) RabbitmqClient {
	reabbitmqClient := RabbitmqClient{config: rabbitmqConfig}
	return reabbitmqClient
}

// SendMessage sends a message through queueName
func (client RabbitmqClient) SendMessage(queueName string, message []byte) error {

	conn, errDial := amqp.Dial(client.config.ConnectionString)
	if errDial != nil {
		return errDial
	}
	defer conn.Close()

	channel, errChannel := conn.Channel()

	if errChannel != nil {
		return errChannel
	}

	defer channel.Close()

	queue, errQueue := channel.QueueDeclare(

		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments

	)

	if errQueue != nil {
		return errQueue
	}

	// send message

	err := channel.Publish(

		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         message,
		})

	if err != nil {
		return err
	}

	return nil
}

// ReceiveMessages stores messages in channel until it is closed using context
func (client RabbitmqClient) ReceiveMessages(ctx context.Context, queueName string, messages chan<- []byte, errors chan<- error) {

	conn, errDial := amqp.Dial(client.config.ConnectionString)

	if errDial != nil {
		errors <- errDial
		return
	}

	defer conn.Close()

	channel, errChannel := conn.Channel()

	if errChannel != nil {
		errors <- errChannel
		return
	}

	_, errQueue := channel.QueueDeclare(
		queueName,
		true,  // Durable
		false, // DeleteWhenUnused
		false, // Exclusive
		false, // NoWait
		nil,   // arguments
	)

	if errQueue != nil {
		errors <- errQueue
		return
	}

	errChannelQos := channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)

	if errChannelQos != nil {
		errors <- errChannelQos
		return
	}

	messagesToReceive, errMessageReceived := channel.Consume(
		queueName,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)

	if errMessageReceived != nil {
		errors <- errMessageReceived
		return
	}

	for receivedMessage := range messagesToReceive {

		messages <- receivedMessage.Body
		select {
		case <-ctx.Done(): //exit function
			errors <- nil
			return
		default:
			continue
		}
	}

	errors <- nil
}

// MessageBroker uses Client in order to operate against messagebroker instance
type MessageBroker struct {
	client Client
}

// SendMessage sends a message through queueName using Client
func (messageBroker MessageBroker) SendMessage(queueName string, message []byte) error {
	return messageBroker.client.SendMessage(queueName, message)
}

// ReceiveMessages receives messages through queueName using Client
func (messageBroker MessageBroker) ReceiveMessages(ctx context.Context, queueName string, messages chan<- []byte, errors chan<- error) {
	messageBroker.client.ReceiveMessages(ctx, queueName, messages, errors)
}
