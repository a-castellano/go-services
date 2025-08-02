# MemoryDatabase

This service manages interactions with memory databases. Currently supports Redis as the primary memory database implementation.

## Overview

The MemoryDatabase service provides a high-level abstraction for memory database operations. It uses a Client interface pattern, allowing you to easily swap implementations and test your code with mocks. The service currently supports Redis through the `RedisClient` implementation.

## Features

- **High-level interface** for key-value operations
- **TTL support** for automatic expiration of stored values

## Usage

MemoryDatabase requires a Client interface for being used. This library offers Redis as a Client implementation.

### Client Interface

```go
type Client interface {
    WriteString(context.Context, string, string, int) error
    ReadString(context.Context, string) (string, bool, error)
    IsClientInitiated() bool
}
```

### Basic Example

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/a-castellano/go-services/memorydatabase"
    redisconfig "github.com/a-castellano/go-types/redis"
)

func main() {
    // Create Redis configuration from environment variables
    config, err := redisconfig.NewConfig()
    if err != nil {
        log.Fatal(err)
    }

    // Create Redis client
    redisClient := memorydatabase.NewRedisClient(config)

    // Initialize the client (establishes connection and validates it)
    ctx := context.Background()
    if err := redisClient.Initiate(ctx); err != nil {
        log.Fatal(err)
    }

    // Create MemoryDatabase instance with the Redis client
    memoryDatabase := memorydatabase.NewMemoryDatabase(&redisClient)

    // Write a value with TTL (time-to-live in seconds)
    err = memoryDatabase.WriteString(ctx, "user:123", "John Doe", 3600) // 1 hour TTL
    if err != nil {
        log.Fatal(err)
    }

    // Read the value
    value, found, err := memoryDatabase.ReadString(ctx, "user:123")
    if err != nil {
        log.Fatal(err)
    }

    if found {
        log.Printf("User: %s", value)
    } else {
        log.Println("User not found")
    }
}
```

## API Reference

### MemoryDatabase

#### `NewMemoryDatabase(client Client) MemoryDatabase`

Creates a new MemoryDatabase instance with the provided client.

#### `WriteString(ctx context.Context, key string, value string, ttl int) error`

Writes a string value to the memory database with the specified TTL (time-to-live in seconds).

**Parameters:**

- `ctx`: Context for the operation
- `key`: The key to store the value under
- `value`: The string value to store
- `ttl`: Time-to-live in seconds (0 for no expiration)

**Returns:** Error if the operation fails

#### `ReadString(ctx context.Context, key string) (string, bool, error)`

Reads a string value from the memory database by key.

**Parameters:**

- `ctx`: Context for the operation
- `key`: The key to read

**Returns:**

- `string`: The value (empty if not found)
- `bool`: True if the key was found
- `error`: Error if the operation fails

### RedisClient

#### `NewRedisClient(redisConfig *redisconfig.Config) RedisClient`

Creates a new RedisClient instance with the provided configuration.

#### `Initiate(ctx context.Context) error`

Establishes a connection to the Redis server and validates it with a ping.

**Parameters:**

- `ctx`: Context for the operation

**Returns:** Error if connection fails

#### `IsClientInitiated() bool`

Returns true if the Redis client has been successfully initialized.

#### `WriteString(ctx context.Context, key string, value string, ttl int) error`

Stores a string value in Redis with the specified TTL.

#### `ReadString(ctx context.Context, key string) (string, bool, error)`

Retrieves a string value from Redis by key.

## Configuration

The Redis client uses configuration from the `go-types` library. Environment variables are used to configure the connection:

- `REDIS_HOST`: Redis server hostname or IP address
- `REDIS_PORT`: Redis server port (default: 6379)
- `REDIS_PASSWORD`: Redis server password (optional)
- `REDIS_DATABASE`: Redis database number (default: 0)

### Example Configuration

```bash
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=mysecret
export REDIS_DATABASE=0
```

## Testing

The service includes comprehensive unit and integration tests. Unit tests use mocked Redis clients, while integration tests require a real Redis server.

### Running Tests

```bash
# Run unit tests only
make test_memorydatabase_unit

# Run integration tests (requires Redis server)
make test_memorydatabase

# Run all tests
make test
```

### Integration Testing

For integration tests, you need a Redis server running. You can use the provided Docker Compose setup:

```bash
docker-compose -f development/docker-compose.yml up -d valkey
```

Then run the integration tests:

```bash
REDIS_HOST=localhost REDIS_PORT=6379 make test_integration
```

## Error Handling

The service provides detailed error messages for different failure scenarios:

- **Connection failures**: Network issues, invalid host/port
- **Authentication failures**: Invalid password
- **Operation failures**: Redis server errors
- **Initialization failures**: Client not properly initialized

### Common Error Scenarios

```go
// Client not initialized
if !redisClient.IsClientInitiated() {
    return errors.New("client is not initiated, cannot perform operation")
}

// Connection failure
if err := redisClient.Initiate(ctx); err != nil {
    return fmt.Errorf("failed to connect to Redis: %w", err)
}

// Key not found (not an error, but handled specially)
if err == goredis.Nil {
    return "", false, nil // Key doesn't exist
}
```

## Dependencies

- [go-types](https://git.windmaker.net/a-castellano/go-types) - Configuration types
- [go-redis](https://github.com/redis/go-redis) - Redis client library
- [redismock](https://github.com/go-redis/redismock) - Redis mocking for tests
