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

// TestRedIsClientInitiatedWithoutEnvVariablesWriteNotInitiated tests that WriteString
// fails when the Redis client is not initialized, even when environment variables are set.
func TestRedIsClientInitiatedWithoutEnvVariablesWriteNotInitiated(t *testing.T) {

	setUp()
	defer teardown()

	config, err := redisconfig.NewConfig()

	if err != nil {
		t.Errorf("NewConfig method without any env variable suited shouldn't fail, error was '%s'.", err.Error())
	} else {
		redisClient := NewRedisClient(config)
		if ok := redisClient.IsClientInitiated(); ok {
			t.Errorf("RedisClient should not be initiated after being created")
		} else {
			// Test WriteString operation with uninitialized client
			ctx := context.Background()
			err := redisClient.WriteString(ctx, "anykey", "anyvalue", 0)
			if err == nil {
				t.Errorf("redisClient.WriteString call without redisClient being initiated should fail as redisClient is not initiated")
			} else {
				if err.Error() != "Redis client is not initiated, cannot perform WriteString operation" {
					t.Errorf("redisClient.WriteString call without redisClient being initiated should return error \"Redis client is not initiated, cannot perform WriteString operation\", it has returned \"%s\"", err.Error())
				}
			}
		}
	}
}

// TestRedIsClientInitiatedWithoutEnvVariablesReadNotInitiated tests that ReadString
// fails when the Redis client is not initialized, even when environment variables are set.
func TestRedIsClientInitiatedWithoutEnvVariablesReadNotInitiated(t *testing.T) {

	setUp()
	defer teardown()

	config, err := redisconfig.NewConfig()

	if err != nil {
		t.Errorf("NewConfig method without any env variable suited shouldn't fail, error was '%s'.", err.Error())
	} else {
		redisClient := NewRedisClient(config)
		if ok := redisClient.IsClientInitiated(); ok {
			t.Errorf("RedisClient should not be initiated after being created")
		} else {
			// Test WriteString operation with uninitialized client
			ctx := context.Background()
			_, _, err := redisClient.ReadString(ctx, "anykey")
			if err == nil {
				t.Errorf("redisClient.ReadString call without redisClient being initiated should fail as redisClient is not initiated")
			} else {
				if err.Error() != "Redis client is not initiated, cannot perform ReadString operation" {
					t.Errorf("redisClient.ReadString call without redisClient being initiated should return error \"Redis client is not initiated, cannot perform ReadString operation\", it has returned \"%s\"", err.Error())
				}
			}
		}
	}
}

func TestFailedWriteWithMock(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mock.Regexp().ExpectSet(`[a-z]+`, `[a-z]+`, 30*time.Second).SetErr(errors.New("Mocked write fail"))
	config := redisconfig.Config{}
	client := RedisClient{config: &config, client: db, clientInitiated: true}

	ctx := context.Background()
	writeErr := client.WriteString(ctx, "testkey", "value", 30)
	if writeErr == nil {
		t.Errorf("WriteString with mocked client that will error in write should fail too.")
	} else {
		expectedError := "Mocked write fail"
		if writeErr.Error() != expectedError {
			t.Fatalf("Expected error '%s' but got '%s'", expectedError, writeErr.Error())
		}

	}
}

func TestFailedReadWithMock(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mock.Regexp().ExpectGet(`[a-z]+`).SetErr(errors.New("Mocked read fail"))
	config := redisconfig.Config{}
	client := RedisClient{config: &config, client: db, clientInitiated: true}

	ctx := context.Background()
	_, _, readErr := client.ReadString(ctx, "testkey")
	if readErr == nil {
		t.Errorf("WriteString with mocked client that will error in read should fail too.")
	} else {
		expectedError := "Mocked read fail"
		if readErr.Error() != expectedError {
			t.Fatalf("Expected error '%s' but got '%s'", expectedError, readErr.Error())
		}

	}
}
