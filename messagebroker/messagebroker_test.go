//go:build integration_tests || unit_tests || messagebroker_tests || messagebroker_unit_tests

package messagebroker

import (
	"context"
	//	rabbitmqconfig "github.com/a-castellano/go-types/rabbitmq"
	"errors"
	"os"
	"testing"
	"time"
)

var currentHost string
var currentHostDefined bool

var currentPort string
var currentPortDefined bool

var currentUser string
var currentUserDefined bool

var currentPassword string
var currentPasswordDefined bool

func setUp() {

	if envHost, found := os.LookupEnv("RABBITMQ_HOST"); found {
		currentHost = envHost
		currentHostDefined = true
	} else {
		currentHostDefined = false
	}

	if envPort, found := os.LookupEnv("RABBITMQ_PORT"); found {
		currentPort = envPort
		currentPortDefined = true
	} else {
		currentPortDefined = false
	}

	if envUser, found := os.LookupEnv("RABBITMQ_USER"); found {
		currentUser = envUser
		currentUserDefined = true
	} else {
		currentUserDefined = false
	}

	if envPassword, found := os.LookupEnv("RABBITMQ_PASSWORD"); found {
		currentPassword = envPassword
		currentPasswordDefined = true
	} else {
		currentPasswordDefined = false
	}

	os.Unsetenv("RABBITMQ_HOST")
	os.Unsetenv("RABBITMQ_PORT")
	os.Unsetenv("RABBITMQ_DATABASE")
	os.Unsetenv("RABBITMQ_PASSWORD")

}

func teardown() {

	if currentHostDefined {
		os.Setenv("RABBITMQ_HOST", currentHost)
	} else {
		os.Unsetenv("RABBITMQ_HOST")
	}

	if currentPortDefined {
		os.Setenv("RABBITMQ_PORT", currentPort)
	} else {
		os.Unsetenv("RABBITMQ_PORT")
	}

	if currentUserDefined {
		os.Setenv("RABBITMQ_USER", currentUser)
	} else {
		os.Unsetenv("RABBITMQ_USER")
	}

	if currentPasswordDefined {
		os.Setenv("RABBITMQ_PASSWORD", currentPassword)
	} else {
		os.Unsetenv("RABBITMQ_PASSWORD")
	}

}

type RabbitmqMock struct {
	LaunchError bool
}

func (client RabbitmqMock) SendMessage(queueName string, message []byte) error {
	if client.LaunchError {
		return errors.New("Error")
	}
	return nil
}

func (client RabbitmqMock) ReceiveMessages(ctx context.Context, queueName string, messages chan<- []byte, errorsChan chan<- error) {
	if client.LaunchError {
		errorsChan <- errors.New("Error")
	} else {
		okMessage := []byte("This is ok")
		messages <- okMessage
		errorsChan <- nil
	}
}

func TestSendMessageWithMockFailedSendMessage(t *testing.T) {

	setUp()
	defer teardown()

	rabbitmock := RabbitmqMock{LaunchError: true}
	messageBroker := MessageBroker{client: rabbitmock}

	testMessage := []byte("This is a test")

	sendErr := messageBroker.SendMessage("anyque", testMessage)

	if sendErr == nil {
		t.Errorf("messageBroker.SendMessage method with mocked failure should fail.")
	}
}

func TestSendMessageWithMockSendMessage(t *testing.T) {

	setUp()
	defer teardown()

	rabbitmock := RabbitmqMock{LaunchError: false}
	messageBroker := MessageBroker{client: rabbitmock}

	testMessage := []byte("This is a test")

	sendErr := messageBroker.SendMessage("anyque", testMessage)

	if sendErr != nil {
		t.Errorf("messageBroker.SendMessage method with mocked non failure shouldn't fail, erro was '%s'.", sendErr.Error())
	}
}

func TestReceiveMessageWithMockFailure(t *testing.T) {

	setUp()
	defer teardown()

	rabbitmock := RabbitmqMock{LaunchError: true}
	queueName := "test"
	messagesReceived := make(chan []byte)

	ctx, _ := context.WithCancel(context.Background())

	receiveErrors := make(chan error)

	messageBroker := MessageBroker{client: rabbitmock}

	go messageBroker.ReceiveMessages(ctx, queueName, messagesReceived, receiveErrors)

	time.Sleep(1 * time.Second)

	select {
	case receivedError := <-receiveErrors:
		if receivedError.Error() != "Error" {
			t.Errorf("TestReceiveMessageWithMockFailure should fail as mock is triggering error, error should be \"Error\", not \"%s\".", receivedError.Error())
		}
	default:
		t.Errorf("TestReceiveMessageWithMockFailure should return a message or an error")
	}
}

func TestReceiveMessageWithMock(t *testing.T) {

	setUp()
	defer teardown()

	rabbitmock := RabbitmqMock{LaunchError: false}
	queueName := "test"
	messagesReceived := make(chan []byte)

	ctx, _ := context.WithCancel(context.Background())

	receiveErrors := make(chan error)

	messageBroker := MessageBroker{client: rabbitmock}

	go messageBroker.ReceiveMessages(ctx, queueName, messagesReceived, receiveErrors)

	time.Sleep(1 * time.Second)

	select {
	case receivedError := <-receiveErrors:
		t.Errorf("TestReceiveMessageWithMock shouldn't fail, errorwas \"%s\".", receivedError.Error())
	case messageReceived := <-messagesReceived:
		stringReceived := string(messageReceived)
		if stringReceived != "This is ok" {
			t.Errorf("Received Message should be \"This is ok\", not \"%s\".", stringReceived)
		}
	default:
		t.Errorf("TestReceiveMessageWithMock should return a message or an error")
	}
}
