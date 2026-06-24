// Package redis provides a driver for Redis/Valkey server.
package redis

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	logger "github.com/a-castellano/go-services/infra/logger"
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

	log := logger.FromContext(ctx).With("operation", "Initiate")
	log.DebugContext(ctx, "initiating Redis connection", "redisConfig", client.config)
	var actualHost string
	// Check if config.Host is a valid IP address
	if ip := net.ParseIP(client.config.Host); ip != nil {
		log.DebugContext(ctx, "Redis host is a valid IP address, using it directly", "ip", ip)
		actualHost = client.config.Host
	} else {
		// client.config.Host is a domain name, resolve it to an IP address
		log.DebugContext(ctx, "Redis host is a domain name, resolving it to an IP address", "clientHost", client.config.Host)
		ips, lookupErr := net.LookupIP(client.config.Host)
		if lookupErr != nil {
			log.ErrorContext(ctx, "cannot lookup Redis host domain name", "clientHost", client.config.Host, "errorMessage", lookupErr.Error())
			return lookupErr
		}
		actualHost = fmt.Sprintf("%s", ips[0])
		log.DebugContext(ctx, "Redis host IP resolved", "clientHost", client.config.Host, "actualIP", actualHost)
	}

	// Construct the Redis server address
	redisAddr := fmt.Sprintf("%s:%d", actualHost, client.config.Port)

	log.DebugContext(ctx, "creating Redis client")
	// Create and configure the Redis client
	client.client = goredis.NewClient(&goredis.Options{
		Addr:     redisAddr,
		Password: client.config.Password,
		DB:       client.config.Database,
	})

	log.DebugContext(ctx, "validating Redis client connection with ping")
	// Test the connection with a ping
	_, pingErr := client.client.Ping(ctx).Result()
	if pingErr != nil {
		log.ErrorContext(ctx, "Redis ping failed, cannot validate connection", "errorMessage", pingErr.Error())
		return pingErr
	}

	// Mark the client as successfully initialized so future operations are allowed
	log.InfoContext(ctx, "Redis client initiated")
	client.clientInitiated = true
	return nil
}

// WriteString stores a string value in Redis with the specified key and TTL.
// The TTL is specified in seconds. Use 0 for no expiration (infinite TTL).
// Returns an error if the operation fails or if the client is not initialized.
func (client *RedisClient) WriteString(ctx context.Context, key string, value string, ttl int) error {

	log := logger.FromContext(ctx).With("operation", "WriteString")
	log.DebugContext(ctx, "checking if Redis client is initiated")
	if client.clientInitiated == false {
		log.ErrorContext(ctx, "Redis client is not initiated, cannot perform WriteString operation")
		return errors.New("redis client is not initiated, cannot perform WriteString operation")
	}
	log.DebugContext(ctx, "writing value to Redis", "key", key, "value", value)
	status := client.client.Set(ctx, key, value, time.Duration(ttl)*time.Second)
	return status.Err()
}

// ReadString retrieves a string value from Redis by key.
// Returns the value, a boolean indicating if the key was found, and any error.
// If the key doesn't exist, the boolean will be false and the string will be empty.
func (client *RedisClient) ReadString(ctx context.Context, key string) (string, bool, error) {
	var found bool = true
	var emptyValue string = ""
	log := logger.FromContext(ctx).With("operation", "ReadString")
	log.DebugContext(ctx, "checking if Redis client is initiated")

	if client.clientInitiated == false {
		log.ErrorContext(ctx, "Redis client is not initiated, cannot perform ReadString operation")
		return emptyValue, found, errors.New("redis client is not initiated, cannot perform ReadString operation")
	}

	log.DebugContext(ctx, "reading from Redis", "key", key)
	readValue, err := client.client.Get(ctx, key).Result()
	if err != nil {
		found = false
		if err == goredis.Nil {
			// Key doesn't exist in Redis
			log.DebugContext(ctx, "Redis key is not set", "key", key)
			return emptyValue, found, nil
		} else {
			// Other error occurred
			log.ErrorContext(ctx, "cannot perform read operation from Redis", "error", err.Error())
			return emptyValue, found, err
		}
	}
	log.DebugContext(ctx, "Redis key found", "key", key, "value", readValue)
	return readValue, found, nil
}
