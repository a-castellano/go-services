//go:build integration_tests || memorydatabase_tests

package memorydatabase

import (
	"context"
	redisconfig "github.com/a-castellano/go-types/redis"
	"os"
	"testing"
)

func TestRedisClientInvalidPort(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("REDIS_HOST", "redis")
	os.Setenv("REDIS_PORT", "1234")

	config, err := redisconfig.NewConfig()

	if err != nil {
		t.Errorf("NewConfig method with valid host and port shouldn't fail, error was '%s'.", err.Error())
	} else {
		redisClient := NewRedisClient(config)
		if ok := redisClient.isClientInitiated(); ok {
			t.Errorf("RedisClient should not be initiated after being created")
		} else {
			// Initiate RedisClient
			ctx := context.Background()
			initiateErr := redisClient.Initiate(ctx)
			if initiateErr == nil {
				t.Errorf("Redis required port is notvalid, Initiate shold fail")
			}
		}
	}
}

func TestRedisClientInvalidHost(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("REDIS_HOST", "invalidhost")

	config, err := redisconfig.NewConfig()

	if err != nil {
		t.Errorf("NewConfig method with valid host and port shouldn't fail, error was '%s'.", err.Error())
	} else {
		redisClient := NewRedisClient(config)
		if ok := redisClient.isClientInitiated(); ok {
			t.Errorf("RedisClient should not be initiated after being created")
		} else {
			// Initiate RedisClient
			ctx := context.Background()
			initiateErr := redisClient.Initiate(ctx)
			if initiateErr == nil {
				t.Errorf("Redis required host is notvalid, Initiate shold fail")
			}
		}
	}
}

func TestRedisClientInitiate(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("REDIS_HOST", "redis")

	config, err := redisconfig.NewConfig()

	if err != nil {
		t.Errorf("NewConfig method with valid host env variable shouldn't fail, error was '%s'.", err.Error())
	} else {
		redisClient := NewRedisClient(config)
		if ok := redisClient.isClientInitiated(); ok {
			t.Errorf("RedisClient should not be initiated after being created")
		} else {
			// Initiate RedisClient
			ctx := context.Background()
			initiateErr := redisClient.Initiate(ctx)
			if initiateErr != nil {
				t.Errorf("Initiate shold notfail, Error was %s", initiateErr.Error())
			} else {
				if redisClient.isClientInitiated() != true {
					t.Error("After successful init, redisClient.isClientInitiated() should be true.")
				}
			}
		}
	}
}

func TestRedisClientInitiateWithIP(t *testing.T) {

	setUp()
	defer teardown()

	redisIP := os.Getenv("REDIS_IP")
	os.Setenv("REDIS_HOST", redisIP)

	config, err := redisconfig.NewConfig()

	if err != nil {
		t.Errorf("NewConfig method with valid host env variable shouldn't fail, error was '%s'.", err.Error())
	} else {
		redisClient := NewRedisClient(config)
		if ok := redisClient.isClientInitiated(); ok {
			t.Errorf("RedisClient should not be initiated after being created")
		} else {
			// Initiate RedisClient
			ctx := context.Background()
			initiateErr := redisClient.Initiate(ctx)
			if initiateErr != nil {
				t.Errorf("Initiate shold notfail, Error was %s", initiateErr.Error())
			} else {
				if redisClient.isClientInitiated() != true {
					t.Error("After successful init, redisClient.isClientInitiated() should be true.")
				}
			}
			// Create MemoryDatabase instance
			memoryDatabase := MemoryDatabase{client: &redisClient}
			// Test WriteString
			err := memoryDatabase.WriteString(ctx, "anykey", "anyvalue", 0)
			if err != nil {
				t.Errorf("WriteString with redisClient initiated should not fail")
			}
			// Test ReadString withinexistent key
			_, readError := memoryDatabase.ReadString(ctx, "inexistentkey")
			if readError == nil {
				t.Error("ReadString from initiated redisClient should fail with inexistent key.")
			}
			// Test ReadString with key "anykey"
			// Test ReadString withinexistent key
			readedValue, readExistentKeyError := memoryDatabase.ReadString(ctx, "anykey")
			if readExistentKeyError != nil {
				t.Errorf("ReadString from initiated redisClient shouldn't fail reading 'anykey'.")
			}
			if readedValue != "anyvalue" {
				t.Errorf("ReadString from initiated redisClient should return 'anyvalue' but '%s' was returned", readedValue)
			}

		}
	}
}
