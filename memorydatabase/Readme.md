# MemoryDatabase

This service manages interactions with memory databases.
For the time being the only Memoery Databse supported is Redis.

## Usage

MemoryDatabase requires a Client interface for being used, this library offers Redis as Client interface.

### Example

```go
type Client interface {
	WriteString(context.Context, string, string, int) error
	ReadString(context.Context, string) (string, error)
	isClientInitiated() bool
}
```

[Redis](https://git.windmaker.net/a-castellano/go-types/-/tree/master/redis) type from [go-types](https://git.windmaker.net/a-castellano/go-types/) can be used as Client.

```go
package main

import (
	"context"
	redisconfig "github.com/a-castellano/go-types/redis"
	"os"
)

func main() {
	redisIP := os.Getenv("REDIS_IP")
	config, _ := redisconfig.NewConfig()
	redisClient := NewRedisClient(config)
	// Initiate Client
	if ok := redisClient.isClientInitiated(); ! ok{
		initiateErr := redisClient.Initiate(ctx)
		if initiateErr != nil {
		  panic(initiateErr)
		}
		// Now redisClient is initiate use it ad MemoryDatabase Client
		memoryDatabase := MemoryDatabase{client: &redisClient}

		// Write some value
		writeError := memoryDatabase.WriteString(ctx, "anykey", "anyvalue", 0)

		// Read some value
		readValue, valueFound, readExistentKeyError := memoryDatabase.ReadString(ctx, "anykey")
	}
}
```

### Initiate client

Before perform any read or write Client instace must be initiated:
```go
initiateErr := redisClient.Initiate(ctx)
```

Otherwise, memoryDatabase functions will fail and won't perform any action.

## Available Functions

The following functions are available for interacting with MemoryDatabase

### Initiate
```go
func (client *RedisClient) Initiate(ctx context.Context) error
```

This function initiate Redis and pings the server

### WriteString
```go
func (client *RedisClient) WriteString(ctx context.Context, key string, value string, ttl int) error
```

Writes string value in required key, ttl of that key must be set, use 0 value for infite ttl.

### ReadString
```go
func (memorydatabase *MemoryDatabase) ReadString(ctx context.Context, key string) (string, bool, error)
```

Reads string value in required key. If value is not found, bool returned will be false.

## Integration and coverage tests

Tests expect a Redis server without auth protection running, you can set Redis IP using env variables:
```bash
REDIS_IP="127.0.0.1" make test_integration
REDIS_IP="127.0.0.1" make coverage
```
