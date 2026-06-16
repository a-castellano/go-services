//go:build integration_tests || rabbitmq_tests

// Package rabbitmq_integration_test contains integration tests for the rabbitmq package.
// These tests require a real RabbitMQ server to be running and test actual RabbitMQ operations.
package rabbitmq

import (
	"context"
	"os"
	"testing"
	"time"

	rabbitmqconfig "github.com/a-castellano/go-types/rabbitmq"
)

// Global variables to store original environment variable values
// These are used to restore the environment after tests that modify them
var currentHost string
var currentHostDefined bool

var currentPort string
var currentPortDefined bool

var currentUser string
var currentUserDefined bool

var currentPassword string
var currentPasswordDefined bool

// setUp saves the current environment variables and clears them for testing.
// This ensures tests start with a clean environment and can set their own values.
func setUp() {

	// Save and clear RABBITMQ_HOST environment variable
	if envHost, found := os.LookupEnv("RABBITMQ_HOST"); found {
		currentHost = envHost
		currentHostDefined = true
	} else {
		currentHostDefined = false
	}

	// Save and clear RABBITMQ_PORT environment variable
	if envPort, found := os.LookupEnv("RABBITMQ_PORT"); found {
		currentPort = envPort
		currentPortDefined = true
	} else {
		currentPortDefined = false
	}

	// Save and clear RABBITMQ_USER environment variable
	if envUser, found := os.LookupEnv("RABBITMQ_USER"); found {
		currentUser = envUser
		currentUserDefined = true
	} else {
		currentUserDefined = false
	}

	// Save and clear RABBITMQ_PASSWORD environment variable
	if envPassword, found := os.LookupEnv("RABBITMQ_PASSWORD"); found {
		currentPassword = envPassword
		currentPasswordDefined = true
	} else {
		currentPasswordDefined = false
	}

	// Clear all RabbitMQ environment variables for a clean test state
	os.Unsetenv("RABBITMQ_HOST")
	os.Unsetenv("RABBITMQ_PORT")
	os.Unsetenv("RABBITMQ_USER")
	os.Unsetenv("RABBITMQ_PASSWORD")

}

// teardown restores the original environment variables that were saved in setUp.
// This ensures that tests don't affect the environment for other tests or processes.
func teardown() {

	// Restore RABBITMQ_HOST environment variable
	if currentHostDefined {
		os.Setenv("RABBITMQ_HOST", currentHost)
	} else {
		os.Unsetenv("RABBITMQ_HOST")
	}

	// Restore RABBITMQ_PORT environment variable
	if currentPortDefined {
		os.Setenv("RABBITMQ_PORT", currentPort)
	} else {
		os.Unsetenv("RABBITMQ_PORT")
	}

	// Restore RABBITMQ_USER environment variable
	if currentUserDefined {
		os.Setenv("RABBITMQ_USER", currentUser)
	} else {
		os.Unsetenv("RABBITMQ_USER")
	}

	// Restore RABBITMQ_PASSWORD environment variable
	if currentPasswordDefined {
		os.Setenv("RABBITMQ_PASSWORD", currentPassword)
	} else {
		os.Unsetenv("RABBITMQ_PASSWORD")
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

	// Test sending a message with invalid credentials
	dial_error := rabbitmqClient.SendMessage(context.Background(), queueName, testString)

	if dial_error == nil {
		t.Errorf("TestRabbitmqInvalidCredentials should fail.")
	} else {
		expectedError := "Exception (403) Reason: \"username or password not allowed\""
		if dial_error.Error() != expectedError {
			t.Fatalf("Expected error '%s' but got '%s'", expectedError, dial_error.Error())
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

	// Test sending a message with valid configuration
	sendError := rabbitmqClient.SendMessage(context.Background(), queueName, testString)

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

	// Create channels for receiving messages and errors
	messagesReceived := make(chan []byte)
	ctx, cancel := context.WithCancel(context.Background())
	receiveErrors := make(chan error)

	// Start receiving messages in a goroutine
	go rabbitmqClient.ReceiveMessages(ctx, queueName, messagesReceived, receiveErrors)

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

	// Create channels for receiving messages and errors
	messagesReceived := make(chan []byte)
	ctx, cancel := context.WithCancel(context.Background())
	receiveErrors := make(chan error)

	// Start receiving messages in a goroutine
	go rabbitmqClient.ReceiveMessages(ctx, queueName, messagesReceived, receiveErrors)

	// Wait for an error or timeout
	select {
	case receivedError := <-receiveErrors:
		if receivedError == nil {
			t.Errorf("TestRabbitmqReceiveMessageFromInvalidCredentials should fail.")
		} else {
			expectedError := "Exception (403) Reason: \"username or password not allowed\""
			if receivedError.Error() != expectedError {
				t.Fatalf("Expected error '%s' but got '%s'", expectedError, receivedError.Error())
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

	// Create channels for receiving messages and errors
	messagesReceived := make(chan []byte)
	ctx, cancel := context.WithCancel(context.Background())
	receiveErrors := make(chan error)

	// Start receiving messages in a goroutine
	go rabbitmqClient.ReceiveMessages(ctx, queueName, messagesReceived, receiveErrors)

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

// TestRabbitmqReceiveMessageEmpty tests RabbitMQ message receiving when context cancels the method
func TestRabbitmqReceiveMessageEmpty(t *testing.T) {

	setUp()
	defer teardown()

	// Set environment variables for RabbitMQ with valid credentials
	os.Setenv("RABBITMQ_HOST", "rabbitmq")
	os.Setenv("RABBITMQ_PORT", "5672")
	os.Setenv("RABBITMQ_USER", "guest")
	os.Setenv("RABBITMQ_PASSWORD", "guest")

	rabbitmqConfig, _ := rabbitmqconfig.NewConfig()
	queueName := "empty"

	rabbitmqClient := NewRabbitmqClient(rabbitmqConfig)

	// Create channels for receiving messages and errors
	messagesReceived := make(chan []byte)
	ctx, cancel := context.WithCancel(context.Background())
	receiveErrors := make(chan error)

	// Start receiving messages in a goroutine
	go rabbitmqClient.ReceiveMessages(ctx, queueName, messagesReceived, receiveErrors)

	// Stop message receiver
	cancel()
	// Wait for a message or error
	select {
	case receivedError := <-receiveErrors:
		if receivedError != nil {
			t.Errorf("TestRabbitmqReceiveMessage with canceled broker shouldn't fail, error was '%s'.", receivedError.Error())
		}
	case <-time.After(5 * time.Second):
		t.Logf("TestRabbitmqReceiveMessage didn't receive any message in 5 seconds, which is expected if no messages are sent to the queue.")
	}

}
