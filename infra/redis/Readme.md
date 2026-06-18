# Redis driver

Standalone Redis/Valkey client. It is a self-contained backend driver: it can be used directly, or injected into any service whose `Client` interface it satisfies. Today it satisfies the [MemoryDatabase](../../services/memorydatabase/Readme.md) service's `memorydatabase.Client` interface, providing key-value storage with TTL.

- **Package**: `redis`
- **Import**: `github.com/a-castellano/go-services/infra/redis`

## Overview

`RedisClient` wraps the [go-redis](https://github.com/redis/go-redis) client and adds a small, reusable surface: initialization with a connection check, an initialization guard, and string read/write with TTL. That surface matches the `MemoryDatabase` service's `Client` interface:

```go
type Client interface {
    WriteString(context.Context, string, string, int) error
    ReadString(context.Context, string) (string, bool, error)
    IsClientInitiated() bool
}
```

The connection is created lazily: `NewRedisClient` only stores the configuration, and `Initiate` opens and validates the connection. Operations called before a successful `Initiate` return an error instead of panicking.

## Usage

### With the MemoryDatabase service

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
    config, err := redisconfig.NewConfig()
    if err != nil {
        log.Fatal(err)
    }

    // Build and initialize the driver.
    redisClient := redis.NewRedisClient(config)
    ctx := context.Background()
    if err := redisClient.Initiate(ctx); err != nil {
        log.Fatal(err)
    }

    // Inject it into the service.
    memoryDatabase := memorydatabase.NewMemoryDatabase(&redisClient)

    if err := memoryDatabase.WriteString(ctx, "user:123", "John Doe", 3600); err != nil {
        log.Fatal(err)
    }

    value, found, err := memoryDatabase.ReadString(ctx, "user:123")
    if err != nil {
        log.Fatal(err)
    }
    if found {
        log.Printf("User: %s", value)
    }
}
```

### Standalone

The driver can also be used directly, without the service wrapper:

```go
redisClient := redis.NewRedisClient(config)
if err := redisClient.Initiate(ctx); err != nil {
    log.Fatal(err)
}

if err := redisClient.WriteString(ctx, "key", "value", 60); err != nil {
    log.Fatal(err)
}

value, found, err := redisClient.ReadString(ctx, "key")
```

> Pass the driver by pointer (`&redisClient`) when injecting it into the service: its methods have pointer receivers.

## API reference

#### `NewRedisClient(redisConfig *redisconfig.Config) RedisClient`

Creates a `RedisClient` from a configuration. Does **not** open the connection — call `Initiate` first.

#### `Initiate(ctx context.Context) error`

Opens the connection and validates it with a ping. The host may be an IP address or a domain name; domain names are resolved to an IP before connecting. Must be called before any read/write.

#### `IsClientInitiated() bool`

Returns `true` once `Initiate` has succeeded. The `MemoryDatabase` service calls this before every operation.

#### `WriteString(ctx context.Context, key string, value string, ttl int) error`

Stores `value` under `key` with a TTL in seconds (`0` means no expiration). Returns an error if the client is not initiated or the write fails.

#### `ReadString(ctx context.Context, key string) (string, bool, error)`

Reads the value for `key`. Returns the value, `true` when the key exists, and an error. A missing key returns `("", false, nil)` — it is not treated as an error.

## Logging

The driver is instrumented with the [logger](../logger/Readme.md) infra service. Every method retrieves the logger from the context with `logger.FromContext(ctx)`, so logs inherit any handler and attributes configured upstream. Operations are traced at `Debug` level (connection setup, host resolution, reads and writes) and failures are reported at `Error` level; each entry carries an `operation` attribute (`initiate`, `WriteString`, `ReadString`) to ease filtering.

When the configuration is logged (for example on `Initiate`), the password is **not** exposed: `redisconfig.Config` implements `slog.LogValuer` and its `LogValue` method masks the password as `*****` (and omits it entirely when empty). Passing `client.config` straight to the logger is therefore safe.

## Configuration

Configuration comes from the [`go-types`](https://git.windmaker.net/a-castellano/go-types) library via `redisconfig.NewConfig()`, which reads these environment variables:

- `REDIS_HOST` — server hostname or IP address
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

Unit tests use [redismock](https://github.com/go-redis/redismock) (no server needed); integration tests require a real Redis/Valkey server. See the [development guide](../../Readme.md#development) for the container setup.

```bash
make test_redis_unit   # unit only
make test_redis        # unit + integration
```

Integration tests use a hardcoded IP (Valkey at `172.17.0.2`). Start the server and run:

```bash
podman-compose -f development/docker-compose.yml up -d valkey
REDIS_HOST=localhost REDIS_PORT=6379 make test_integration
```

## Dependencies

- [go-types](https://git.windmaker.net/a-castellano/go-types) — configuration types
- [go-redis](https://github.com/redis/go-redis) — Redis client library
- [redismock](https://github.com/go-redis/redismock) — Redis mocking for tests
</content>
