# MessageBroker

This service manages interactions with message broker services. It exposes a small, driver-agnostic interface for sending and receiving messages. The driver bundled in this repository is RabbitMQ ([infra/rabbitmq](../../infra/rabbitmq/Readme.md)); any client that satisfies the `Client` interface works.

## Overview

`MessageBroker` is a high-level abstraction that delegates every operation to a `Client` implementation. Your application depends on `MessageBroker` (and the `Client` interface), not on a concrete broker, which keeps your code decoupled and easy to test with mocks. The repository ships a RabbitMQ driver (`rabbitmq.RabbitmqClient`) that satisfies the interface.

## Features

- **Asynchronous messaging** with persistent delivery
- **Context-based cancellation** for long-running receivers
- **Automatic queue declaration** on both send and receive
- **Quality of service configuration** for message processing
- **Pluggable drivers** through the `Client` interface
- **Mock support** for testing without a real broker

## Client interface

A `MessageBroker` needs a value that implements this interface:

```go
type Client interface {
    SendMessage(context.Context, string, []byte) error
    ReceiveMessages(context.Context, string, chan<- []byte, chan<- error)
}
```

The RabbitMQ driver in [infra/rabbitmq](../../infra/rabbitmq/Readme.md) implements it.

## Usage

Import the service and the driver:

- Service: `github.com/a-castellano/go-services/services/messagebroker`
- RabbitMQ driver: `github.com/a-castellano/go-services/infra/rabbitmq`

The `MessageBroker` struct has a single exported field, `Client`, so you inject the driver directly:

```go
messageBroker := messagebroker.MessageBroker{Client: rabbitmqClient}
```

### Basic example

```go
package main

import (
    "context"
    "log"

    "github.com/a-castellano/go-services/infra/rabbitmq"
    "github.com/a-castellano/go-services/services/messagebroker"
    rabbitmqconfig "github.com/a-castellano/go-types/rabbitmq"
)

func main() {
    // Build the RabbitMQ configuration from environment variables.
    config, err := rabbitmqconfig.NewConfig()
    if err != nil {
        log.Fatal(err)
    }

    // Build the driver and inject it into the service.
    rabbitmqClient := rabbitmq.NewRabbitmqClient(config)
    messageBroker := messagebroker.MessageBroker{Client: rabbitmqClient}

    ctx := context.Background()

    // Send a message.
    if err := messageBroker.SendMessage(ctx, "my-queue", []byte("Hello, World!")); err != nil {
        log.Fatal(err)
    }
    log.Println("Message sent successfully")
}
```

### Receiving messages

`ReceiveMessages` runs until the context is canceled or an error occurs, streaming message bodies to the `messages` channel and errors to the `errors` channel. A `nil` value on the `errors` channel signals a clean stop.

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/a-castellano/go-services/infra/rabbitmq"
    "github.com/a-castellano/go-services/services/messagebroker"
    rabbitmqconfig "github.com/a-castellano/go-types/rabbitmq"
)

func main() {
    config, err := rabbitmqconfig.NewConfig()
    if err != nil {
        log.Fatal(err)
    }

    rabbitmqClient := rabbitmq.NewRabbitmqClient(config)
    messageBroker := messagebroker.MessageBroker{Client: rabbitmqClient}

    messagesChan := make(chan []byte)
    errorsChan := make(chan error)

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    go messageBroker.ReceiveMessages(ctx, "my-queue", messagesChan, errorsChan)

    for {
        select {
        case msg := <-messagesChan:
            log.Printf("Received: %s", string(msg))
        case err := <-errorsChan:
            if err != nil {
                log.Printf("Error receiving message: %v", err)
            }
            return // clean stop or error
        case <-ctx.Done():
            log.Println("Timeout reached")
            return
        }
    }
}
```

## API reference

### Service: `messagebroker`

#### `MessageBroker{Client: client}`

Builds a `MessageBroker` around an injected `Client`. There is no constructor function — set the `Client` field directly.

#### `SendMessage(ctx context.Context, queueName string, message []byte) error`

Sends `message` to `queueName`. Returns an error if the operation fails.

#### `ReceiveMessages(ctx context.Context, queueName string, messages chan<- []byte, errors chan<- error)`

Continuously receives from `queueName`, writing message bodies to `messages` and errors to `errors`. Cancel `ctx` to stop; a `nil` on `errors` marks a clean stop.

### Driver: `rabbitmq` (`infra/rabbitmq`)

#### `NewRabbitmqClient(rabbitmqConfig *rabbitmqconfig.Config) RabbitmqClient`

Creates a `RabbitmqClient` wired to a real broker connection (it injects the production dial function). The driver opens a fresh connection and channel per operation.

#### `SendMessage` / `ReceiveMessages`

Same signatures as the service methods; these perform the actual RabbitMQ work described below.

> The driver is built around small `AMQPConnection` / `AMQPChannel` interfaces plus an injectable `DialFunc`. Production code uses the real dialer; unit tests build a `RabbitmqClient` with a fake dialer to avoid touching a real server.

## Logging

The service is instrumented with the [logger](../../infra/logger/Readme.md) infra service. Both `SendMessage` and `ReceiveMessages` retrieve the logger from the context with `logger.FromContext(ctx)`, so logs inherit any handler and attributes configured upstream. Each method logs the operation start at `Debug` level and carries an `operation` attribute (`SendMessage` or `ReceiveMessages`) to ease filtering. The actual broker-level lifecycle (dial, channel, queue, publish/consume) is traced by the underlying driver ([infra/rabbitmq](../../infra/rabbitmq/Readme.md)).

To attach a logger, wire it at startup and pass the enriched context down:

```go
import (
    "github.com/a-castellano/go-services/infra/logger"
    slogconfig "github.com/a-castellano/go-types/slog"
)

config, _ := slogconfig.NewConfig()
appLogger := logger.NewLogger(config)
ctx := logger.WithLogger(context.Background(), appLogger)
```

## Behavior

### Queue declaration

Both sending and receiving declare the queue first, so it always exists:

- **Durable**: `true` — survives server restarts
- **Auto-delete**: `false` — not deleted when unused
- **Exclusive**: `false` — usable by multiple connections
- **No-wait**: `false` — wait for server confirmation

### Message properties (send)

- **Delivery mode**: `Persistent` — messages survive server restarts
- **Content type**: `text/plain`
- **Routing**: published to the default exchange with the queue name as routing key

### Quality of service (receive)

- **Prefetch count**: `1` — one message at a time
- **Prefetch size**: `0` — no size limit
- **Global**: `false` — applies to this channel only
- Messages are auto-acknowledged as they are consumed.

## Configuration

The RabbitMQ driver reads its configuration from the [`go-types`](https://git.windmaker.net/a-castellano/go-types) library via `rabbitmqconfig.NewConfig()`, which uses these environment variables:

- `RABBITMQ_HOST` — server hostname or IP address
- `RABBITMQ_PORT` — server port (default: `5672`)
- `RABBITMQ_USER` — username
- `RABBITMQ_PASSWORD` — password

```bash
export RABBITMQ_HOST=localhost
export RABBITMQ_PORT=5672
export RABBITMQ_USER=guest
export RABBITMQ_PASSWORD=guest
```

## Testing

Unit tests use a fake dialer (no broker needed); integration tests require a real RabbitMQ server. Run them with `make` (see the [development guide](../../Readme.md#development) for the container setup):

```bash
# MessageBroker service tests
make test_messagebroker_unit    # unit only

# RabbitMQ driver tests
make test_rabbitmq_unit         # unit only
make test_rabbitmq              # unit + integration
```

Integration tests use hardcoded IP addresses (RabbitMQ at `172.17.0.30`) to stay consistent across environments. Start the server from the development Compose file and run:

```bash
podman-compose -f development/docker-compose.yml up -d rabbitmq
RABBITMQ_HOST=localhost RABBITMQ_PORT=5672 RABBITMQ_USER=guest RABBITMQ_PASSWORD=guest make test_integration
```

## Error handling

The driver surfaces errors for the common failure modes:

- **Connection failures** — wrong host/port, server unreachable
- **Authentication failures** — invalid username/password
- **Queue declaration failures** — permission issues or incompatible queue settings
- **Publishing failures** — server errors while publishing

On `SendMessage` these are returned directly. On `ReceiveMessages` they are delivered through the `errors` channel.

## Dependencies

- [go-types](https://git.windmaker.net/a-castellano/go-types) — configuration types
- [amqp091-go](https://github.com/rabbitmq/amqp091-go) — RabbitMQ client library
</content>
