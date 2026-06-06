//go:build integration_tests || unit_tests || redis_tests || redis_unit_tests

// Package redis_test contains unit tests for the redis package.
// Tests cover Redis client functionality.
package redis

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	redisconfig "github.com/a-castellano/go-types/redis"
	redismock "github.com/go-redis/redismock/v9"
	goredis "github.com/redis/go-redis/v9"
)

// Global variables to store original environment variable values
// These are used to restore the environment after tests that modify them
var currentHost string
var currentHostDefined bool

var currentPort string
var currentPortDefined bool

var currentDatabase string
var currentDatabaseDefined bool

var currentPassword string
var currentPasswordDefined bool

// setUp saves the current environment variables and clears them for testing.
// This ensures tests start with a clean environment and can set their own values.
func setUp() {

	// Save and clear REDIS_HOST environment variable
	if envHost, found := os.LookupEnv("REDIS_HOST"); found {
		currentHost = envHost
		currentHostDefined = true
	} else {
		currentHostDefined = false
	}

	// Save and clear REDIS_PORT environment variable
	if envPort, found := os.LookupEnv("REDIS_PORT"); found {
		currentPort = envPort
		currentPortDefined = true
	} else {
		currentPortDefined = false
	}

	// Save and clear REDIS_DATABASE environment variable
	if envDatabase, found := os.LookupEnv("REDIS_DATABASE"); found {
		currentDatabase = envDatabase
		currentDatabaseDefined = true
	} else {
		currentDatabaseDefined = false
	}

	// Save and clear REDIS_PASSWORD environment variable
	if envPassword, found := os.LookupEnv("REDIS_PASSWORD"); found {
		currentPassword = envPassword
		currentPasswordDefined = true
	} else {
		currentPasswordDefined = false
	}

	// Clear all Redis environment variables for clean test state
	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("REDIS_PORT")
	os.Unsetenv("REDIS_DATABASE")
	os.Unsetenv("REDIS_PASSWORD")

}

// teardown restores the original environment variables that were saved in setUp.
// This ensures that tests don't affect the environment for other tests or processes.
func teardown() {

	// Restore REDIS_HOST environment variable
	if currentHostDefined {
		os.Setenv("REDIS_HOST", currentHost)
	} else {
		os.Unsetenv("REDIS_HOST")
	}

	// Restore REDIS_PORT environment variable
	if currentPortDefined {
		os.Setenv("REDIS_PORT", currentPort)
	} else {
		os.Unsetenv("REDIS_PORT")
	}

	// Restore REDIS_DATABASE environment variable
	if currentDatabaseDefined {
		os.Setenv("REDIS_DATABASE", currentDatabase)
	} else {
		os.Unsetenv("REDIS_DATABASE")
	}

	// Restore REDIS_PASSWORD environment variable
	if currentPasswordDefined {
		os.Setenv("REDIS_PASSWORD", currentPassword)
	} else {
		os.Unsetenv("REDIS_PASSWORD")
	}

}

// RedisClientMock implements the Client interface for testing purposes.
// It uses a mocked Redis client to simulate Redis operations without requiring a real Redis server.
type RedisClientMock struct {
	client *goredis.Client // Mocked Redis client for testing
}

// IsClientInitiated always returns true for the mock client.
// This simulates a properly initialized Redis client.
func (mock *RedisClientMock) IsClientInitiated() bool {
	return true
}

// WriteString uses the mocked Redis client to simulate writing a string value.
// This method delegates to the underlying mocked client's Set operation.
func (mock *RedisClientMock) WriteString(ctx context.Context, key string, value string, ttl int) error {
	return mock.client.Set(ctx, key, value, time.Duration(ttl)*time.Second).Err()
}

// ReadString uses the mocked Redis client to simulate reading a string value.
// This method delegates to the underlying mocked client's Get operation.
// It properly handles the case where a key doesn't exist (goredis.Nil error).
func (mock *RedisClientMock) ReadString(ctx context.Context, key string) (string, bool, error) {
	var found bool = true
	value, getError := mock.client.Get(ctx, key).Result()

	if getError != nil {
		found = false
		if getError == goredis.Nil {
			// Key doesn't exist in Redis
			return value, found, nil
		} else {
			// Other error occurred
			return value, found, getError
		}
	}
	return value, found, nil
}
