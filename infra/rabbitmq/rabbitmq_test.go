//go:build integration_tests || unit_tests || rabbitmq_tests || rabbitmq_unit_tests

// Build tags: this file only compiles (and its tests only run) when one of these
// tags is passed, e.g. `go test -tags unit_tests`. A plain `go test` skips it.

// Unit tests for the rabbitmq package. They exercise RabbitmqClient against
// in-memory fakes (no real broker), relying on the injected DialFunc to swap the
// real connection for a configurable fake one.
package rabbitmq

import (
	"errors"
	rabbitmqconfig "github.com/a-castellano/go-types/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
	"testing"
)

// fakeConnection satisfies the AMQPConnection interface without touching a real
// broker. Each fail* flag lets a test force the corresponding method to error,
// and the flags meant for the channel are forwarded to the fakeChannel it creates
// in Channel(). deliveries holds the messages a test wants Consume() to emit.
type fakeConnection struct {
	failClose        bool // make Close() return an error
	failChannel      bool // make Channel() return an error
	failChannelClose bool // forwarded to the channel's Close()
	failQueueDeclare bool // forwarded to the channel's QueueDeclare()
	failPublish      bool // forwarded to the channel's Publish()
	failQos          bool // forwarded to the channel's Qos()
	failConsume      bool // forwarded to the channel's Consume()
	deliveries       []amqp.Delivery
}

// Close can simulate failures
func (f *fakeConnection) Close() error {
	if f.failClose {
		return errors.New("Fatal error closing connection")
	} else {
		return nil
	}
}

// Channel returns a fakeChannel pre-configured with the connection's fail flags
// and deliveries, or an error when failChannel is set.
func (f *fakeConnection) Channel() (AMQPChannel, error) {
	if f.failChannel {
		return nil, errors.New("Fatal error opening channel")
	}
	return &fakeChannel{
		failClose:        f.failChannelClose,
		failQueueDeclare: f.failQueueDeclare,
		failPublish:      f.failPublish,
		failQos:          f.failQos,
		failConsume:      f.failConsume,
		deliveries:       f.deliveries,
	}, nil
}

// fakeChannel satisfies AMQPChannel interface.
type fakeChannel struct {
	failClose        bool
	failQueueDeclare bool
	failPublish      bool
	failQos          bool
	failConsume      bool
	deliveries       []amqp.Delivery
}

// Close can simulate failures
func (f *fakeChannel) Close() error {
	if f.failClose {
		return errors.New("Fatal error closing channel")
	} else {
		return nil
	}
}

// QueueDeclare is a stub that returns a dummy queue and no error.
func (f *fakeChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	if f.failQueueDeclare {
		return amqp.Queue{}, errors.New("Fatal error declaring queue")
	} else {
		return amqp.Queue{Name: name}, nil
	}
}

// Publish is a stub that mocks publishing errors.
func (f *fakeChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	if f.failPublish {
		return errors.New("Fatal error publishing message")
	} else {
		return nil
	}
}

// Qos is a stub that mocks erors.
func (f *fakeChannel) Qos(prefetchCount, prefetchSize int, global bool) error {
	if f.failQos {
		return errors.New("Fatal error setting QoS")
	} else {
		return nil
	}
}

// Consume returns a buffered channel pre-loaded with the configured deliveries,
// then closes it. Closing is essential: ReceiveMessages drains the channel with a
// `for range`, which only ends once the channel is closed. Without close() the
// consumer would block forever.
func (f *fakeChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	if f.failConsume {
		return nil, errors.New("Fatal error consuming messages")
	} else {

		ch := make(chan amqp.Delivery, len(f.deliveries))
		for _, d := range f.deliveries {
			ch <- d
		}
		close(ch)
		return ch, nil
	}
}

// dialReturning builds a DialFunc that always hands back the given connection.
// Each test configures its own *fakeConnection (deliveries, fail flags) and
// passes it here, so the injected dialer carries exactly that behaviour.
func dialReturning(conn *fakeConnection) DialFunc {
	return func(url string) (AMQPConnection, error) {
		return conn, nil
	}
}

// dialFailing builds a DialFunc that always fails, simulating amqp.Dial returning
// an error (e.g. the broker is unreachable).
func dialFailing() DialFunc {
	return func(url string) (AMQPConnection, error) {
		return nil, errors.New("Fatal error: dial failed")
	}
}

// TestSendMessageFailDial verifies that when the dial step fails, SendMessage
// surfaces that error and does not proceed.
func TestSendMessageFailDial(t *testing.T) {

	rabbitmqConfig := rabbitmqconfig.Config{}

	client := RabbitmqClient{config: &rabbitmqConfig, dial: dialFailing()}

	err := client.SendMessage("test-queue", []byte("test message"))

	if err == nil {
		t.Fatal("Dial mocks a fail but SendMessage did not return an error")
	}
	expectedError := "Fatal error: dial failed"
	if err.Error() != expectedError {
		t.Fatalf("Expected error '%s' but got '%s'", expectedError, err.Error())
	}

}

// TestSendMessageFailConnChannel verifies that when opening the channel fails
// (dial succeeds, conn.Channel() errors), SendMessage returns that error.
func TestSendMessageFailConnChannel(t *testing.T) {

	rabbitmqConfig := rabbitmqconfig.Config{}

	conn := &fakeConnection{failChannel: true}
	client := RabbitmqClient{config: &rabbitmqConfig, dial: dialReturning(conn)}

	err := client.SendMessage("test-queue", []byte("test message"))

	if err == nil {
		t.Fatal("Dial mocks a fail but SendMessage did not return an error")
	}
	expectedError := "Fatal error opening channel"
	if err.Error() != expectedError {
		t.Fatalf("Expected error '%s' but got '%s'", expectedError, err.Error())
	}

}
