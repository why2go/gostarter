package goredisstarter

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/why2go/gostarter/config"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

var (
	logger         = log.With().Str("ltag", "redisStarter").Logger()
	clients        = make(map[string]*redis.Client)
	clusterClients = make(map[string]*redis.ClusterClient)
)

var (
	ErrClientNotFound = errors.New("client not found")
)

// 支持server client， cluster client

type redisConfig struct {
	Clients        map[string]*clientConfig        `yaml:"clients" json:"clients"`
	ClusterClients map[string]*clusterClientConfig `yaml:"cluster_clients" json:"cluster_clients"`
}

func (redisConfig) ConfigName() string {
	return "go_redis"
}

type clientConfig struct {
	ConnUrl string `yaml:"conn_url" json:"conn_url"`
}

type clusterClientConfig struct {
	ConnUrl string `yaml:"conn_url" json:"conn_url"`
}

func init() {
	var err error
	cfg := &redisConfig{}
	err = config.GetConfig(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("load redis config failed")
		return
	}
	// redis server client
	for k, v := range cfg.Clients {
		opts, err := redis.ParseURL(strings.TrimSpace(v.ConnUrl))
		if err != nil {
			logger.Fatal().Err(err).Msgf("malformed redis conn url: %s", v.ConnUrl)
			return
		}
		client := redis.NewClient(opts)
		ctx, cf := context.WithTimeout(context.Background(), time.Second)
		sc := client.Ping(ctx)
		cf()
		if sc.Err() != nil {
			logger.Fatal().Err(err).Msgf("failed to ping redis: %s", k)
			return
		}
		k = strings.TrimSpace(k)
		clients[k] = client
		logger.Info().Msgf("successfully connected to redis: %s", k)
	}
	// redis cluster client
	for k, v := range cfg.ClusterClients {
		opts, err := redis.ParseClusterURL(strings.TrimSpace(v.ConnUrl))
		if err != nil {
			logger.Fatal().Err(err).Msgf("malformed redis cluster conn url: %s", v.ConnUrl)
			return
		}
		cc := redis.NewClusterClient(opts)
		ctx, cf := context.WithTimeout(context.Background(), time.Second)
		sc := cc.Ping(ctx)
		cf()
		if sc.Err() != nil {
			logger.Fatal().Err(err).Msgf("failed to ping redis cluster: %s", k)
			return
		}
		k = strings.TrimSpace(k)
		clusterClients[k] = cc
		logger.Info().Msgf("successfully connected to redis cluster: %s", k)
	}
}

// 获取已经建立连接的redis客户端
func GetRedisClient(which string) (*redis.Client, error) {
	if c, ok := clients[which]; ok {
		return c, nil
	} else {
		return nil, ErrClientNotFound
	}
}

// 获取已经建立连接的redis cluster客户端
func GetRedisClusterClient(which string) (*redis.ClusterClient, error) {
	if c, ok := clusterClients[which]; ok {
		return c, nil
	} else {
		return nil, ErrClientNotFound
	}
}
