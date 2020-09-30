package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisConfig defines how a RedisCache should be constructed.
type RedisConfig struct {
	Endpoint    string        `yaml:"endpoint"`
	MasterName  string        `yaml:"master_name"`
	Timeout     time.Duration `yaml:"timeout"`
	Expiration  time.Duration `yaml:"expiration"`
	DB          int           `yaml:"db"`
	PoolSize    int           `yaml:"pool_size"`
	Password    string        `yaml:"password"`
	EnableTLS   bool          `yaml:"enable_tls"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
	MaxConnAge  time.Duration `yaml:"max_connection_age"`
}

type RedisClient struct {
	expiration time.Duration
	timeout    time.Duration
	rdb        redis.UniversalClient
}

// NewRedisClient creates Redis client
func NewRedisClient(cfg *RedisConfig) *RedisClient {
	opt := &redis.UniversalOptions{
		Addrs:       strings.Split(cfg.Endpoint, ","),
		MasterName:  cfg.MasterName,
		Password:    cfg.Password,
		DB:          cfg.DB,
		PoolSize:    cfg.PoolSize,
		IdleTimeout: cfg.IdleTimeout,
		MaxConnAge:  cfg.MaxConnAge,
	}
	if cfg.EnableTLS {
		opt.TLSConfig = &tls.Config{}
	}
	return &RedisClient{
		expiration: cfg.Expiration,
		timeout:    cfg.Timeout,
		rdb:        redis.NewUniversalClient(opt),
	}
}

func (c *RedisClient) Ping(ctx context.Context) error {
	var cancel context.CancelFunc
	if c.timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	pong, err := c.rdb.Ping(ctx).Result()
	if err != nil {
		return err
	}
	if pong != "PONG" {
		return fmt.Errorf("redis: Unexpected PING response %q", pong)
	}
	return nil
}

func (c *RedisClient) MSet(ctx context.Context, keys []string, values []string) error {
	var cancel context.CancelFunc
	if c.timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	pipe := c.rdb.TxPipeline()
	for i := range keys {
		pipe.Set(ctx, keys[i], values[i], c.expiration)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (c *RedisClient) MGet(ctx context.Context, keys []string) ([]string, error) {
	var cancel context.CancelFunc
	if c.timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	cmd := c.rdb.MGet(ctx, keys...)
	if err := cmd.Err(); err != nil {
		return nil, err
	}

	ret := make([]string, len(keys))
	for i, val := range cmd.Val() {
		if val != nil {
			ret[i] = val.(string)
		}
	}
	return ret, nil
}

func (c *RedisClient) GetAll(ctx context.Context) ([]string, error) {
	var cancel context.CancelFunc
	if c.timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	keys, err := c.rdb.Keys(ctx, "*").Result()
	if err != nil {
		return nil, err
	}
	n := len(keys)
	if n == 0 {
		return nil, nil
	}

	cmd := c.rdb.MGet(ctx, keys...)
	if err := cmd.Err(); err != nil {
		return nil, err
	}

	ret := make([]string, 0, n)
	for i, val := range cmd.Val() {
		if val != nil {
			ret = append(ret, keys[i]+" "+val.(string))
		}
	}
	return ret, nil
}

func (c *RedisClient) Close() error {
	return c.rdb.Close()
}
