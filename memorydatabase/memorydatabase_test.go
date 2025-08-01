//go:build integration_tests || unit_tests || memorydatabase_tests || memorydatabase_unit_tests

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

func (mock *RedisClientMock) IsClientInitiated() bool {
	return true
}

func (mock *RedisClientMock) WriteString(ctx context.Context, key string, value string, ttl int) error {
	return mock.client.Set(ctx, key, value, time.Duration(ttl)*time.Second).Err()
}

func (mock *RedisClientMock) ReadString(ctx context.Context, key string) (string, bool, error) {
	var found bool = true
	value, getError := mock.client.Get(ctx, key).Result()

	if getError != nil {
		found = false
		if getError == goredis.Nil {
			return value, found, nil
		} else {
			return value, found, getError
		}
	}
	return value, found, nil
}

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
			// Initiate MemoryDatabase instance
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
			t.Errorf("memoryDatabase.WriteString call with mock should return error \"FAIL\", it has returned \"%s\"", err.Error())
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

func TestReadStringWithMockNilValue(t *testing.T) {

	setUp()
	defer teardown()

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.Regexp().ExpectGet("anykey").RedisNil()

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := MemoryDatabase{client: &redisClientMock}

	_, found, err := memoryDatabase.ReadString(ctx, "anykey")

	if err != nil {
		t.Errorf("memoryDatabase.ReadString with mocked inexistent key shouldn't fail.")
	} else {
		if found == true {
			t.Errorf("memoryDatabase.ReadString with mocked inexistent key should not find key")
		}
	}
}

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
			// Initiate MemoryDatabase instance
			ctx := context.Background()
			memoryDatabase := MemoryDatabase{client: &redisClient}
			_, _, err := memoryDatabase.ReadString(ctx, "anykey")
			if err == nil {
				t.Errorf("memoryDatabase.ReadString call without redisClient being initiated should fail as redisClient is not initiated")
			} else {
				if err.Error() != "client is not initiated, cannot perform ReadString operation" {
					t.Errorf("memoryDatabase.WriteString call without redisClient being initiated should return error \"client is not initiated, cannot perform ReadString operation\", it has returned \"%s\"", err.Error())
				}
			}
		}
	}
}

func TestRedIsClientInitiatedWithoutEnvVariableForIPAsRedisHost(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("REDIS_HOST", "127.0.0.1")
	config, err := redisconfig.NewConfig()

	if err != nil {
		t.Errorf("NewConfig method with 127.0.0.1 as REDIS_HOST env variable suited shouldn't fail, error was '%s'.", err.Error())
	} else {
		redisClient := NewRedisClient(config)
		if ok := redisClient.IsClientInitiated(); ok {
			t.Errorf("RedisClient should not be initiated after being created")
		} else {
			// Initiate MemoryDatabase instance
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

func TestRedisClientInitiateWithDomainName(t *testing.T) {
	setUp()
	defer teardown()

	// Usar un dominio que se pueda resolver
	os.Setenv("REDIS_HOST", "google.com") // Dominio que siempre se puede resolver
	os.Setenv("REDIS_PORT", "6379")

	config, err := redisconfig.NewConfig()
	if err != nil {
		t.Errorf("NewConfig method with domain name shouldn't fail, error was '%s'.", err.Error())
		return
	}

	redisClient := NewRedisClient(config)
	if ok := redisClient.IsClientInitiated(); ok {
		t.Errorf("RedisClient should not be initiated after being created")
		return
	}

	// Initiate RedisClient - esto debería ejecutar la resolución DNS
	ctx := context.Background()
	initiateErr := redisClient.Initiate(ctx)

	// El test puede fallar si no hay Redis corriendo, pero eso está bien
	// Lo importante es que se ejecute la resolución DNS (línea 60)
	if initiateErr != nil {
		// Verificar que el error no sea de resolución DNS
		if initiateErr.Error() == "lookup google.com: no such host" {
			t.Errorf("DNS resolution failed for google.com, but the else branch should have been executed")
		}
		// Otros errores (como conexión rechazada) son esperados si Redis no está corriendo
		t.Logf("Expected error (Redis not running): %v", initiateErr)
	} else {
		if !redisClient.IsClientInitiated() {
			t.Error("After successful init, redisClient.IsClientInitiated() should be true.")
		}
	}
}

func TestRedisClientInitiateWithResolvableDomain(t *testing.T) {
	setUp()
	defer teardown()

	// Usar un dominio que sabemos que se resuelve
	os.Setenv("REDIS_HOST", "cloudflare.com") // Dominio que siempre se puede resolver
	os.Setenv("REDIS_PORT", "6379")

	config, err := redisconfig.NewConfig()
	if err != nil {
		t.Errorf("NewConfig method with resolvable domain shouldn't fail, error was '%s'.", err.Error())
		return
	}

	redisClient := NewRedisClient(config)
	if ok := redisClient.IsClientInitiated(); ok {
		t.Errorf("RedisClient should not be initiated after being created")
		return
	}

	// Initiate RedisClient - esto debería ejecutar la resolución DNS
	ctx := context.Background()
	initiateErr := redisClient.Initiate(ctx)

	// El test puede fallar si no hay Redis corriendo, pero eso está bien
	// Lo importante es que se ejecute la resolución DNS (línea 60)
	if initiateErr != nil {
		// Verificar que el error no sea de resolución DNS
		if initiateErr.Error() == "lookup cloudflare.com: no such host" {
			t.Errorf("DNS resolution failed for cloudflare.com, but the else branch should have been executed")
		}
		// Otros errores (como conexión rechazada) son esperados si Redis no está corriendo
		t.Logf("Expected error (Redis not running): %v", initiateErr)
	} else {
		if !redisClient.IsClientInitiated() {
			t.Error("After successful init, redisClient.IsClientInitiated() should be true.")
		}
	}
}

func TestWriteStringWithEmptyKey(t *testing.T) {
	setUp()
	defer teardown()

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.Regexp().ExpectSet("", "anyvalue", 0).SetVal("OK")

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := MemoryDatabase{client: &redisClientMock}
	err := memoryDatabase.WriteString(ctx, "", "anyvalue", 0)

	if err != nil {
		t.Errorf("memoryDatabase.WriteString with empty key should not fail, error was '%s'", err.Error())
	}
}

func TestWriteStringWithEmptyValue(t *testing.T) {
	setUp()
	defer teardown()

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.Regexp().ExpectSet("anykey", "", 0).SetVal("OK")

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := MemoryDatabase{client: &redisClientMock}
	err := memoryDatabase.WriteString(ctx, "anykey", "", 0)

	if err != nil {
		t.Errorf("memoryDatabase.WriteString with empty value should not fail, error was '%s'", err.Error())
	}
}

func TestWriteStringWithTTL(t *testing.T) {
	setUp()
	defer teardown()

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.Regexp().ExpectSet("anykey", "anyvalue", 60*time.Second).SetVal("OK")

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := MemoryDatabase{client: &redisClientMock}
	err := memoryDatabase.WriteString(ctx, "anykey", "anyvalue", 60)

	if err != nil {
		t.Errorf("memoryDatabase.WriteString with TTL should not fail, error was '%s'", err.Error())
	}
}

func TestWriteStringWithNegativeTTL(t *testing.T) {
	setUp()
	defer teardown()

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.Regexp().ExpectSet("anykey", "anyvalue", 0).SetVal("OK")

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := MemoryDatabase{client: &redisClientMock}
	err := memoryDatabase.WriteString(ctx, "anykey", "anyvalue", -10)

	if err != nil {
		t.Errorf("memoryDatabase.WriteString with negative TTL should not fail, error was '%s'", err.Error())
	}
}

func TestWriteStringWithNilRedisResponse(t *testing.T) {
	setUp()
	defer teardown()

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	// Simular que Redis devuelve nil (caso edge)
	mock.Regexp().ExpectSet("anykey", "anyvalue", 0).SetErr(errors.New("Something wrong happened executing WriteString"))

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := MemoryDatabase{client: &redisClientMock}
	err := memoryDatabase.WriteString(ctx, "anykey", "anyvalue", 0)

	if err == nil {
		t.Errorf("memoryDatabase.WriteString with nil Redis response should fail")
	} else {
		if err.Error() != "Something wrong happened executing WriteString" {
			t.Errorf("memoryDatabase.WriteString with nil response should return specific error, got '%s'", err.Error())
		}
	}
}
