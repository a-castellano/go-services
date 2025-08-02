//go:build integration_tests || memorydatabase_tests

// Package memorydatabase_integration_test contains integration tests for the memorydatabase package.
// These tests require a real Redis server to be running and test actual Redis operations.
package memorydatabase

import (
	"context"
	"os"
	"testing"

	redisconfig "github.com/a-castellano/go-types/redis"
)

// TestRedisClientInvalidPort tests Redis client initialization with an invalid port.
// This test verifies that the client properly handles connection failures when the port is not accessible.
func TestRedisClientInvalidPort(t *testing.T) {

	setUp()
	defer teardown()

	// Set environment variables for an invalid Redis configuration
	os.Setenv("REDIS_HOST", "redis")
	os.Setenv("REDIS_PORT", "1234")

	config, err := redisconfig.NewConfig()

	if err != nil {
		t.Errorf("NewConfig method with valid host and port shouldn't fail, error was '%s'.", err.Error())
	} else {
		redisClient := NewRedisClient(config)
		if ok := redisClient.IsClientInitiated(); ok {
			t.Errorf("RedisClient should not be initiated after being created")
		} else {
			// Test client initialization with invalid port
			ctx := context.Background()
			initiateErr := redisClient.Initiate(ctx)
			if initiateErr == nil {
				t.Errorf("Redis required port is not valid, Initiate should fail")
			}
		}
	}
}

// TestRedisClientInvalidHost tests Redis client initialization with an invalid host.
// This test verifies that the client properly handles connection failures when the host is not accessible.
func TestRedisClientInvalidHost(t *testing.T) {

	setUp()
	defer teardown()

	// Set environment variable for an invalid Redis host
	os.Setenv("REDIS_HOST", "invalidhost")

	config, err := redisconfig.NewConfig()

	if err != nil {
		t.Errorf("NewConfig method with valid host and port shouldn't fail, error was '%s'.", err.Error())
	} else {
		redisClient := NewRedisClient(config)
		if ok := redisClient.IsClientInitiated(); ok {
			t.Errorf("RedisClient should not be initiated after being created")
		} else {
			// Test client initialization with invalid host
			ctx := context.Background()
			initiateErr := redisClient.Initiate(ctx)
			if initiateErr == nil {
				t.Errorf("Redis required host is not valid, Initiate should fail")
			}
		}
	}
}

// TestRedisClientInitiate tests Redis client initialization with a valid configuration.
// This test verifies that the client can successfully connect to a running Redis server.
func TestRedisClientInitiate(t *testing.T) {

	setUp()
	defer teardown()

	// Set environment variable for a valid Redis host
	os.Setenv("REDIS_HOST", "valkey")

	config, err := redisconfig.NewConfig()

	if err != nil {
		t.Errorf("NewConfig method with valid host env variable shouldn't fail, error was '%s'.", err.Error())
	} else {
		redisClient := NewRedisClient(config)
		if ok := redisClient.IsClientInitiated(); ok {
			t.Errorf("RedisClient should not be initiated after being created")
		} else {
			// Test successful client initialization
			ctx := context.Background()
			initiateErr := redisClient.Initiate(ctx)
			if initiateErr != nil {
				t.Errorf("Initiate should not fail, Error was %s", initiateErr.Error())
			} else {
				if redisClient.IsClientInitiated() != true {
					t.Error("After successful init, redisClient.IsClientInitiated() should be true.")
				}
			}
		}
	}
}

// TestRedisClientInitiateWithIP tests Redis client initialization when the host is specified as an IP address.
// This test verifies that the client can handle IP addresses correctly without DNS resolution.
func TestRedisClientInitiateWithIP(t *testing.T) {

	setUp()
	defer teardown()

	// Set environment variables for Redis with IP address
	os.Setenv("REDIS_HOST", "172.20.0.20")
	os.Setenv("REDIS_PORT", "6379")

	config, err := redisconfig.NewConfig()

	if err != nil {
		t.Errorf("NewConfig method with valid host and port shouldn't fail, error was '%s'.", err.Error())
	} else {
		redisClient := NewRedisClient(config)
		if ok := redisClient.IsClientInitiated(); ok {
			t.Errorf("RedisClient should not be initiated after being created")
		} else {
			// Test client initialization with IP address
			ctx := context.Background()
			initiateErr := redisClient.Initiate(ctx)
			if initiateErr != nil {
				t.Errorf("Initiate should not fail, Error was %s", initiateErr.Error())
			} else {
				if redisClient.IsClientInitiated() != true {
					t.Error("After successful init, redisClient.IsClientInitiated() should be true.")
				}
			}
		}
	}
}
