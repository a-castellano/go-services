//go:build integration_tests || unit_tests || messagebroker_tests || messagebroker_unit_tests

// Package messagebroker_test contains unit tests for the messagebroker package.
package messagebroker

import (
	"context"
	"testing"
)

type mockBroker struct{}

func (m *mockBroker) SendMessage(string, []byte) error {
	return nil
}

func (m *mockBroker) ReceiveMessages(ctx context.Context, queueName string, messages chan<- []byte, errors chan<- error) {
	errors <- nil
}

func TestSendMessage(t *testing.T) {

	client := mockBroker{}
	messageBroker := MessageBroker{Client: &client}

	// Test sending a message with a failing mock
	messageBroker.SendMessage("test", []byte("test"))

}

// TestReceiveMessage tests ReceiveMessages
func TestReceiveMessage(t *testing.T) {

	client := mockBroker{}
	messageBroker := MessageBroker{Client: &client}

	// Create channels for receiving messages and errors
	messages := make(chan []byte)
	errorsChan := make(chan error)

	ctx := context.Background()

	// ReceiveMessages is a blocking, continuous receiver, so run it in a
	// separate goroutine and consume from the channels here.
	go messageBroker.ReceiveMessages(ctx, "test", messages, errorsChan)

	select {
	case <-messages:
	case <-errorsChan:
	}

}
