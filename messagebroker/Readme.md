# MessageBroker

This service manages interactions with message broker services. Currently supports RabbitMQ as the primary message broker implementation.

## Overview

The MessageBroker service provides a high-level abstraction for message broker operations. It uses a Client interface pattern, allowing you to easily swap implementations and test your code with mocks. The service currently supports RabbitMQ through the `RabbitmqClient` implementation.

## Features

- **Asynchronous messaging** with persistent delivery
- **Context-based cancellation** for long-running operations
- **Automatic queue declaration** and management
- **Quality of service configuration** for message processing
- **Comprehensive error handling** with detailed error messages
- **Mock support** for testing without a real RabbitMQ server

## Usage

MessageBroker requires a Client interface for being used. This library offers RabbitMQ as a Client implementation.

### Client Interface

```go
type Client interface {
    SendMessage(string, []byte) error
    ReceiveMessages(context.Context, string, chan<- []byte, chan<- error)
}
```

### Basic Example

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/a-castellano/go-services/messagebroker"
    rabbitmqconfig "github.com/a-castellano/go-types/rabbitmq"
)

func main() {
    // Create RabbitMQ configuration from environment variables
    config, err := rabbitmqconfig.NewConfig()
    if err != nil {
        log.Fatal(err)
    }

    // Create RabbitMQ client and MessageBroker
    rabbitmqClient := messagebroker.NewRabbitmqClient(config)
    messageBroker := messagebroker.MessageBroker{client: rabbitmqClient}

    // Send a message
    err = messageBroker.SendMessage("my-queue", []byte("Hello, World!"))
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Message sent successfully")
}
```

### Advanced Example with Message Receiving

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/a-castellano/go-services/messagebroker"
    rabbitmqconfig "github.com/a-castellano/go-types/rabbitmq"
)

func main() {
    config, err := rabbitmqconfig.NewConfig()
    if err != nil {
        log.Fatal(err)
    }

    rabbitmqClient := messagebroker.NewRabbitmqClient(config)
    messageBroker := messagebroker.MessageBroker{client: rabbitmqClient}

    // Send multiple messages
    messages := []string{
        "First message",
        "Second message",
        "Third message",
    }

    for _, msg := range messages {
        err := messageBroker.SendMessage("test-queue", []byte(msg))
        if err != nil {
            log.Printf("Failed to send message: %v", err)
            continue
        }
        log.Printf("Sent: %s", msg)
    }

    // Receive messages with timeout
    messagesChan := make(chan []byte)
    errorsChan := make(chan error)

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Start receiving messages in a goroutine
    go messageBroker.ReceiveMessages(ctx, "test-queue", messagesChan, errorsChan)

    // Process received messages
    for {
        select {
        case msg := <-messagesChan:
            log.Printf("Received: %s", string(msg))
        case err := <-errorsChan:
            if err != nil {
                log.Printf("Error receiving message: %v", err)
            }
            return // Exit when context is cancelled or error occurs
        case <-ctx.Done():
            log.Println("Timeout reached")
            return
        }
    }
}
```

## API Reference

### MessageBroker

#### `MessageBroker{client Client}`

Creates a new MessageBroker instance with the provided client.

#### `SendMessage(queueName string, message []byte) error`

Sends a message to the specified queue.

**Parameters:**

- `queueName`: The name of the queue to send the message to
- `message`: The message content as bytes

**Returns:** Error if the operation fails

#### `ReceiveMessages(ctx context.Context, queueName string, messages chan<- []byte, errors chan<- error)`

Continuously receives messages from the specified queue.

**Parameters:**

- `ctx`: Context for cancellation
- `queueName`: The name of the queue to receive messages from
- `messages`: Channel to receive message content
- `errors`: Channel to receive errors

### RabbitmqClient

#### `NewRabbitmqClient(rabbitmqConfig *rabbitmqconfig.Config) RabbitmqClient`

Creates a new RabbitmqClient instance with the provided configuration.

#### `SendMessage(queueName string, message []byte) error`

Sends a message to the specified queue in RabbitMQ.

**Features:**

- Automatic queue declaration
- Persistent message delivery
- Default exchange routing

#### `ReceiveMessages(ctx context.Context, queueName string, messages chan<- []byte, errors chan<- error)`

Continuously receives messages from the specified queue.

**Features:**

- Automatic queue declaration
- Quality of service configuration (prefetch count: 1)
- Context-based cancellation
- Automatic message acknowledgment

## Configuration

The RabbitMQ client uses configuration from the `go-types` library. Environment variables are used to configure the connection:

- `RABBITMQ_HOST`: RabbitMQ server hostname or IP address
- `RABBITMQ_PORT`: RabbitMQ server port (default: 5672)
- `RABBITMQ_USER`: RabbitMQ username
- `RABBITMQ_PASSWORD`: RabbitMQ password

### Example Configuration

```bash
export RABBITMQ_HOST=localhost
export RABBITMQ_PORT=5672
export RABBITMQ_USER=guest
export RABBITMQ_PASSWORD=guest
```

## Queue Configuration

The service automatically declares queues with the following settings:

- **Durable**: `true` - Queue survives server restarts
- **Auto-delete**: `false` - Queue is not deleted when unused
- **Exclusive**: `false` - Queue can be used by multiple connections
- **No-wait**: `false` - Wait for server confirmation

## Message Properties

Messages are sent with the following properties:

- **Delivery Mode**: `Persistent` - Messages survive server restarts
- **Content Type**: `text/plain` - Plain text content
- **Body**: User-provided message content

## Quality of Service

The consumer is configured with the following QoS settings:

- **Prefetch Count**: `1` - Process one message at a time
- **Prefetch Size**: `0` - No size limit
- **Global**: `false` - Apply to this channel only

## Testing

The service includes comprehensive unit and integration tests. Unit tests use mocked RabbitMQ clients, while integration tests require a real RabbitMQ server.

### Running Tests

```bash
# Run unit tests only
make test_messagebroker_unit

# Run integration tests (requires RabbitMQ server)
make test_messagebroker

# Run all tests
make test
```

### Integration Testing

For integration tests, you need a RabbitMQ server running. You can use the provided Docker Compose setup:

```bash
docker-compose -f development/docker-compose.yml -d rabbitmq
```

Then run the integration tests:

```bash
RABBITMQ_HOST=localhost RABBITMQ_PORT=5672 RABBITMQ_USER=guest RABBITMQ_PASSWORD=guest make test_integration
```

## Error Handling

The service provides detailed error messages for different failure scenarios:

- **Connection failures**: Network issues, invalid host/port
- **Authentication failures**: Invalid username/password
- **Queue declaration failures**: Permission issues, invalid queue settings
- **Message publishing failures**: Queue doesn't exist, server errors

### Common Error Scenarios

```go
// Connection failure
if err := amqp.Dial(connectionString); err != nil {
    return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
}

// Authentication failure
if err.Error() == "Exception (403) Reason: \"username or password not allowed\"" {
    return fmt.Errorf("authentication failed: %w", err)
}

// Queue declaration failure
if err := channel.QueueDeclare(...); err != nil {
    return fmt.Errorf("failed to declare queue: %w", err)
}
```

## Dependencies

- [go-types](https://git.windmaker.net/a-castellano/go-types) - Configuration types
- [amqp091-go](https://github.com/rabbitmq/amqp091-go) - RabbitMQ client library
