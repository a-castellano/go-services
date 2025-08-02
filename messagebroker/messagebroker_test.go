//go:build integration_tests || unit_tests || messagebroker_tests || messagebroker_unit_tests

// Package messagebroker_test contains unit tests for the messagebroker package.
// Tests cover both RabbitMQ client functionality and MessageBroker wrapper operations.
package messagebroker

import (
	"context"
	//	rabbitmqconfig "github.com/a-castellano/go-types/rabbitmq"
	"errors"
	"os"
	"testing"
	"time"
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

	// Clear all RabbitMQ environment variables for clean test state
	os.Unsetenv("RABBITMQ_HOST")
	os.Unsetenv("RABBITMQ_PORT")
	os.Unsetenv("RABBITMQ_DATABASE")
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

// RabbitmqMock implements the Client interface for testing purposes.
// It simulates RabbitMQ operations without requiring a real RabbitMQ server.
type RabbitmqMock struct {
	LaunchError bool // Flag to control whether operations should fail
}

// SendMessage simulates sending a message through RabbitMQ.
// If LaunchError is true, it returns an error; otherwise, it succeeds.
func (client RabbitmqMock) SendMessage(queueName string, message []byte) error {
	if client.LaunchError {
		return errors.New("Error")
	}
	return nil
}

// ReceiveMessages simulates receiving messages from RabbitMQ.
// It sends a test message to the messages channel and then exits.
// If LaunchError is true, it sends an error to the errors channel.
func (client RabbitmqMock) ReceiveMessages(ctx context.Context, queueName string, messages chan<- []byte, errorsChan chan<- error) {
	if client.LaunchError {
		errorsChan <- errors.New("Error")
		return
	}
	// Send a test message and then exit
	messages <- []byte("Test message")
	errorsChan <- nil
}

// TestSendMessageWithMockFailedSendMessage tests SendMessage operation when the underlying
// RabbitMQ SendMessage operation fails. It uses a mocked client that simulates a failure.
func TestSendMessageWithMockFailedSendMessage(t *testing.T) {

	setUp()
	defer teardown()

	// Create a mock client that will fail
	rabbitmqMock := RabbitmqMock{LaunchError: true}
	messageBroker := MessageBroker{Client: rabbitmqMock}

	// Test sending a message with a failing mock
	err := messageBroker.SendMessage("test", []byte("test"))

	if err == nil {
		t.Errorf("messageBroker.SendMessage with failing mock should fail.")
	} else {
		if err.Error() != "Error" {
			t.Errorf("messageBroker.SendMessage call with failing mock should return error \"Error\", it has returned \"%s\"", err.Error())
		}
	}
}

// TestSendMessageWithMockSendMessage tests SendMessage operation when the underlying
// RabbitMQ SendMessage operation succeeds. It uses a mocked client that simulates success.
func TestSendMessageWithMockSendMessage(t *testing.T) {

	setUp()
	defer teardown()

	// Create a mock client that will succeed
	rabbitmqMock := RabbitmqMock{LaunchError: false}
	messageBroker := MessageBroker{Client: rabbitmqMock}

	// Test sending a message with a successful mock
	err := messageBroker.SendMessage("test", []byte("test"))

	if err != nil {
		t.Errorf("messageBroker.SendMessage with successful mock should not fail.")
	}
}

// TestReceiveMessageWithMockFailure tests ReceiveMessages operation when the underlying
// RabbitMQ ReceiveMessages operation fails. It uses a mocked client that simulates a failure.
func TestReceiveMessageWithMockFailure(t *testing.T) {

	setUp()
	defer teardown()

	// Create a mock client that will fail
	rabbitmqMock := RabbitmqMock{LaunchError: true}
	messageBroker := MessageBroker{Client: rabbitmqMock}

	// Create channels for receiving messages and errors
	messages := make(chan []byte)
	errorsChan := make(chan error)

	// Start receiving messages in a goroutine
	ctx := context.Background()
	go messageBroker.ReceiveMessages(ctx, "test", messages, errorsChan)

	// Wait for the error
	select {
	case err := <-errorsChan:
		if err == nil {
			t.Errorf("messageBroker.ReceiveMessages with failing mock should return an error.")
		} else {
			if err.Error() != "Error" {
				t.Errorf("messageBroker.ReceiveMessages call with failing mock should return error \"Error\", it has returned \"%s\"", err.Error())
			}
		}
	case <-time.After(5 * time.Second):
		t.Errorf("messageBroker.ReceiveMessages with failing mock should return an error within 5 seconds.")
	}
}

// TestReceiveMessageWithMock tests ReceiveMessages operation when the underlying
// RabbitMQ ReceiveMessages operation succeeds. It uses a mocked client that simulates success.
func TestReceiveMessageWithMock(t *testing.T) {

	setUp()
	defer teardown()

	// Create a mock client that will succeed
	rabbitmqMock := RabbitmqMock{LaunchError: false}
	messageBroker := MessageBroker{Client: rabbitmqMock}

	// Create channels for receiving messages and errors
	messages := make(chan []byte)
	errorsChan := make(chan error)

	// Start receiving messages in a goroutine
	ctx := context.Background()
	go messageBroker.ReceiveMessages(ctx, "test", messages, errorsChan)

	// Wait for a message or error
	select {
	case message := <-messages:
		if string(message) != "Test message" {
			t.Errorf("messageBroker.ReceiveMessages with successful mock should return \"Test message\", it has returned \"%s\"", string(message))
		}
	case err := <-errorsChan:
		if err != nil {
			t.Errorf("messageBroker.ReceiveMessages with successful mock should not return an error.")
		}
	case <-time.After(5 * time.Second):
		t.Errorf("messageBroker.ReceiveMessages with successful mock should return a message within 5 seconds.")
	}
}
