// Package redis provides a driver for Redis/Valkey server.
package redis

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	redisconfig "github.com/a-castellano/go-types/redis"
	goredis "github.com/redis/go-redis/v9"
)

// RedisClient implements the Client interface for Redis database operations.
// It manages the connection to a Redis server and provides methods for reading/writing data.
type RedisClient struct {
	config          *redisconfig.Config // Redis configuration (host, port, password, etc.)
	client          *goredis.Client     // Underlying Redis client from go-redis library
	clientInitiated bool                // Flag indicating if the client has been successfully initialized
}

// NewRedisClient creates a new RedisClient instance with the provided configuration.
// The client is not automatically initialized - call Initiate() to establish the connection.
func NewRedisClient(redisConfig *redisconfig.Config) RedisClient {
	redisClient := RedisClient{config: redisConfig}

	return redisClient
}

// IsClientInitiated returns true if the Redis client has been successfully initialized
// and is ready to perform operations.
func (client *RedisClient) IsClientInitiated() bool {
	return client.clientInitiated
}

// Initiate establishes a connection to the Redis server and validates it with a ping.
// This method must be called before any read/write operations can be performed.
// It handles both IP addresses and domain names for the Redis host.
func (client *RedisClient) Initiate(ctx context.Context) error {

	var actualHost string
	// Check if config.Host is a valid IP address
	if ip := net.ParseIP(client.config.Host); ip != nil {
		// Host is a valid IP address, use it directly
		actualHost = client.config.Host
	} else {
		// client.config.Host is a domain name, resolve it to an IP address
		ips, lookupErr := net.LookupIP(client.config.Host)
		if lookupErr != nil {
			return lookupErr
		}
		actualHost = fmt.Sprintf("%s", ips[0])
	}

	// Construct the Redis server address
	redisAddr := fmt.Sprintf("%s:%d", actualHost, client.config.Port)

	// Create and configure the Redis client
	client.client = goredis.NewClient(&goredis.Options{
		Addr:     redisAddr,
		Password: client.config.Password,
		DB:       client.config.Database,
	})

	// Test the connection with a ping
	_, pingErr := client.client.Ping(ctx).Result()
	if pingErr != nil {
		return pingErr
	}

	// Mark the client as successfully initialized
	client.clientInitiated = true
	return nil
}

// WriteString stores a string value in Redis with the specified key and TTL.
// The TTL is specified in seconds. Use 0 for no expiration (infinite TTL).
// Returns an error if the operation fails or if the client is not initialized.
func (client *RedisClient) WriteString(ctx context.Context, key string, value string, ttl int) error {
	status := client.client.Set(ctx, key, value, time.Duration(ttl)*time.Second)
	if status == nil {
		return errors.New("Something wrong happened executing WriteString")
	}
	return status.Err()
}

// ReadString retrieves a string value from Redis by key.
// Returns the value, a boolean indicating if the key was found, and any error.
// If the key doesn't exist, the boolean will be false and the string will be empty.
func (client *RedisClient) ReadString(ctx context.Context, key string) (string, bool, error) {
	var found bool = true
	readValue, err := client.client.Get(ctx, key).Result()
	if err != nil {
		found = false
		if err == goredis.Nil {
			// Key doesn't exist in Redis
			return "", found, nil
		} else {
			// Other error occurred
			return "", found, err
		}
	}
	return readValue, found, nil
}
