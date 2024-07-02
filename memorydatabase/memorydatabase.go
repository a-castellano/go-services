package memorydatabase

import (
	"context"
	"errors"
	"fmt"
	"net"
	"regexp"
	"time"

	redisconfig "github.com/a-castellano/go-types/redis"
	goredis "github.com/redis/go-redis/v9"
)

// Client interface iperates against memory database instance
type Client interface {
	WriteString(context.Context, string, string, int) error
	ReadString(context.Context, string) (string, error)
	isClientInitiated() bool
}

// MemoryDatabase string uses Client in order to operate against database instance
type MemoryDatabase struct {
	client Client
}

// RedisClient isthe real Redis client, it has Client methods
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

// isClientInitiated returns if client has been initiated
func (client *RedisClient) isClientInitiated() bool {
	return client.clientInitiated
}

// Initiate initiate RedisClient validating redis connection
func (client *RedisClient) Initiate(ctx context.Context) error {

	var actualHost string
	// check if config.Host is a IP
	var validIPv4 = regexp.MustCompile(`^(?:25[0-5]|2[0-4]\d|[0-1]?\d{1,2})(?:\.(?:25[0-5]|2[0-4]\d|[0-1]?\d{1,2})){3}$`)
	var validIPv6 = regexp.MustCompile(`^([[:xdigit:]]{1,4}(?::[[:xdigit:]]{1,4}){7}|::|:(?::[[:xdigit:]]{1,4}){1,6}|[[:xdigit:]]{1,4}:(?::[[:xdigit:]]{1,4}){1,5}|(?:[[:xdigit:]]{1,4}:){2}(?::[[:xdigit:]]{1,4}){1,4}|(?:[[:xdigit:]]{1,4}:){3}(?::[[:xdigit:]]{1,4}){1,3}|(?:[[:xdigit:]]{1,4}:){4}(?::[[:xdigit:]]{1,4}){1,2}|(?:[[:xdigit:]]{1,4}:){5}:[[:xdigit:]]{1,4}|(?:[[:xdigit:]]{1,4}:){1,6}:)$`)

	if !validIPv4.MatchString(client.config.Host) && !validIPv6.MatchString(client.config.Host) {
		// client.config.Host is a domain
		ips, lookupErr := net.LookupIP(client.config.Host)
		if lookupErr != nil {
			return lookupErr
		}
		actualHost = fmt.Sprintf("%s", ips[0])
	} else {
		// Host is a IP
		actualHost = client.config.Host
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
		return errors.New("Something worng happend executing WriteString")
	}
	return status.Err()
}

// ReadString uses RedisClient for reading a string as value of required key in Redis
func (client *RedisClient) ReadString(ctx context.Context, key string) (string, error) {
	readedValue, err := client.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return readedValue, nil
}

// WriteString writes string in MemoryDatabase
func (memorydatabase *MemoryDatabase) WriteString(ctx context.Context, key string, value string, ttl int) error {
	if memorydatabase.client.isClientInitiated() {
		return memorydatabase.client.WriteString(ctx, key, value, ttl)
	} else {
		return errors.New("Client is not initiated, cannot perform WriteString operation.")
	}
}

// ReadString writes string in MemoryDatabase
func (memorydatabase *MemoryDatabase) ReadString(ctx context.Context, key string) (string, error) {
	if memorydatabase.client.isClientInitiated() {
		return memorydatabase.client.ReadString(ctx, key)
	} else {
		return "", errors.New("Client is not initiated, cannot perform ReadString operation.")
	}
}
