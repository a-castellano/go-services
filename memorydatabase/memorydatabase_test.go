//go:build integration_tests || unit_tests || memorydatabase_tests || memorydatabase_unit_tests

// Package memorydatabase_test contains unit tests for the memorydatabase package.
// Tests cover both Redis client functionality and MemoryDatabase wrapper operations.
package memorydatabase

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

// TestRedIsClientInitiatedWithoutEnvVariablesWriteNotInitiated tests that WriteString
// fails when the Redis client is not initialized, even without environment variables set.
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
			memoryDatabase := NewMemoryDatabase(&redisClient)
			err := memoryDatabase.WriteString(ctx, "anykey", "anyvalue", 0)
			if err == nil {
				t.Errorf("memoryDatabase.WriteString call without redisClient being initiated should fail as redisClient is not initiated")
			} else {
				if err.Error() != "client is not initiated, cannot perform WriteString operation" {
					t.Errorf("memoryDatabase.WriteString call without redisClient being initiated should return error \"client is not initiated, cannot perform WriteString operation\", it has returned \"%s\"", err.Error())
				}
			}
		}
	}
}

// TestWriteStringWithMockSetFail tests WriteString operation when the underlying Redis
// Set operation fails. It uses a mocked Redis client that simulates a Set failure.
func TestWriteStringWithMockSetFail(t *testing.T) {

	setUp()
	defer teardown()

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	// Configure mock to expect a Set operation and return an error
	mock.Regexp().ExpectSet("anykey", "anyvalue", 0).SetErr(errors.New("FAIL"))

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := MemoryDatabase{client: &redisClientMock}
	err := memoryDatabase.WriteString(ctx, "anykey", "anyvalue", 0)

	if err == nil {
		t.Errorf("memoryDatabase.WriteString with mocked redis error should fail.")
	} else {
		if err.Error() != "FAIL" {
			t.Errorf("memoryDatabase.WriteString call with mock should return error \"FAIL\", it has returned \"%s\"", err.Error())
		}
	}
}

// TestReadStringWithMockSetValue tests ReadString operation when the key exists
// and has a value. It uses a mocked Redis client that simulates a successful Get operation.
func TestReadStringWithMockSetValue(t *testing.T) {

	setUp()
	defer teardown()

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	// Configure mock to expect a Get operation and return a value
	mock.Regexp().ExpectGet("anykey").SetVal("anyvalue")

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := MemoryDatabase{client: &redisClientMock}

	value, found, err := memoryDatabase.ReadString(ctx, "anykey")

	if err != nil {
		t.Errorf("memoryDatabase.ReadString with mocked value should not fail.")
	} else {
		if value != "anyvalue" {
			t.Errorf("memoryDatabase.ReadString with mocked value should return \"anyvalue\", it has returned \"%s\"", value)
		}
		if found == false {
			t.Errorf("memoryDatabase.ReadString with mocked value should find value")
		}
	}
}

// TestReadStringWithMockNilValue tests ReadString operation when the key doesn't exist.
// It uses a mocked Redis client that simulates a Get operation returning goredis.Nil.
func TestReadStringWithMockNilValue(t *testing.T) {

	setUp()
	defer teardown()

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	// Configure mock to expect a Get operation and return goredis.Nil (key not found)
	mock.Regexp().ExpectGet("anykey").RedisNil()

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := MemoryDatabase{client: &redisClientMock}

	value, found, err := memoryDatabase.ReadString(ctx, "anykey")

	if err != nil {
		t.Errorf("memoryDatabase.ReadString with mocked nil value should not fail.")
	} else {
		if value != "" {
			t.Errorf("memoryDatabase.ReadString with mocked nil value should return empty string, it has returned \"%s\"", value)
		}
		if found == true {
			t.Errorf("memoryDatabase.ReadString with mocked nil value should not find value")
		}
	}
}

// TestRedIsClientInitiatedWithoutEnvVariablesReadNotInitiated tests that ReadString
// fails when the Redis client is not initialized, even without environment variables set.
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
			// Test ReadString operation with uninitialized client
			ctx := context.Background()
			memoryDatabase := NewMemoryDatabase(&redisClient)
			_, _, err := memoryDatabase.ReadString(ctx, "anykey")
			if err == nil {
				t.Errorf("memoryDatabase.ReadString call without redisClient being initiated should fail as redisClient is not initiated")
			} else {
				if err.Error() != "client is not initiated, cannot perform ReadString operation" {
					t.Errorf("memoryDatabase.ReadString call without redisClient being initiated should return error \"client is not initiated, cannot perform ReadString operation\", it has returned \"%s\"", err.Error())
				}
			}
		}
	}
}

// TestRedIsClientInitiatedWithoutEnvVariableForIPAsRedisHost tests Redis client initialization
// when the Redis host is specified as an IP address without environment variables.
func TestRedIsClientInitiatedWithoutEnvVariableForIPAsRedisHost(t *testing.T) {

	setUp()
	defer teardown()

	// Set Redis host as an IP address
	os.Setenv("REDIS_HOST", "127.0.0.1")
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
			if initiateErr == nil {
				t.Errorf("Redis required host is not valid, Initiate should fail")
			}
		}
	}
}

// TestRedisClientInitiateWithDomainName tests Redis client initialization
// when the Redis host is specified as a domain name.
func TestRedisClientInitiateWithDomainName(t *testing.T) {

	setUp()
	defer teardown()

	// Set Redis host as a domain name
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")

	config, err := redisconfig.NewConfig()

	if err != nil {
		t.Errorf("NewConfig method with valid host and port shouldn't fail, error was '%s'.", err.Error())
	} else {
		redisClient := NewRedisClient(config)
		if ok := redisClient.IsClientInitiated(); ok {
			t.Errorf("RedisClient should not be initiated after being created")
		} else {
			// Test client initialization with domain name
			ctx := context.Background()
			initiateErr := redisClient.Initiate(ctx)
			if initiateErr == nil {
				t.Errorf("Redis required host is not valid, Initiate should fail")
			}
		}
	}
}

// TestRedisClientInitiateWithResolvableDomain tests Redis client initialization
// when the Redis host is specified as a resolvable domain name.
func TestRedisClientInitiateWithResolvableDomain(t *testing.T) {

	setUp()
	defer teardown()

	// Set Redis host as a resolvable domain name
	os.Setenv("REDIS_HOST", "redis")
	os.Setenv("REDIS_PORT", "6379")

	config, err := redisconfig.NewConfig()

	if err != nil {
		t.Errorf("NewConfig method with valid host and port shouldn't fail, error was '%s'.", err.Error())
	} else {
		redisClient := NewRedisClient(config)
		if ok := redisClient.IsClientInitiated(); ok {
			t.Errorf("RedisClient should not be initiated after being created")
		} else {
			// Test client initialization with resolvable domain
			ctx := context.Background()
			initiateErr := redisClient.Initiate(ctx)
			if initiateErr == nil {
				t.Errorf("Redis required host is not valid, Initiate should fail")
			}
		}
	}
}

// TestWriteStringWithEmptyKey tests WriteString operation with an empty key.
// This should work as Redis allows empty keys.
func TestWriteStringWithEmptyKey(t *testing.T) {

	setUp()
	defer teardown()

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	// Configure mock to expect a Set operation with empty key
	mock.Regexp().ExpectSet("", "anyvalue", 0).SetVal("OK")

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := MemoryDatabase{client: &redisClientMock}
	err := memoryDatabase.WriteString(ctx, "", "anyvalue", 0)

	if err != nil {
		t.Errorf("memoryDatabase.WriteString with empty key should not fail.")
	}
}

// TestWriteStringWithEmptyValue tests WriteString operation with an empty value.
// This should work as Redis allows empty values.
func TestWriteStringWithEmptyValue(t *testing.T) {

	setUp()
	defer teardown()

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	// Configure mock to expect a Set operation with empty value
	mock.Regexp().ExpectSet("anykey", "", 0).SetVal("OK")

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := MemoryDatabase{client: &redisClientMock}
	err := memoryDatabase.WriteString(ctx, "anykey", "", 0)

	if err != nil {
		t.Errorf("memoryDatabase.WriteString with empty value should not fail.")
	}
}

// TestWriteStringWithTTL tests WriteString operation with a positive TTL value.
// This verifies that TTL is properly converted to time.Duration.
func TestWriteStringWithTTL(t *testing.T) {

	setUp()
	defer teardown()

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	// Configure mock to expect a Set operation with TTL
	mock.Regexp().ExpectSet("anykey", "anyvalue", 60*time.Second).SetVal("OK")

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := MemoryDatabase{client: &redisClientMock}
	err := memoryDatabase.WriteString(ctx, "anykey", "anyvalue", 60)

	if err != nil {
		t.Errorf("memoryDatabase.WriteString with TTL should not fail.")
	}
}

// TestWriteStringWithNegativeTTL tests WriteString operation with a negative TTL value.
// This should work as Redis handles negative TTL values.
func TestWriteStringWithNegativeTTL(t *testing.T) {

	setUp()
	defer teardown()

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	// Configure mock to expect a Set operation with negative TTL
	mock.Regexp().ExpectSet("anykey", "anyvalue", -60*time.Second).SetVal("OK")

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := MemoryDatabase{client: &redisClientMock}
	err := memoryDatabase.WriteString(ctx, "anykey", "anyvalue", -60)

	if err != nil {
		t.Errorf("memoryDatabase.WriteString with negative TTL should not fail.")
	}
}

// TestWriteStringWithNilRedisResponse tests WriteString operation when the Redis
// Set operation returns nil (which should be treated as an error).
func TestWriteStringWithNilRedisResponse(t *testing.T) {

	setUp()
	defer teardown()

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	// Configure mock to return nil for Set operation
	mock.Regexp().ExpectSet("anykey", "anyvalue", 0).SetVal("")

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := MemoryDatabase{client: &redisClientMock}
	err := memoryDatabase.WriteString(ctx, "anykey", "anyvalue", 0)

	if err == nil {
		t.Errorf("memoryDatabase.WriteString with nil redis response should fail.")
	} else {
		if err.Error() != "Something wrong happened executing WriteString" {
			t.Errorf("memoryDatabase.WriteString call with nil redis response should return error \"Something wrong happened executing WriteString\", it has returned \"%s\"", err.Error())
		}
	}
}
