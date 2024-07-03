# MessageBroker

This service manages interactions with memory message broker services.
For the time being the only service supported is RabbitMQ.

## Usage

MessageBroker requires a Client interface for being used, this library offers RabbitMQ as Client interface.

### Example

```go
type Client interface {
	SendMessage(string, []byte) error
	ReceiveMessages(context.Context, string, chan<- []byte, chan<- error)
}
```

[Rebbimq](https://git.windmaker.net/a-castellano/go-types/-/tree/master/rabbitmq) type from [go-types](https://git.windmaker.net/a-castellano/go-types/) can be used as Client.

```go
package main

import (
	"context"
	rabbitmqconfig "github.com/a-castellano/go-types/rabbitmq"
	"os"
)

func main() {

	rabbitmqConfig, _ := rabbitmqconfig.NewConfig()
	queueName := "test"
	testString := []byte("This is a Test")

	rabbitmqClient := NewRabbimqClient(rabbitmqConfig)

	messageBroker := MessageBroker{client: rabbitmqClient}

	send_error := messageBroker.SendMessage(queueName, testString)
}
```
## Available Functions

The following functions are avaible for interacting with MemoryDatabase

### SendMessage
```go
func (client RabbitmqClient) SendMessage(queueName string, message []byte) error
```

Sends message through required queue.

### ReceiveMessages
```go
func (client RabbitmqClient) ReceiveMessages(ctx context.Context, queueName string, messages chan<- []byte, errors chan<- error)
```

Receives messages from required queue. I can be sottped using context:
```go
rabbitmqConfig, _ := rabbitmqconfig.NewConfig()
queueName := "test"
rabbitmqClient := NewRabbimqClient(rabbitmqConfig)
messageBroker := MessageBroker{client: rabbitmqClient}

messagesReceived := make(chan []byte)

ctx, cancel := context.WithCancel(context.Background())

receiveErrors := make(chan error)

go messageBroker.ReceiveMessages(ctx, queueName, messagesReceived, receiveErrors)

time.Sleep(2 * time.Second)
cancel()

messageReceived := <-messagesReceived
```

