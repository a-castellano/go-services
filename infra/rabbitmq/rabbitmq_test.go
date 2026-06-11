//go:build integration_tests || unit_tests || rabbitmq_tests || rabbitmq_unit_tests

// Tests cover both RabbitMQ client functionality.
package rabbitmq

import (
	"context"
	"errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
	"strings"
	"testing"
	"time"
)

// fakeConnection satisfies AMQPConnection interface.
type fakeConnection struct {
	failClose        bool
	failChannel      bool
	failChannelClose bool
	failQueueDeclare bool
	failPublish      bool
	failQos          bool
	failConsume      bool
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

// Channel
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

// Consume is a stub that returns a dummy channel and no error.
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

func fakeDial(url string) (AMQPConnection, error) {
	// need to add failures and behaviour
	fakeConn := fakeConnection{failClose: false}
	return &fakeConn, nil
}
