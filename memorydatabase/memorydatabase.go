package memorydatabase

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	redisconfig "github.com/a-castellano/go-types/redis"
	goredis "github.com/redis/go-redis/v9"
)

// Client interface operates against memory database instance
type Client interface {
	WriteString(context.Context, string, string, int) error
	ReadString(context.Context, string) (string, bool, error)
	IsClientInitiated() bool
}

// MemoryDatabase uses Client in order to operate against database instance
type MemoryDatabase struct {
	client Client
}

// NewMemoryDatabase creates MemoryDatabase instance
func NewMemoryDatabase(client Client) MemoryDatabase {
	memorydatabase := MemoryDatabase{client: client}

	return memorydatabase
}

// RedisClient is the real Redis client, it has Client methods
type RedisClient struct {
	config          *redisconfig.Config
	client          *goredis.Client
	clientInitiated bool
}

// NewRedisClient initiates RedisClient
func NewRedisClient(redisConfig *redisconfig.Config) RedisClient {
	redisClient := RedisClient{config: redisConfig}

	return redisClient
}

// IsClientInitiated returns if client has been initiated
func (client *RedisClient) IsClientInitiated() bool {
	return client.clientInitiated
}

// Initiate initiate RedisClient validating redis connection
func (client *RedisClient) Initiate(ctx context.Context) error {

	var actualHost string
	// check if config.Host is a IP
	if ip := net.ParseIP(client.config.Host); ip != nil {
		// Host is a valid IP
		actualHost = client.config.Host
	} else {
		// client.config.Host is a domain
		ips, lookupErr := net.LookupIP(client.config.Host)
		if lookupErr != nil {
			return lookupErr
		}
		actualHost = fmt.Sprintf("%s", ips[0])
	}

	redisAddr := fmt.Sprintf("%s:%d", actualHost, client.config.Port)
	// Set redis config
	client.client = goredis.NewClient(&goredis.Options{
		Addr:     redisAddr,
		Password: client.config.Password,
		DB:       client.config.Database,
	})

	_, pingErr := client.client.Ping(ctx).Result()
	if pingErr != nil {
		return pingErr
	}

	client.clientInitiated = true
	return nil
}

// WriteString uses RedisClient for writing a string as value of required key in Redis
func (client *RedisClient) WriteString(ctx context.Context, key string, value string, ttl int) error {
	status := client.client.Set(ctx, key, value, time.Duration(ttl)*time.Second)
	if status == nil {
		return errors.New("Something wrong happened executing WriteString")
	}
	return status.Err()
}

// ReadString uses RedisClient for reading a string as value of required key in Redis
func (client *RedisClient) ReadString(ctx context.Context, key string) (string, bool, error) {
	var found bool = true
	readValue, err := client.client.Get(ctx, key).Result()
	if err != nil {
		found = false
		if err == goredis.Nil {
			return "", found, nil
		} else {
			return "", found, err
		}
	}
	return readValue, found, nil
}

// WriteString writes string in MemoryDatabase
func (memorydatabase *MemoryDatabase) WriteString(ctx context.Context, key string, value string, ttl int) error {
	if memorydatabase.client.IsClientInitiated() {
		return memorydatabase.client.WriteString(ctx, key, value, ttl)
	} else {
		return errors.New("client is not initiated, cannot perform WriteString operation")
	}
}

// ReadString writes string in MemoryDatabase
func (memorydatabase *MemoryDatabase) ReadString(ctx context.Context, key string) (string, bool, error) {
	if memorydatabase.client.IsClientInitiated() {
		return memorydatabase.client.ReadString(ctx, key)
	} else {
		return "", false, errors.New("client is not initiated, cannot perform ReadString operation")
	}
}
