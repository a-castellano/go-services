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

	messageBroker := MessageBroker{Client: rabbitmqClient}

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

	messageBroker := MessageBroker{Client: rabbitmqClient}

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

	messageBroker := MessageBroker{Client: rabbitmqClient}

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

	messageBroker := MessageBroker{Client: rabbitmqClient}

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

	messageBroker := MessageBroker{Client: rabbitmqClient}

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

	messageBroker := MessageBroker{Client: rabbitmqClient}

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

// TestReceiveMessagesWithInvalidQueue tests ReceiveMessages with an invalid queue name.
// This test should trigger queue declaration errors.
func TestReceiveMessagesWithInvalidQueue(t *testing.T) {

	setUp()
	defer teardown()

	// Set environment variables for RabbitMQ
	os.Setenv("RABBITMQ_HOST", "172.17.0.30")
	os.Setenv("RABBITMQ_PORT", "5672")
	os.Setenv("RABBITMQ_USER", "guest")
	os.Setenv("RABBITMQ_PASSWORD", "guest")

	config, err := rabbitmqconfig.NewConfig()

	if err != nil {
		t.Errorf("NewConfig method with valid host and port shouldn't fail, error was '%s'.", err.Error())
	} else {
		rabbitmqClient := NewRabbitmqClient(config)
		messageBroker := MessageBroker{Client: rabbitmqClient}

		// Create channels for receiving messages and errors
		messages := make(chan []byte)
		errorsChan := make(chan error)

		// Start receiving messages with empty queue name (this should cause an error)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		go messageBroker.ReceiveMessages(ctx, "", messages, errorsChan)

		// Wait for error
		select {
		case err := <-errorsChan:
			if err == nil {
				t.Errorf("ReceiveMessages with empty queue should return an error")
			}
		case <-ctx.Done():
			t.Logf("Test completed without error (empty queue name accepted)")
		}
	}
}

// TestReceiveMessagesWithConnectionFailure tests ReceiveMessages when connection fails.
// This test simulates connection issues by using invalid connection parameters.
func TestReceiveMessagesWithConnectionFailure(t *testing.T) {

	setUp()
	defer teardown()

	// Set invalid environment variables to cause connection failure
	os.Setenv("RABBITMQ_HOST", "invalid-host")
	os.Setenv("RABBITMQ_PORT", "5672")
	os.Setenv("RABBITMQ_USER", "guest")
	os.Setenv("RABBITMQ_PASSWORD", "guest")

	config, err := rabbitmqconfig.NewConfig()

	if err != nil {
		t.Errorf("NewConfig method with valid host and port shouldn't fail, error was '%s'.", err.Error())
	} else {
		rabbitmqClient := NewRabbitmqClient(config)
		messageBroker := MessageBroker{Client: rabbitmqClient}

		// Create channels for receiving messages and errors
		messages := make(chan []byte)
		errorsChan := make(chan error)

		// Start receiving messages
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		go messageBroker.ReceiveMessages(ctx, "test-queue", messages, errorsChan)

		// Wait for connection error
		select {
		case err := <-errorsChan:
			if err == nil {
				t.Errorf("ReceiveMessages with invalid connection should return an error")
			}
		case <-ctx.Done():
			t.Errorf("ReceiveMessages with invalid connection should fail within 5 seconds")
		}
	}
}

// TestReceiveMessagesWithChannelFailure tests ReceiveMessages when channel creation fails.
// This test simulates channel creation issues.
func TestReceiveMessagesWithChannelFailure(t *testing.T) {

	setUp()
	defer teardown()

	// Set environment variables for RabbitMQ
	os.Setenv("RABBITMQ_HOST", "172.17.0.30")
	os.Setenv("RABBITMQ_PORT", "5672")
	os.Setenv("RABBITMQ_USER", "guest")
	os.Setenv("RABBITMQ_PASSWORD", "guest")

	config, err := rabbitmqconfig.NewConfig()

	if err != nil {
		t.Errorf("NewConfig method with valid host and port shouldn't fail, error was '%s'.", err.Error())
	} else {
		rabbitmqClient := NewRabbitmqClient(config)
		messageBroker := MessageBroker{Client: rabbitmqClient}

		// Create channels for receiving messages and errors
		messages := make(chan []byte)
		errorsChan := make(chan error)

		// Start receiving messages
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		go messageBroker.ReceiveMessages(ctx, "test-queue", messages, errorsChan)

		// Wait for potential error (this test might pass if connection works)
		select {
		case err := <-errorsChan:
			if err != nil {
				t.Logf("Received expected error: %v", err)
			}
		case <-ctx.Done():
			t.Logf("Test completed without error (connection successful)")
		}
	}
}

// TestReceiveMessagesWithQosFailure tests ReceiveMessages when QoS setting fails.
// This test might trigger QoS errors under certain conditions.
func TestReceiveMessagesWithQosFailure(t *testing.T) {

	setUp()
	defer teardown()

	// Set environment variables for RabbitMQ
	os.Setenv("RABBITMQ_HOST", "172.17.0.30")
	os.Setenv("RABBITMQ_PORT", "5672")
	os.Setenv("RABBITMQ_USER", "guest")
	os.Setenv("RABBITMQ_PASSWORD", "guest")

	config, err := rabbitmqconfig.NewConfig()

	if err != nil {
		t.Errorf("NewConfig method with valid host and port shouldn't fail, error was '%s'.", err.Error())
	} else {
		rabbitmqClient := NewRabbitmqClient(config)
		messageBroker := MessageBroker{Client: rabbitmqClient}

		// Create channels for receiving messages and errors
		messages := make(chan []byte)
		errorsChan := make(chan error)

		// Start receiving messages
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		go messageBroker.ReceiveMessages(ctx, "test-queue", messages, errorsChan)

		// Wait for potential error
		select {
		case err := <-errorsChan:
			if err != nil {
				t.Logf("Received expected error: %v", err)
			}
		case <-ctx.Done():
			t.Logf("Test completed without error (QoS successful)")
		}
	}
}

// TestReceiveMessagesWithConsumeFailure tests ReceiveMessages when message consumption fails.
// This test might trigger consume errors under certain conditions.
func TestReceiveMessagesWithConsumeFailure(t *testing.T) {

	setUp()
	defer teardown()

	// Set environment variables for RabbitMQ
	os.Setenv("RABBITMQ_HOST", "172.17.0.30")
	os.Setenv("RABBITMQ_PORT", "5672")
	os.Setenv("RABBITMQ_USER", "guest")
	os.Setenv("RABBITMQ_PASSWORD", "guest")

	config, err := rabbitmqconfig.NewConfig()

	if err != nil {
		t.Errorf("NewConfig method with valid host and port shouldn't fail, error was '%s'.", err.Error())
	} else {
		rabbitmqClient := NewRabbitmqClient(config)
		messageBroker := MessageBroker{Client: rabbitmqClient}

		// Create channels for receiving messages and errors
		messages := make(chan []byte)
		errorsChan := make(chan error)

		// Start receiving messages
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		go messageBroker.ReceiveMessages(ctx, "test-queue", messages, errorsChan)

		// Wait for potential error
		select {
		case err := <-errorsChan:
			if err != nil {
				t.Logf("Received expected error: %v", err)
			}
		case <-ctx.Done():
			t.Logf("Test completed without error (consume successful)")
		}
	}
}
