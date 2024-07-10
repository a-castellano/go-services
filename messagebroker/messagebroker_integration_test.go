//go:build integration_tests || messagebroker_tests

package messagebroker

import (
	"context"
	rabbitmqconfig "github.com/a-castellano/go-types/rabbitmq"
	"os"
	"testing"
	"time"
)

func TestRabbitmqFailedConnection(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("RABBITMQ_HOST", "rabbitmq")
	os.Setenv("RABBITMQ_PORT", "1123")
	os.Setenv("RABBITMQ_USER", "user")
	os.Setenv("RABBITMQ_PASSWORD", "password")

	rabbitmqConfig, _ := rabbitmqconfig.NewConfig()
	queueName := "test"
	testString := []byte("This is a Test")

	rabbitmqClient := NewRabbimqClient(rabbitmqConfig)

	messageBroker := MessageBroker{Client: rabbitmqClient}

	dial_error := messageBroker.SendMessage(queueName, testString)

	if dial_error == nil {
		t.Errorf("TestRabbitmqFailedConnection should fail.")
	}
}

func TestRabbitmqInvalidCredentials(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("RABBITMQ_HOST", "rabbitmq")
	os.Setenv("RABBITMQ_PORT", "5672")
	os.Setenv("RABBITMQ_USER", "user")
	os.Setenv("RABBITMQ_PASSWORD", "password")

	rabbitmqConfig, _ := rabbitmqconfig.NewConfig()
	queueName := "test"
	testString := []byte("This is a Test")

	rabbitmqClient := NewRabbimqClient(rabbitmqConfig)

	messageBroker := MessageBroker{Client: rabbitmqClient}

	dial_error := messageBroker.SendMessage(queueName, testString)

	if dial_error == nil {
		t.Errorf("TestRabbitmqFailedConnection should fail.")
	} else {
		if dial_error.Error() != "Exception (403) Reason: \"username or password not allowed\"" {
			t.Errorf("TestRabbitmqFailedConnection error should be 'Exception (403) Reason: \"username or password not allowed\"' but was '%s'.", dial_error.Error())
		}
	}
}

func TestRabbitmqSendMessage(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("RABBITMQ_HOST", "rabbitmq")
	os.Setenv("RABBITMQ_PORT", "5672")
	os.Setenv("RABBITMQ_USER", "guest")
	os.Setenv("RABBITMQ_PASSWORD", "guest")

	rabbitmqConfig, _ := rabbitmqconfig.NewConfig()
	queueName := "test"
	testString := []byte("This is a Test")

	rabbitmqClient := NewRabbimqClient(rabbitmqConfig)

	messageBroker := MessageBroker{Client: rabbitmqClient}

	sendError := messageBroker.SendMessage(queueName, testString)

	if sendError != nil {
		t.Errorf("TestRabbitmqSendMessage shouldn't fail. Error was '%s'.", sendError.Error())
	}
}

func TestRabbitmqReceiveMessageFromInvalidConfig(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("RABBITMQ_HOST", "rabbitmq")
	os.Setenv("RABBITMQ_PORT", "1122")
	os.Setenv("RABBITMQ_USER", "guest")
	os.Setenv("RABBITMQ_PASSWORD", "badPassword")

	rabbitmqConfig, _ := rabbitmqconfig.NewConfig()
	queueName := "test"
	rabbitmqClient := NewRabbimqClient(rabbitmqConfig)
	messageBroker := MessageBroker{Client: rabbitmqClient}

	messagesReceived := make(chan []byte)

	ctx, _ := context.WithCancel(context.Background())

	receiveErrors := make(chan error)

	go messageBroker.ReceiveMessages(ctx, queueName, messagesReceived, receiveErrors)

	time.Sleep(3 * time.Second)

	receiveError := <-receiveErrors

	if receiveError == nil {
		t.Errorf("TestRabbitmqReceiveMessageFromInvalidConfig should fail.")
	}
}

func TestRabbitmqReceiveMessageFromInvalidCredentials(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("RABBITMQ_HOST", "rabbitmq")
	os.Setenv("RABBITMQ_PORT", "5672")
	os.Setenv("RABBITMQ_USER", "guest")
	os.Setenv("RABBITMQ_PASSWORD", "badPassword")

	rabbitmqConfig, _ := rabbitmqconfig.NewConfig()
	queueName := "test"
	rabbitmqClient := NewRabbimqClient(rabbitmqConfig)
	messageBroker := MessageBroker{Client: rabbitmqClient}

	messagesReceived := make(chan []byte)

	ctx, _ := context.WithCancel(context.Background())

	receiveErrors := make(chan error)

	go messageBroker.ReceiveMessages(ctx, queueName, messagesReceived, receiveErrors)

	time.Sleep(3 * time.Second)

	receiveError := <-receiveErrors

	if receiveError == nil {
		t.Errorf("TestRabbitmqReceiveMessageFromInvalidConfig should fail.")
	} else {
		if receiveError.Error() != "Exception (403) Reason: \"username or password not allowed\"" {
			t.Errorf("TestRabbitmqReceiveMessageFromInvalidConfig should be 'Exception (403) Reason: \"username or password not allowed\"' but was '%s'.", receiveError.Error())
		}

	}
}

func TestRabbitmqReceiveMessage(t *testing.T) {

	defer teardown()

	os.Setenv("RABBITMQ_HOST", "rabbitmq")
	os.Setenv("RABBITMQ_PORT", "5672")
	os.Setenv("RABBITMQ_USER", "guest")
	os.Setenv("RABBITMQ_PASSWORD", "guest")

	rabbitmqConfig, _ := rabbitmqconfig.NewConfig()
	queueName := "test"
	rabbitmqClient := NewRabbimqClient(rabbitmqConfig)
	messageBroker := MessageBroker{Client: rabbitmqClient}

	messagesReceived := make(chan []byte)

	ctx, cancel := context.WithCancel(context.Background())

	receiveErrors := make(chan error)

	go messageBroker.ReceiveMessages(ctx, queueName, messagesReceived, receiveErrors)

	time.Sleep(2 * time.Second)
	cancel()
	time.Sleep(2 * time.Second)

	select {
	case receivedError := <-receiveErrors:
		t.Errorf("TestRabbitmqReceiveMessage shouldn't fail, errorwas \"%s\".", receivedError.Error())
	case messageReceived := <-messagesReceived:
		stringReceived := string(messageReceived)
		if stringReceived != "This is a Test" {
			t.Errorf("Received Message shold be \"This is a Test\", not \"%s\".", stringReceived)
		}
	default:
		t.Errorf("TestRabbitmqReceiveMessage shold return a message or an error")
	}
}
