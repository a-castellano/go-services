# RabbitMQ driver

Standalone RabbitMQ client. It is a self-contained backend driver: it can be used directly, or injected into any service whose `Client` interface it satisfies. Today it satisfies the [MessageBroker](../../services/messagebroker/Readme.md) service's `messagebroker.Client` interface, sending and receiving messages.

- **Package**: `rabbitmq`
- **Import**: `github.com/a-castellano/go-services/infra/rabbitmq`

## Overview

`RabbitmqClient` wraps the [amqp091-go](https://github.com/rabbitmq/amqp091-go) client and exposes a reusable send/receive surface. That surface matches the `MessageBroker` service's `Client` interface:

```go
type Client interface {
    SendMessage(string, []byte) error
    ReceiveMessages(context.Context, string, chan<- []byte, chan<- error)
}
```

Each operation opens a fresh connection and channel, declares the target queue, and then publishes or consumes. There is no long-lived connection to manage.

### Testable design

`amqp091-go` exposes concrete structs (`*amqp.Connection`, `*amqp.Channel`) that cannot be mocked directly. To keep the driver unit-testable without a real broker, it defines its own small interfaces and an injectable dial function:

- `AMQPConnection` / `AMQPChannel` — the subset of connection/channel methods the driver uses. Thin `realConnection` / `realChannel` wrappers adapt the concrete types to these interfaces in production.
- `DialFunc` — `func(url string) (AMQPConnection, error)`, the single injection point. `NewRabbitmqClient` wires the production dialer (`realDial`); unit tests build a `RabbitmqClient` with a fake dialer that returns fake connections and channels.

## Usage

### With the MessageBroker service

```go
package main

import (
    "log"

    "github.com/a-castellano/go-services/infra/rabbitmq"
    "github.com/a-castellano/go-services/services/messagebroker"
    rabbitmqconfig "github.com/a-castellano/go-types/rabbitmq"
)

func main() {
    config, err := rabbitmqconfig.NewConfig()
    if err != nil {
        log.Fatal(err)
    }

    // Build the driver and inject it into the service.
    rabbitmqClient := rabbitmq.NewRabbitmqClient(config)
    messageBroker := messagebroker.MessageBroker{Client: rabbitmqClient}

    if err := messageBroker.SendMessage("my-queue", []byte("Hello, World!")); err != nil {
        log.Fatal(err)
    }
    log.Println("Message sent successfully")
}
```

See the [MessageBroker README](../../services/messagebroker/Readme.md) for a full receive example.

### Standalone

The driver can also be used directly:

```go
rabbitmqClient := rabbitmq.NewRabbitmqClient(config)

if err := rabbitmqClient.SendMessage("my-queue", []byte("Hello")); err != nil {
    log.Fatal(err)
}
```

## API reference

#### `NewRabbitmqClient(rabbitmqConfig *rabbitmqconfig.Config) RabbitmqClient`

Creates a `RabbitmqClient` wired to a real broker connection (it injects `realDial`).

#### `SendMessage(queueName string, message []byte) error`

Opens a connection and channel, declares `queueName`, and publishes `message`. Returns an error on any failure.

#### `ReceiveMessages(ctx context.Context, queueName string, messages chan<- []byte, errors chan<- error)`

Opens a connection and channel, declares `queueName`, sets QoS, and consumes. Message bodies are written to `messages`, errors to `errors`. Runs until `ctx` is canceled or an error occurs; a `nil` on `errors` marks a clean stop.

## Behavior

### Queue declaration

Both `SendMessage` and `ReceiveMessages` declare the queue before using it:

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

Configuration comes from the [`go-types`](https://git.windmaker.net/a-castellano/go-types) library via `rabbitmqconfig.NewConfig()`, which reads these environment variables:

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

Unit tests use a fake dialer (no broker needed); integration tests require a real RabbitMQ server. See the [development guide](../../Readme.md#development) for the container setup.

```bash
make test_rabbitmq_unit   # unit only
make test_rabbitmq        # unit + integration
```

Integration tests use a hardcoded IP (RabbitMQ at `172.17.0.30`). Start the server and run:

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
