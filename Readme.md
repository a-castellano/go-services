# go-services

[![pipeline status](https://git.windmaker.net/a-castellano/go-services/badges/master/pipeline.svg)](https://git.windmaker.net/a-castellano/go-services/pipelines)[![coverage report](https://git.windmaker.net/a-castellano/go-services/badges/master/coverage.svg)](https://a-castellano.gitpages.windmaker.net/go-services/coverage.html)[![Quality Gate Status](https://sonarqube.windmaker.net/api/project_badges/measure?project=a-castellano_go-services_7930712b-1aab-4ea2-a917-853d91ec9cc6&metric=alert_status&token=sqb_a42785fa06f27139e2134dd8221c060aa2324877)](https://sonarqube.windmaker.net/dashboard?id=a-castellano_go-services_7930712b-1aab-4ea2-a917-853d91ec9cc6)

This repository stores reusable services used by many of my projects. The aim is to save time and reduce code duplication by unifying common functionality in a single source.

## Overview

The go-services library provides high-level abstractions for common infrastructure services like memory databases and message brokers. It follows the dependency injection pattern, allowing you to easily swap implementations and test your code with mocks.

## Available Services

### MemoryDatabase

A service for managing interactions with memory databases. Currently supports Redis as the primary implementation.

**Location:** [MemoryDatabase](/memorydatabase)

### MessageBroker

A service for managing interactions with message broker services. Currently supports RabbitMQ as the primary implementation.

**Location:** [MessageBroker](/messagebroker)

## Installation

```bash
go get github.com/a-castellano/go-services
```

## Quick Start

### Using MemoryDatabase with Redis

```go
package main

import (
    "context"
    "log"

    "github.com/a-castellano/go-services/memorydatabase"
    redisconfig "github.com/a-castellano/go-types/redis"
)

func main() {
    // Create Redis configuration
    config, err := redisconfig.NewConfig()
    if err != nil {
        log.Fatal(err)
    }

    // Create and initialize Redis client
    redisClient := memorydatabase.NewRedisClient(config)
    ctx := context.Background()

    if err := redisClient.Initiate(ctx); err != nil {
        log.Fatal(err)
    }

    // Create MemoryDatabase instance
    memoryDB := memorydatabase.NewMemoryDatabase(&redisClient)

    // Write a value with TTL
    err = memoryDB.WriteString(ctx, "my-key", "my-value", 3600) // 1 hour TTL
    if err != nil {
        log.Fatal(err)
    }

    // Read the value
    value, found, err := memoryDB.ReadString(ctx, "my-key")
    if err != nil {
        log.Fatal(err)
    }

    if found {
        log.Printf("Value: %s", value)
    }
}
```

### Using MessageBroker with RabbitMQ

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
    // Create RabbitMQ configuration
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

    // Receive messages
    messages := make(chan []byte)
    errors := make(chan error)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    go messageBroker.ReceiveMessages(ctx, "my-queue", messages, errors)

    select {
    case msg := <-messages:
        log.Printf("Received: %s", string(msg))
    case err := <-errors:
        log.Printf("Error: %v", err)
    case <-ctx.Done():
        log.Println("Timeout")
    }
}
```

## Development

### Setting up the Development Environment

1. Clone the repository:

```bash
git clone https://git.windmaker.net/a-castellano/go-services.git
cd go-services
```

2. Start the development environment:

```bash
docker-compose -f development/docker-compose up -d
```

This will start:

- A Go development container
- A Valkey (Redis-compatible) server
- A RabbitMQ server

### Running Tests

The project includes comprehensive unit and integration tests:

```bash
# Run all unit tests
make test

# Run all integration tests (requires running services)
make test_integration

# Run tests for specific services
make test_memorydatabase
make test_messagebroker

# Run tests with race detection
make race

# Generate coverage report
make coverage
```

### Available Make Targets

- `make help` - Show all available targets
- `make lint` - Run code linting
- `make test` - Run unit tests
- `make test_integration` - Run integration tests
- `make coverage` - Generate coverage report
- `make coverhtml` - Generate HTML coverage report
- `make race` - Run tests with race detection
- `make msan` - Run tests with memory sanitizer

### Environment Variables

#### For MemoryDatabase (Redis)

- `REDIS_HOST` - Redis server hostname or IP
- `REDIS_PORT` - Redis server port (default: 6379)
- `REDIS_PASSWORD` - Redis server password (optional)
- `REDIS_DATABASE` - Redis database number (default: 0)

#### For MessageBroker (RabbitMQ)

- `RABBITMQ_HOST` - RabbitMQ server hostname or IP
- `RABBITMQ_PORT` - RabbitMQ server port (default: 5672)
- `RABBITMQ_USER` - RabbitMQ username
- `RABBITMQ_PASSWORD` - RabbitMQ password

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Dependencies

- [go-types](https://git.windmaker.net/a-castellano/go-types) - Configuration types
- [go-redis](https://github.com/redis/go-redis) - Redis client
- [amqp091-go](https://github.com/rabbitmq/amqp091-go) - RabbitMQ client
- [redismock](https://github.com/go-redis/redismock) - Redis mocking for tests
