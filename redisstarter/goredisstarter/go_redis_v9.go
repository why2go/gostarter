package goredisstarter

import (
	"github.com/why2go/gostarter/config"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

var (
	logger  = log.With().Str("ltag", "redisStarter").Logger()
	clients = make(map[string]*redis.Client)
)

// 支持server client， cluster client

type redisConfig struct {
	Clients        map[string]*clientConfig        `yaml:"clients" json:"clients"`
	ClusterClients map[string]*clusterClientConfig `yaml:"cluster_clients" json:"cluster_clients"`
}

func (redisConfig) GetConfigName() string {
	return "go_redis"
}

type clientConfig struct {
	ConnUrl string `yaml:"conn_url" json:"conn_url"`
}

type clusterClientConfig struct {
}

type redisClient = redis.Client

func init() {
	cfg := &redisConfig{}
	err := config.GetConfig(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("load redis config failed")
		return
	}

	Client = newRedisClient(cfg)
}

func GetRedisClient(which string) *redis.Client {

}

// todo: support redis cluster
