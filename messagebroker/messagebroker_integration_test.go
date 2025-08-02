//go:build integration_tests || messagebroker_tests

// Package messagebroker_integration_test contains integration tests for the messagebroker package.
// These tests require a real RabbitMQ server to be running and test actual RabbitMQ operations.
package messagebroker

import (
	"context"
	"os"
	"testing"
	"time"

	rabbitmqconfig "github.com/a-castellano/go-types/rabbitmq"
)

// TestRabbitmqFailedConnection tests RabbitMQ client connection with an invalid port.
// This test verifies that the client properly handles connection failures when the port is not accessible.
func TestRabbitmqFailedConnection(t *testing.T) {

	setUp()
	defer teardown()

	// Set environment variables for an invalid RabbitMQ configuration
	os.Setenv("RABBITMQ_HOST", "rabbitmq")
	os.Setenv("RABBITMQ_PORT", "1123")
	os.Setenv("RABBITMQ_USER", "user")
	os.Setenv("RABBITMQ_PASSWORD", "password")

	rabbitmqConfig, _ := rabbitmqconfig.NewConfig()
	queueName := "test"
	testString := []byte("This is a Test")

	rabbitmqClient := NewRabbitmqClient(rabbitmqConfig)

	messageBroker := MessageBroker{client: rabbitmqClient}

	// Test sending a message with invalid connection
	dial_error := messageBroker.SendMessage(queueName, testString)

	if dial_error == nil {
		t.Errorf("TestRabbitmqFailedConnection should fail.")
	}
}

// TestRabbitmqInvalidCredentials tests RabbitMQ client connection with invalid credentials.
// This test verifies that the client properly handles authentication failures.
func TestRabbitmqInvalidCredentials(t *testing.T) {

	setUp()
	defer teardown()

	// Set environment variables for RabbitMQ with invalid credentials
	os.Setenv("RABBITMQ_HOST", "rabbitmq")
	os.Setenv("RABBITMQ_PORT", "5672")
	os.Setenv("RABBITMQ_USER", "user")
	os.Setenv("RABBITMQ_PASSWORD", "password")

	rabbitmqConfig, _ := rabbitmqconfig.NewConfig()
	queueName := "test"
	testString := []byte("This is a Test")

	rabbitmqClient := NewRabbitmqClient(rabbitmqConfig)

	messageBroker := MessageBroker{client: rabbitmqClient}

	// Test sending a message with invalid credentials
	dial_error := messageBroker.SendMessage(queueName, testString)

	if dial_error == nil {
		t.Errorf("TestRabbitmqFailedConnection should fail.")
	} else {
		if dial_error.Error() != "Exception (403) Reason: \"username or password not allowed\"" {
			t.Errorf("TestRabbitmqFailedConnection error should be 'Exception (403) Reason: \"username or password not allowed\"' but was '%s'.", dial_error.Error())
		}
	}
}

// TestRabbitmqSendMessage tests RabbitMQ message sending with valid credentials.
// This test verifies that the client can successfully send messages to a running RabbitMQ server.
func TestRabbitmqSendMessage(t *testing.T) {

	setUp()
	defer teardown()

	// Set environment variables for RabbitMQ with valid credentials
	os.Setenv("RABBITMQ_HOST", "rabbitmq")
	os.Setenv("RABBITMQ_PORT", "5672")
	os.Setenv("RABBITMQ_USER", "guest")
	os.Setenv("RABBITMQ_PASSWORD", "guest")

	rabbitmqConfig, _ := rabbitmqconfig.NewConfig()
	queueName := "test"
	testString := []byte("This is a Test")

	rabbitmqClient := NewRabbitmqClient(rabbitmqConfig)

	messageBroker := MessageBroker{client: rabbitmqClient}

	// Test sending a message with valid configuration
	sendError := messageBroker.SendMessage(queueName, testString)

	if sendError != nil {
		t.Errorf("TestRabbitmqSendMessage shouldn't fail. Error was '%s'.", sendError.Error())
	}
}

// TestRabbitmqReceiveMessageFromInvalidConfig tests RabbitMQ message receiving with an invalid configuration.
// This test verifies that the client properly handles connection failures when receiving messages.
func TestRabbitmqReceiveMessageFromInvalidConfig(t *testing.T) {

	setUp()
	defer teardown()

	// Set environment variables for an invalid RabbitMQ configuration
	os.Setenv("RABBITMQ_HOST", "rabbitmq")
	os.Setenv("RABBITMQ_PORT", "1122")
	os.Setenv("RABBITMQ_USER", "guest")
	os.Setenv("RABBITMQ_PASSWORD", "guest")

	rabbitmqConfig, _ := rabbitmqconfig.NewConfig()
	queueName := "test"

	rabbitmqClient := NewRabbitmqClient(rabbitmqConfig)

	messageBroker := MessageBroker{client: rabbitmqClient}

	// Create channels for receiving messages and errors
	messagesReceived := make(chan []byte)
	ctx, cancel := context.WithCancel(context.Background())
	receiveErrors := make(chan error)

	// Start receiving messages in a goroutine
	go messageBroker.ReceiveMessages(ctx, queueName, messagesReceived, receiveErrors)

	// Wait for an error or timeout
	select {
	case receivedError := <-receiveErrors:
		if receivedError == nil {
			t.Errorf("TestRabbitmqReceiveMessageFromInvalidConfig should fail.")
		}
	case <-time.After(5 * time.Second):
		t.Errorf("TestRabbitmqReceiveMessageFromInvalidConfig should fail within 5 seconds.")
	}

	cancel()
}

// TestRabbitmqReceiveMessageFromInvalidCredentials tests RabbitMQ message receiving with invalid credentials.
// This test verifies that the client properly handles authentication failures when receiving messages.
func TestRabbitmqReceiveMessageFromInvalidCredentials(t *testing.T) {

	setUp()
	defer teardown()

	// Set environment variables for RabbitMQ with invalid credentials
	os.Setenv("RABBITMQ_HOST", "rabbitmq")
	os.Setenv("RABBITMQ_PORT", "5672")
	os.Setenv("RABBITMQ_USER", "user")
	os.Setenv("RABBITMQ_PASSWORD", "password")

	rabbitmqConfig, _ := rabbitmqconfig.NewConfig()
	queueName := "test"

	rabbitmqClient := NewRabbitmqClient(rabbitmqConfig)

	messageBroker := MessageBroker{client: rabbitmqClient}

	// Create channels for receiving messages and errors
	messagesReceived := make(chan []byte)
	ctx, cancel := context.WithCancel(context.Background())
	receiveErrors := make(chan error)

	// Start receiving messages in a goroutine
	go messageBroker.ReceiveMessages(ctx, queueName, messagesReceived, receiveErrors)

	// Wait for an error or timeout
	select {
	case receivedError := <-receiveErrors:
		if receivedError == nil {
			t.Errorf("TestRabbitmqReceiveMessageFromInvalidCredentials should fail.")
		} else {
			if receivedError.Error() != "Exception (403) Reason: \"username or password not allowed\"" {
				t.Errorf("TestRabbitmqReceiveMessageFromInvalidCredentials error should be 'Exception (403) Reason: \"username or password not allowed\"' but was '%s'.", receivedError.Error())
			}
		}
	case <-time.After(5 * time.Second):
		t.Errorf("TestRabbitmqReceiveMessageFromInvalidCredentials should fail within 5 seconds.")
	}

	cancel()
}

// TestRabbitmqReceiveMessage tests RabbitMQ message receiving with valid credentials.
// This test verifies that the client can successfully receive messages from a running RabbitMQ server.
func TestRabbitmqReceiveMessage(t *testing.T) {

	setUp()
	defer teardown()

	// Set environment variables for RabbitMQ with valid credentials
	os.Setenv("RABBITMQ_HOST", "rabbitmq")
	os.Setenv("RABBITMQ_PORT", "5672")
	os.Setenv("RABBITMQ_USER", "guest")
	os.Setenv("RABBITMQ_PASSWORD", "guest")

	rabbitmqConfig, _ := rabbitmqconfig.NewConfig()
	queueName := "test"

	rabbitmqClient := NewRabbitmqClient(rabbitmqConfig)

	messageBroker := MessageBroker{client: rabbitmqClient}

	// Create channels for receiving messages and errors
	messagesReceived := make(chan []byte)
	ctx, cancel := context.WithCancel(context.Background())
	receiveErrors := make(chan error)

	// Start receiving messages in a goroutine
	go messageBroker.ReceiveMessages(ctx, queueName, messagesReceived, receiveErrors)

	// Wait for a message or error
	select {
	case receivedError := <-receiveErrors:
		if receivedError != nil {
			t.Errorf("TestRabbitmqReceiveMessage shouldn't fail, error was '%s'.", receivedError.Error())
		}
	case <-time.After(5 * time.Second):
		t.Logf("TestRabbitmqReceiveMessage didn't receive any message in 5 seconds, which is expected if no messages are sent to the queue.")
	}

	cancel()
}
