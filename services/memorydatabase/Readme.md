# MemoryDatabase

This service manages interactions with memory databases. It exposes a small, driver-agnostic interface for key-value operations with TTL support. The driver bundled in this repository is Redis/Valkey ([infra/redis](../../infra/redis/Readme.md)); any client that satisfies the `Client` interface works.

## Overview

`MemoryDatabase` is a high-level abstraction that delegates every operation to a `Client` implementation. Your application depends on `MemoryDatabase` (and the `Client` interface), not on a concrete database, which keeps your code decoupled and easy to test with mocks. The repository ships a Redis driver (`redis.RedisClient`) that satisfies the interface.

## Features

- **High-level interface** for key-value operations
- **TTL support** for automatic expiration of stored values
- **Pluggable drivers** through the `Client` interface
- **Initialization guard** — operations fail fast if the driver was not initiated

## Client interface

A `MemoryDatabase` needs a value that implements this interface:

```go
type Client interface {
    WriteString(context.Context, string, string, int) error
    ReadString(context.Context, string) (string, bool, error)
    IsClientInitiated() bool
}
```

The Redis driver in [infra/redis](../../infra/redis/Readme.md) implements it.

## Usage

Import the service and the driver:

- Service: `github.com/a-castellano/go-services/services/memorydatabase`
- Redis driver: `github.com/a-castellano/go-services/infra/redis`

### Basic example

```go
package main

import (
    "context"
    "log"

    "github.com/a-castellano/go-services/infra/redis"
    "github.com/a-castellano/go-services/services/memorydatabase"
    redisconfig "github.com/a-castellano/go-types/redis"
)

func main() {
    // Build the Redis configuration from environment variables.
    config, err := redisconfig.NewConfig()
    if err != nil {
        log.Fatal(err)
    }

    // Build the Redis driver and initialize it (connects and pings the server).
    redisClient := redis.NewRedisClient(config)
    ctx := context.Background()
    if err := redisClient.Initiate(ctx); err != nil {
        log.Fatal(err)
    }

    // Inject the driver into the service.
    memoryDatabase := memorydatabase.NewMemoryDatabase(&redisClient)

    // Write a value with a 1 hour TTL.
    if err := memoryDatabase.WriteString(ctx, "user:123", "John Doe", 3600); err != nil {
        log.Fatal(err)
    }

    // Read it back.
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

> `NewMemoryDatabase` takes a `Client`. Pass the driver by pointer (`&redisClient`) because its methods have pointer receivers.

## API reference

### Service: `memorydatabase`

#### `NewMemoryDatabase(client Client) MemoryDatabase`

Creates a new `MemoryDatabase` backed by the provided client.

#### `WriteString(ctx context.Context, key string, value string, ttl int) error`

Writes a string value with the given TTL (seconds; `0` means no expiration). Returns an error if the client is not initiated or the write fails.

#### `ReadString(ctx context.Context, key string) (string, bool, error)`

Reads a string value by key. Returns the value, a boolean that is `true` when the key was found, and an error. A missing key is reported with `found == false` and a `nil` error.

### Driver: `redis` (`infra/redis`)

#### `NewRedisClient(redisConfig *redisconfig.Config) RedisClient`

Creates a `RedisClient`. The connection is **not** opened yet — call `Initiate` first.

#### `Initiate(ctx context.Context) error`

Opens the connection and validates it with a ping. Accepts both IP addresses and domain names for the host (domain names are resolved to an IP). Must be called before any read/write.

#### `IsClientInitiated() bool`

Returns `true` once `Initiate` has succeeded.

#### `WriteString` / `ReadString`

Same signatures as the service methods; these perform the actual Redis operations.

## Logging

The service is instrumented with the [logger](../../infra/logger/Readme.md) infra service. Both `WriteString` and `ReadString` retrieve the logger from the context with `logger.FromContext(ctx)`, so logs inherit any handler and attributes configured upstream. Each method carries an `operation` attribute (`WriteString` or `ReadString`) to ease filtering.

The log lifecycle for each operation is:

1. **Debug** — checking whether the driver is initiated.
2. **Debug** — the key (and value, for writes) just before delegating to the driver.
3. **Error** — if the driver was not initiated, before returning the error.

The actual storage-level lifecycle (ping, read, write) is traced by the underlying driver ([infra/redis](../../infra/redis/Readme.md)).

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

## Configuration

The Redis driver reads its configuration from the [`go-types`](https://git.windmaker.net/a-castellano/go-types) library via `redisconfig.NewConfig()`, which uses these environment variables:

- `REDIS_HOST` — Redis/Valkey server hostname or IP address
- `REDIS_PORT` — server port (default: `6379`)
- `REDIS_PASSWORD` — server password (optional)
- `REDIS_DATABASE` — database number (default: `0`)

```bash
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=mysecret
export REDIS_DATABASE=0
```

## Testing

Unit tests use a mocked Redis client; integration tests require a real Redis/Valkey server. Run them with `make` (see the [development guide](../../Readme.md#development) for the container setup):

```bash
# MemoryDatabase service tests
make test_memorydatabase_unit   # unit only
make test_memorydatabase        # unit + integration

# Redis driver tests
make test_redis_unit            # unit only
make test_redis                 # unit + integration
```

Integration tests use hardcoded IP addresses (Valkey at `172.17.0.2`) to stay consistent across environments. Start the server from the development Compose file and run:

```bash
podman-compose -f development/docker-compose.yml up -d valkey
REDIS_HOST=localhost REDIS_PORT=6379 make test_integration
```

## Error handling

The service and driver return descriptive errors for the common failure modes:

- **Client not initiated** — calling an operation before `Initiate` (or on a service whose driver was never initiated)
- **Connection failures** — wrong host/port, server unreachable
- **Authentication failures** — invalid password
- **Operation failures** — errors returned by the Redis server

A missing key on `ReadString` is **not** an error: it returns `("", false, nil)`.

## Dependencies

- [go-types](https://git.windmaker.net/a-castellano/go-types) — configuration types
- [go-redis](https://github.com/redis/go-redis) — Redis client library
- [redismock](https://github.com/go-redis/redismock) — Redis mocking for tests
</content>
