package redisstarter

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/why2go/gostarter/config"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
)

var (
	logger = log.With().Str("ltag", "redisStarter").Logger()
	Client *redisClient
)

type redisClient = redis.Client

func init() {
	cfg := &redisConf{}
	err := config.GetConfig(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("load redis config failed")
		return
	}

	Client = newRedisClient(cfg)
}

// Redis配置
type redisConf struct {
	Address  *string `yaml:"address"`
	Password *string `yaml:"password"`
	DB       *int    `yaml:"db"`
	// Maximum number of retries before giving up.
	// Default is 3 retries; -1 (not 0) disables retries.
	MaxRetries *int `yaml:"maxRetries"`
	// Minimum backoff between each retry.
	// Default is 8 milliseconds; -1 disables backoff.
	MinRetryBackoff *string `yaml:"minRetryBackoff"`
	// Maximum backoff between each retry.
	// Default is 512 milliseconds; -1 disables backoff.
	MaxRetryBackoff *string `yaml:"maxRetryBackoff"`
	// Dial timeout for establishing new connections.
	// Default is 5 seconds.
	DialTimeout *string `yaml:"dialTimeout"`
	// Timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking. Use value -1 for no timeout and 0 for default.
	// Default is 3 seconds.
	ReadTimeout *string `yaml:"readTimeout"`
	// Timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.
	// Default is ReadTimeout.
	WriteTimeout *string `yaml:"writeTimeout"`
	// Type of connection pool.
	// true for FIFO pool, false for LIFO pool.
	// Note that fifo has higher overhead compared to lifo.
	PoolFIFO *bool `yaml:"poolFifo"`
	// Maximum number of socket connections.
	// Default is 10 connections per every available CPU as reported by runtime.GOMAXPROCS.
	PoolSize *int `yaml:"poolSize"`
	// Minimum number of idle connections which is useful when establishing
	// new connection is slow.
	MinIdleConns *int `yaml:"minIdleConns"`
	// Connection age at which client retires (closes) the connection.
	// Default is to not close aged connections.
	MaxConnAge *string `yaml:"maxConnAge"`
	// Amount of time client waits for connection if all connections
	// are busy before returning an error.
	// Default is ReadTimeout + 1 second.
	PoolTimeout *string `yaml:"poolTimeout"`
	// Amount of time after which client closes idle connections.
	// Should be less than server's timeout.
	// Default is 5 minutes. -1 disables idle timeout check.
	IdleTimeout *string `yaml:"idleTimeout"`
	// Frequency of idle checks made by idle connections reaper.
	// Default is 1 minute. -1 disables idle connections reaper,
	// but idle connections are still discarded by the client
	// if IdleTimeout is set.
	IdleCheckFrequency *string `yaml:"idleCheckFrequency"`
	EnableTLS          bool    `yaml:"enableTLS"`
}

func (conf *redisConf) GetConfigName() string {
	return "redis"
}

func newRedisClient(cfg *redisConf) *redisClient {
	if cfg == nil {
		logger.Fatal().Msg("no config found for redis")
		return nil
	}
	log.Info().Msg("connecting to redis server...")
	// 对于standalone
	var opts redis.Options
	address := "localhost:6379"
	if cfg.Address == nil || len(*cfg.Address) == 0 {
		log.Warn().Msg("no redis server address found, use default \"localhost:6379\"")
	} else {
		address = *cfg.Address
	}
	opts.Addr = address
	if cfg.Password != nil {
		opts.Password = *cfg.Password
	}
	if cfg.DB != nil {
		opts.DB = *cfg.DB
	}
	if cfg.MaxRetries != nil {
		opts.MaxRetries = *cfg.MaxRetries
	}
	if cfg.MinRetryBackoff != nil {
		d, err := time.ParseDuration(*cfg.MinRetryBackoff)
		if err != nil {
			log.Fatal().Msgf("invalid duration expression: %s", *cfg.MinRetryBackoff)
			return nil
		}
		opts.MinRetryBackoff = d
	}
	if cfg.MaxRetryBackoff != nil {
		d, err := time.ParseDuration(*cfg.MaxRetryBackoff)
		if err != nil {
			log.Fatal().Msgf("invalid duration expression: %s", *cfg.MaxRetryBackoff)
			return nil
		}
		opts.MaxRetryBackoff = d
	}
	if cfg.DialTimeout != nil {
		d, err := time.ParseDuration(*cfg.DialTimeout)
		if err != nil {
			log.Fatal().Msgf("invalid duration expression: %s", *cfg.DialTimeout)
			return nil
		}
		opts.DialTimeout = d
	}
	if cfg.ReadTimeout != nil {
		d, err := time.ParseDuration(*cfg.ReadTimeout)
		if err != nil {
			log.Fatal().Msgf("invalid duration expression: %s", *cfg.ReadTimeout)
			return nil
		}
		opts.ReadTimeout = d
	}
	if cfg.WriteTimeout != nil {
		d, err := time.ParseDuration(*cfg.WriteTimeout)
		if err != nil {
			log.Fatal().Msgf("invalid duration expression: %s", *cfg.WriteTimeout)
			return nil
		}
		opts.WriteTimeout = d
	}
	if cfg.PoolFIFO != nil {
		opts.PoolFIFO = *cfg.PoolFIFO
	}
	if cfg.PoolSize != nil {
		opts.PoolSize = *cfg.PoolSize
	}
	if cfg.MinIdleConns != nil {
		opts.MinIdleConns = *cfg.MinIdleConns
	}
	if cfg.MaxConnAge != nil {
		d, err := time.ParseDuration(*cfg.MaxConnAge)
		if err != nil {
			log.Fatal().Msgf("invalid duration expression: %s", *cfg.MaxConnAge)
			return nil
		}
		opts.MaxConnAge = d
	}
	if cfg.PoolTimeout != nil {
		d, err := time.ParseDuration(*cfg.PoolTimeout)
		if err != nil {
			log.Fatal().Msgf("invalid duration expression: %s", *cfg.PoolTimeout)
			return nil
		}
		opts.PoolTimeout = d
	}
	if cfg.IdleTimeout != nil {
		d, err := time.ParseDuration(*cfg.IdleTimeout)
		if err != nil {
			log.Fatal().Msgf("invalid duration expression: %s", *cfg.IdleTimeout)
			return nil
		}
		opts.IdleTimeout = d
	}
	if cfg.IdleCheckFrequency != nil {
		d, err := time.ParseDuration(*cfg.IdleCheckFrequency)
		if err != nil {
			log.Fatal().Msgf("invalid duration expression: %s", *cfg.IdleCheckFrequency)
			return nil
		}
		opts.IdleCheckFrequency = d
	}
	if cfg.EnableTLS { // 开启TLS
		opts.TLSConfig = &tls.Config{}
	}
	client := redis.NewClient(&opts)
	ctx, cancleFun := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancleFun()
	err := client.Ping(ctx).Err()
	if err != nil {
		log.Fatal().Err(err).Msg("can't connect to redis server")
	}
	log.Info().Msg("successfully connect to redis server!")
	return client
}

// todo: support redis cluster
