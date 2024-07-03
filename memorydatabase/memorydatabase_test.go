//go:build integration_tests || unit_tests || memorydatabase_tests || memorydatabase_unit_tests

package memorydatabase

import (
	"context"
	"errors"
	redisconfig "github.com/a-castellano/go-types/redis"
	redismock "github.com/go-redis/redismock/v9"
	goredis "github.com/redis/go-redis/v9"
	"os"
	"testing"
	"time"
)

var currentHost string
var currentHostDefined bool

var currentPort string
var currentPortDefined bool

var currentDatabase string
var currentDatabaseDefined bool

var currentPassword string
var currentPasswordDefined bool

func setUp() {

	if envHost, found := os.LookupEnv("REDIS_HOST"); found {
		currentHost = envHost
		currentHostDefined = true
	} else {
		currentHostDefined = false
	}

	if envPort, found := os.LookupEnv("REDIS_PORT"); found {
		currentPort = envPort
		currentPortDefined = true
	} else {
		currentPortDefined = false
	}

	if envDatabase, found := os.LookupEnv("REDIS_DATABASE"); found {
		currentDatabase = envDatabase
		currentDatabaseDefined = true
	} else {
		currentDatabaseDefined = false
	}

	if envPassword, found := os.LookupEnv("REDIS_PASSWORD"); found {
		currentPassword = envPassword
		currentPasswordDefined = true
	} else {
		currentPasswordDefined = false
	}

	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("REDIS_PORT")
	os.Unsetenv("REDIS_DATABASE")
	os.Unsetenv("REDIS_PASSWORD")

}

func teardown() {

	if currentHostDefined {
		os.Setenv("REDIS_HOST", currentHost)
	} else {
		os.Unsetenv("REDIS_HOST")
	}

	if currentPortDefined {
		os.Setenv("REDIS_PORT", currentPort)
	} else {
		os.Unsetenv("REDIS_PORT")
	}

	if currentDatabaseDefined {
		os.Setenv("REDIS_DATABASE", currentDatabase)
	} else {
		os.Unsetenv("REDIS_DATABASE")
	}

	if currentPasswordDefined {
		os.Setenv("REDIS_PASSWORD", currentPassword)
	} else {
		os.Unsetenv("REDIS_PASSWORD")
	}

}

type RedisClientMock struct {
	client *goredis.Client
}

func (mock *RedisClientMock) isClientInitiated() bool {
	return true
}

func (mock *RedisClientMock) WriteString(ctx context.Context, key string, value string, ttl int) error {
	return mock.client.Set(ctx, key, value, time.Duration(ttl)*time.Second).Err()
}

func (mock *RedisClientMock) ReadString(ctx context.Context, key string) (string, error) {
	value, getError := mock.client.Get(ctx, key).Result()
	return value, getError
}

func TestRedisClientInitiatedWithoutEnvVariablesWriteNotInitiated(t *testing.T) {

	setUp()
	defer teardown()

	config, err := redisconfig.NewConfig()

	if err != nil {
		t.Errorf("NewConfig method without any env varible suited shouldn't fail, error was '%s'.", err.Error())
	} else {
		redisClient := NewRedisClient(config)
		if ok := redisClient.isClientInitiated(); ok {
			t.Errorf("RedisClient should not be initiated after being created")
		} else {
			// Initiate MemoryDatabase instance
			ctx := context.Background()
			memoryDatabase := MemoryDatabase{client: &redisClient}
			err := memoryDatabase.WriteString(ctx, "anykey", "anyvalue", 0)
			if err == nil {
				t.Errorf("memoryDatabase.WriteString call without redisClient being initiated should fail as redisClient is not initiated")
			} else {
				if err.Error() != "Client is not initiated, cannot perform WriteString operation." {
					t.Errorf("memoryDatabase.WriteString call without redisClient being initiated should retrn error \"Client is not initiated, cannot perform WriteString operation.\", it has returned \"%s\"", err.Error())
				}
			}
		}
	}
}

func TestWriteStringWithMockSetFail(t *testing.T) {

	setUp()
	defer teardown()

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.Regexp().ExpectSet("anykey", "anyvalue", 0).SetErr(errors.New("FAIL"))

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := MemoryDatabase{client: &redisClientMock}
	err := memoryDatabase.WriteString(ctx, "anykey", "anyvalue", 0)

	if err == nil {
		t.Errorf("memoryDatabase.WriteString with mocked redis error should fail.")
	} else {
		if err.Error() != "FAIL" {
			t.Errorf("memoryDatabase.WriteString call with mock shoud return error \"FAIL\", it has returned \"%s\"", err.Error())
		}
	}
}

func TestReadStringWithMockSetValue(t *testing.T) {

	setUp()
	defer teardown()

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.Regexp().ExpectGet("anykey").SetVal("anyvalue")

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := MemoryDatabase{client: &redisClientMock}

	value, err := memoryDatabase.ReadString(ctx, "anykey")

	if err != nil {
		t.Errorf("memoryDatabase.ReadString with mocked value should not fail.")
	} else {
		if value != "anyvalue" {
			t.Errorf("memoryDatabase.ReadString with mocked value should return \"anyvalue\", it has returned \"%s\"", value)
		}
	}
}

func TestReadStringWithMockNilValue(t *testing.T) {

	setUp()
	defer teardown()

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.Regexp().ExpectGet("anykey").RedisNil()

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := MemoryDatabase{client: &redisClientMock}

	_, err := memoryDatabase.ReadString(ctx, "anykey")

	if err == nil {
		t.Errorf("memoryDatabase.ReadString with mocked inexistent key should fail.")
	} else {
		if err.Error() != "redis: nil" {
			t.Errorf("memoryDatabase.ReadString with mocked inexistent key should return error \"redis: nil\", it has returned \"%s\"", err.Error())
		}
	}
}

func TestRedisClientInitiatedWithoutEnvVariablesReadNotInitiated(t *testing.T) {

	setUp()
	defer teardown()

	config, err := redisconfig.NewConfig()

	if err != nil {
		t.Errorf("NewConfig method without any env varible suited shouldn't fail, error was '%s'.", err.Error())
	} else {
		redisClient := NewRedisClient(config)
		if ok := redisClient.isClientInitiated(); ok {
			t.Errorf("RedisClient should not be initiated after being created")
		} else {
			// Initiate MemoryDatabase instance
			ctx := context.Background()
			memoryDatabase := MemoryDatabase{client: &redisClient}
			_, err := memoryDatabase.ReadString(ctx, "anykey")
			if err == nil {
				t.Errorf("memoryDatabase.ReadString call without redisClient being initiated should fail as redisClient is not initiated")
			} else {
				if err.Error() != "Client is not initiated, cannot perform ReadString operation." {
					t.Errorf("memoryDatabase.WriteString call without redisClient being initiated should retrn error \"Client is not initiated, cannot perform ReadString operation.\", it has returned \"%s\"", err.Error())
				}
			}
		}
	}
}
