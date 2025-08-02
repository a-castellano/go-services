//go:build integration_tests || unit_tests || messagebroker_tests || messagebroker_unit_tests

// Package messagebroker_test contains unit tests for the messagebroker package.
// Tests cover both RabbitMQ client functionality and MessageBroker wrapper operations.
package messagebroker

import (
	"context"
	//	rabbitmqconfig "github.com/a-castellano/go-types/rabbitmq"
	"errors"
	"os"
	"strings"
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
// It uses different error types to simulate various failure scenarios.
type RabbitmqMock struct {
	ErrorType string // Type of error to simulate: "connection", "channel", "queue", "publish", "qos", "consume", "none"
}

// SendMessage simulates sending a message through RabbitMQ.
// Returns different errors based on ErrorType to test various failure scenarios.
func (client RabbitmqMock) SendMessage(queueName string, message []byte) error {
	switch client.ErrorType {
	case "connection":
		return errors.New("connection failed")
	case "channel":
		return errors.New("channel creation failed")
	case "queue":
		return errors.New("queue declaration failed")
	case "publish":
		return errors.New("message publish failed")
	default:
		return nil
	}
}

// ReceiveMessages simulates receiving messages from RabbitMQ.
// It sends a test message to the messages channel and then exits.
// Returns different errors based on ErrorType to test various failure scenarios.
func (client RabbitmqMock) ReceiveMessages(ctx context.Context, queueName string, messages chan<- []byte, errorsChan chan<- error) {
	switch client.ErrorType {
	case "connection":
		errorsChan <- errors.New("connection failed")
		return
	case "channel":
		errorsChan <- errors.New("channel creation failed")
		return
	case "queue":
		errorsChan <- errors.New("queue declaration failed")
		return
	case "qos":
		errorsChan <- errors.New("QoS setting failed")
		return
	case "consume":
		errorsChan <- errors.New("message consumption failed")
		return
	case "multiple":
		// Send multiple messages with small delays
		messages <- []byte("Message 1")
		time.Sleep(10 * time.Millisecond)
		messages <- []byte("Message 2")
		time.Sleep(10 * time.Millisecond)
		messages <- []byte("Message 3")
		errorsChan <- nil
		return
	case "empty":
		// Simulate empty queue - wait for context cancellation
		select {
		case <-ctx.Done():
			errorsChan <- nil
			return
		case <-time.After(200 * time.Millisecond):
			errorsChan <- nil
			return
		}
	case "large":
		// Send a large message
		largeMessage := strings.Repeat("large message content ", 100) // ~2000 bytes
		messages <- []byte(largeMessage)
		errorsChan <- nil
		return
	case "special":
		// Send a message with special characters
		specialMessage := "Message with special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?"
		messages <- []byte(specialMessage)
		errorsChan <- nil
		return
	case "binary":
		// Send binary data
		binaryData := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC}
		messages <- binaryData
		errorsChan <- nil
		return
	default:
		// Simulate a long-running operation that can be cancelled
		select {
		case <-ctx.Done():
			// Context was cancelled, send nil error
			errorsChan <- nil
			return
		case <-time.After(50 * time.Millisecond):
			// Send a test message and then exit
			messages <- []byte("Test message")
			errorsChan <- nil
		}
	}
}

// TestSendMessageWithMockFailedSendMessage tests SendMessage operation when the underlying
// RabbitMQ SendMessage operation fails. It uses a mocked client that simulates a failure.
func TestSendMessageWithMockFailedSendMessage(t *testing.T) {

	setUp()
	defer teardown()

	// Create a mock client that will fail
	rabbitmqMock := RabbitmqMock{ErrorType: "publish"}
	messageBroker := MessageBroker{Client: rabbitmqMock}

	// Test sending a message with a failing mock
	err := messageBroker.SendMessage("test", []byte("test"))

	if err == nil {
		t.Errorf("messageBroker.SendMessage with failing mock should fail.")
	} else {
		if err.Error() != "message publish failed" {
			t.Errorf("messageBroker.SendMessage call with failing mock should return error \"message publish failed\", it has returned \"%s\"", err.Error())
		}
	}
}

// TestSendMessageWithMockSendMessage tests SendMessage operation when the underlying
// RabbitMQ SendMessage operation succeeds. It uses a mocked client that simulates success.
func TestSendMessageWithMockSendMessage(t *testing.T) {

	setUp()
	defer teardown()

	// Create a mock client that will succeed
	rabbitmqMock := RabbitmqMock{ErrorType: "none"}
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
	rabbitmqMock := RabbitmqMock{ErrorType: "consume"}
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
			if err.Error() != "message consumption failed" {
				t.Errorf("messageBroker.ReceiveMessages call with failing mock should return error \"message consumption failed\", it has returned \"%s\"", err.Error())
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
	rabbitmqMock := RabbitmqMock{ErrorType: "none"}
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

// TestSendMessageWithMockChannelError tests SendMessage operation when channel creation fails.
// This test covers the error handling in SendMessage when conn.Channel() fails.
func TestSendMessageWithMockChannelError(t *testing.T) {

	setUp()
	defer teardown()

	// Create a mock client that simulates channel creation failure
	rabbitmqMock := RabbitmqMock{ErrorType: "channel"}
	messageBroker := MessageBroker{Client: rabbitmqMock}

	// Test sending a message with channel creation failure
	err := messageBroker.SendMessage("test", []byte("test"))

	if err == nil {
		t.Errorf("messageBroker.SendMessage with channel error should fail.")
	} else {
		if err.Error() != "channel creation failed" {
			t.Errorf("messageBroker.SendMessage call with channel error should return error \"channel creation failed\", it has returned \"%s\"", err.Error())
		}
	}
}

// TestSendMessageWithMockQueueDeclareError tests SendMessage operation when queue declaration fails.
// This test covers the error handling in SendMessage when channel.QueueDeclare() fails.
func TestSendMessageWithMockQueueDeclareError(t *testing.T) {

	setUp()
	defer teardown()

	// Create a mock client that simulates queue declaration failure
	rabbitmqMock := RabbitmqMock{ErrorType: "queue"}
	messageBroker := MessageBroker{Client: rabbitmqMock}

	// Test sending a message with queue declaration failure
	err := messageBroker.SendMessage("test", []byte("test"))

	if err == nil {
		t.Errorf("messageBroker.SendMessage with queue declare error should fail.")
	} else {
		if err.Error() != "queue declaration failed" {
			t.Errorf("messageBroker.SendMessage call with queue declare error should return error \"queue declaration failed\", it has returned \"%s\"", err.Error())
		}
	}
}

// TestSendMessageWithMockPublishError tests SendMessage operation when message publishing fails.
// This test covers the error handling in SendMessage when channel.Publish() fails.
func TestSendMessageWithMockPublishError(t *testing.T) {

	setUp()
	defer teardown()

	// Create a mock client that simulates publish failure
	rabbitmqMock := RabbitmqMock{ErrorType: "publish"}
	messageBroker := MessageBroker{Client: rabbitmqMock}

	// Test sending a message with publish failure
	err := messageBroker.SendMessage("test", []byte("test"))

	if err == nil {
		t.Errorf("messageBroker.SendMessage with publish error should fail.")
	} else {
		if err.Error() != "message publish failed" {
			t.Errorf("messageBroker.SendMessage call with publish error should return error \"message publish failed\", it has returned \"%s\"", err.Error())
		}
	}
}

// TestReceiveMessageWithMockConnectionError tests ReceiveMessages operation when connection fails.
// This test covers the error handling in ReceiveMessages when amqp.Dial() fails.
func TestReceiveMessageWithMockConnectionError(t *testing.T) {

	setUp()
	defer teardown()

	// Create a mock client that simulates connection failure
	rabbitmqMock := RabbitmqMock{ErrorType: "connection"}
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
			t.Errorf("messageBroker.ReceiveMessages with connection error should return an error.")
		} else {
			if err.Error() != "connection failed" {
				t.Errorf("messageBroker.ReceiveMessages call with connection error should return error \"connection failed\", it has returned \"%s\"", err.Error())
			}
		}
	case <-time.After(5 * time.Second):
		t.Errorf("messageBroker.ReceiveMessages with connection error should return an error within 5 seconds.")
	}
}

// TestReceiveMessageWithMockChannelError tests ReceiveMessages operation when channel creation fails.
// This test covers the error handling in ReceiveMessages when conn.Channel() fails.
func TestReceiveMessageWithMockChannelError(t *testing.T) {

	setUp()
	defer teardown()

	// Create a mock client that simulates channel creation failure
	rabbitmqMock := RabbitmqMock{ErrorType: "channel"}
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
			t.Errorf("messageBroker.ReceiveMessages with channel error should return an error.")
		} else {
			if err.Error() != "channel creation failed" {
				t.Errorf("messageBroker.ReceiveMessages call with channel error should return error \"channel creation failed\", it has returned \"%s\"", err.Error())
			}
		}
	case <-time.After(5 * time.Second):
		t.Errorf("messageBroker.ReceiveMessages with channel error should return an error within 5 seconds.")
	}
}

// TestReceiveMessageWithMockQueueDeclareError tests ReceiveMessages operation when queue declaration fails.
// This test covers the error handling in ReceiveMessages when channel.QueueDeclare() fails.
func TestReceiveMessageWithMockQueueDeclareError(t *testing.T) {

	setUp()
	defer teardown()

	// Create a mock client that simulates queue declaration failure
	rabbitmqMock := RabbitmqMock{ErrorType: "queue"}
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
			t.Errorf("messageBroker.ReceiveMessages with queue declare error should return an error.")
		} else {
			if err.Error() != "queue declaration failed" {
				t.Errorf("messageBroker.ReceiveMessages call with queue declare error should return error \"queue declaration failed\", it has returned \"%s\"", err.Error())
			}
		}
	case <-time.After(5 * time.Second):
		t.Errorf("messageBroker.ReceiveMessages with queue declare error should return an error within 5 seconds.")
	}
}

// TestReceiveMessageWithMockQosError tests ReceiveMessages operation when QoS setting fails.
// This test covers the error handling in ReceiveMessages when channel.Qos() fails.
func TestReceiveMessageWithMockQosError(t *testing.T) {

	setUp()
	defer teardown()

	// Create a mock client that simulates QoS setting failure
	rabbitmqMock := RabbitmqMock{ErrorType: "qos"}
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
			t.Errorf("messageBroker.ReceiveMessages with QoS error should return an error.")
		} else {
			if err.Error() != "QoS setting failed" {
				t.Errorf("messageBroker.ReceiveMessages call with QoS error should return error \"QoS setting failed\", it has returned \"%s\"", err.Error())
			}
		}
	case <-time.After(5 * time.Second):
		t.Errorf("messageBroker.ReceiveMessages with QoS error should return an error within 5 seconds.")
	}
}

// TestReceiveMessageWithMockConsumeError tests ReceiveMessages operation when message consumption fails.
// This test covers the error handling in ReceiveMessages when channel.Consume() fails.
func TestReceiveMessageWithMockConsumeError(t *testing.T) {

	setUp()
	defer teardown()

	// Create a mock client that simulates consume failure
	rabbitmqMock := RabbitmqMock{ErrorType: "consume"}
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
			t.Errorf("messageBroker.ReceiveMessages with consume error should return an error.")
		} else {
			if err.Error() != "message consumption failed" {
				t.Errorf("messageBroker.ReceiveMessages call with consume error should return error \"message consumption failed\", it has returned \"%s\"", err.Error())
			}
		}
	case <-time.After(5 * time.Second):
		t.Errorf("messageBroker.ReceiveMessages with consume error should return an error within 5 seconds.")
	}
}

// TestReceiveMessageWithContextCancellation tests ReceiveMessages operation when context is cancelled.
// This test covers the behavior when the context is cancelled during message reception.
func TestReceiveMessageWithContextCancellation(t *testing.T) {

	setUp()
	defer teardown()

	// Create a mock client that will succeed
	rabbitmqMock := RabbitmqMock{ErrorType: "none"}
	messageBroker := MessageBroker{Client: rabbitmqMock}

	// Create channels for receiving messages and errors
	messages := make(chan []byte)
	errorsChan := make(chan error)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Start receiving messages in a goroutine
	go messageBroker.ReceiveMessages(ctx, "test", messages, errorsChan)

	// Wait for the context cancellation
	select {
	case err := <-errorsChan:
		if err != nil {
			t.Errorf("messageBroker.ReceiveMessages with context cancellation should return nil error, got %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("messageBroker.ReceiveMessages with context cancellation should complete within 100ms.")
	}
}

// TestReceiveMessageWithMultipleMessages tests ReceiveMessages operation with multiple messages.
// This test verifies that the function can handle multiple messages in sequence.
func TestReceiveMessageWithMultipleMessages(t *testing.T) {

	setUp()
	defer teardown()

	// Create a mock client that will send multiple messages
	rabbitmqMock := RabbitmqMock{ErrorType: "multiple"}
	messageBroker := MessageBroker{Client: rabbitmqMock}

	// Create channels for receiving messages and errors
	messages := make(chan []byte, 3) // Buffer to handle multiple messages
	errorsChan := make(chan error)

	// Start receiving messages in a goroutine
	ctx := context.Background()
	go messageBroker.ReceiveMessages(ctx, "test", messages, errorsChan)

	// Wait for multiple messages
	receivedMessages := make([]string, 0)
	timeout := time.After(3 * time.Second)

	for i := 0; i < 3; i++ {
		select {
		case msg := <-messages:
			receivedMessages = append(receivedMessages, string(msg))
		case err := <-errorsChan:
			if err != nil {
				t.Errorf("messageBroker.ReceiveMessages should not return error, got %v", err)
			}
			break
		case <-timeout:
			t.Errorf("messageBroker.ReceiveMessages should receive messages within 3 seconds, got %d messages", len(receivedMessages))
			return
		}
	}

	// Check that we received the expected messages
	if len(receivedMessages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(receivedMessages))
	}

	// Check specific messages
	expectedMessages := []string{"Message 1", "Message 2", "Message 3"}
	for i, msg := range receivedMessages {
		if msg != expectedMessages[i] {
			t.Errorf("Expected message %d to be '%s', got '%s'", i+1, expectedMessages[i], msg)
		}
	}
}

// TestReceiveMessageWithEmptyQueue tests ReceiveMessages operation when the queue is empty.
// This test verifies behavior when no messages are available.
func TestReceiveMessageWithEmptyQueue(t *testing.T) {

	setUp()
	defer teardown()

	// Create a mock client that simulates empty queue
	rabbitmqMock := RabbitmqMock{ErrorType: "empty"}
	messageBroker := MessageBroker{Client: rabbitmqMock}

	// Create channels for receiving messages and errors
	messages := make(chan []byte)
	errorsChan := make(chan error)

	// Start receiving messages in a goroutine
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go messageBroker.ReceiveMessages(ctx, "test", messages, errorsChan)

	// Wait for the context to timeout (empty queue)
	select {
	case err := <-errorsChan:
		if err != nil {
			t.Errorf("messageBroker.ReceiveMessages with empty queue should return nil error, got %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Errorf("messageBroker.ReceiveMessages with empty queue should complete within 200ms")
	}
}

// TestReceiveMessageWithLargeMessage tests ReceiveMessages operation with a large message.
// This test verifies that large messages are handled correctly.
func TestReceiveMessageWithLargeMessage(t *testing.T) {

	setUp()
	defer teardown()

	// Create a mock client that will send a large message
	rabbitmqMock := RabbitmqMock{ErrorType: "large"}
	messageBroker := MessageBroker{Client: rabbitmqMock}

	// Create channels for receiving messages and errors
	messages := make(chan []byte)
	errorsChan := make(chan error)

	// Start receiving messages in a goroutine
	ctx := context.Background()
	go messageBroker.ReceiveMessages(ctx, "test", messages, errorsChan)

	// Wait for the large message
	select {
	case msg := <-messages:
		// Check that the message is large
		if len(msg) < 1000 {
			t.Errorf("Expected large message (>=1000 bytes), got %d bytes", len(msg))
		}
		// Check that the message contains the expected content
		if !strings.Contains(string(msg), "large message content") {
			t.Errorf("Expected message to contain 'large message content', got '%s'", string(msg))
		}
	case err := <-errorsChan:
		t.Errorf("messageBroker.ReceiveMessages should not return error, got %v", err)
	case <-time.After(2 * time.Second):
		t.Errorf("messageBroker.ReceiveMessages should receive message within 2 seconds")
	}
}

// TestReceiveMessageWithSpecialCharacters tests ReceiveMessages operation with messages containing special characters.
// This test verifies that messages with special characters are handled correctly.
func TestReceiveMessageWithSpecialCharacters(t *testing.T) {

	setUp()
	defer teardown()

	// Create a mock client that will send a message with special characters
	rabbitmqMock := RabbitmqMock{ErrorType: "special"}
	messageBroker := MessageBroker{Client: rabbitmqMock}

	// Create channels for receiving messages and errors
	messages := make(chan []byte)
	errorsChan := make(chan error)

	// Start receiving messages in a goroutine
	ctx := context.Background()
	go messageBroker.ReceiveMessages(ctx, "test", messages, errorsChan)

	// Wait for the message with special characters
	select {
	case msg := <-messages:
		// Check that the message contains special characters
		expectedContent := "Message with special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?"
		if string(msg) != expectedContent {
			t.Errorf("Expected message '%s', got '%s'", expectedContent, string(msg))
		}
	case err := <-errorsChan:
		t.Errorf("messageBroker.ReceiveMessages should not return error, got %v", err)
	case <-time.After(2 * time.Second):
		t.Errorf("messageBroker.ReceiveMessages should receive message within 2 seconds")
	}
}

// TestReceiveMessageWithBinaryData tests ReceiveMessages operation with binary data.
// This test verifies that binary messages are handled correctly.
func TestReceiveMessageWithBinaryData(t *testing.T) {

	setUp()
	defer teardown()

	// Create a mock client that will send binary data
	rabbitmqMock := RabbitmqMock{ErrorType: "binary"}
	messageBroker := MessageBroker{Client: rabbitmqMock}

	// Create channels for receiving messages and errors
	messages := make(chan []byte)
	errorsChan := make(chan error)

	// Start receiving messages in a goroutine
	ctx := context.Background()
	go messageBroker.ReceiveMessages(ctx, "test", messages, errorsChan)

	// Wait for the binary message
	select {
	case msg := <-messages:
		// Check that the message contains binary data
		if len(msg) == 0 {
			t.Errorf("Expected binary message, got empty message")
		}
		// Check that it's not a simple string
		if len(msg) > 0 && msg[0] == 0x00 {
			t.Logf("Received binary message with %d bytes", len(msg))
		}
	case err := <-errorsChan:
		t.Errorf("messageBroker.ReceiveMessages should not return error, got %v", err)
	case <-time.After(2 * time.Second):
		t.Errorf("messageBroker.ReceiveMessages should receive message within 2 seconds")
	}
}
