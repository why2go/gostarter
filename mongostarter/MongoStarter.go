package mongostarter

import (
	"context"
	"errors"
	"strings"
	"time"

	config "github.com/why2go/gostarter/config"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	clients = make(map[string]*mongo.Client)
	logger  = log.With().Str("ltag", "mongoStarter").Logger()
)

var (
	ErrMongoClientNotFound = errors.New("mongo client not found")
)

func init() {
	cfg := &mongoConf{}
	err := config.GetConfig(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("load mongo config failed")
		return
	}
	for k, cc := range cfg.ClientsConfig {
		c, err := newMongoClient(k, strings.TrimSpace(cc.ConnectionString))
		if err != nil {
			logger.Fatal().Err(err).Msgf("failed to connect to mongo: %s", k)
			return
		}
		clients[strings.TrimSpace(k)] = c
	}
}

func GetMongoClient(which string) (*mongo.Client, error) {
	if c, ok := clients[which]; ok {
		return c, nil
	}
	return nil, ErrMongoClientNotFound
}

func newMongoClient(clientName, connStr string) (*mongo.Client, error) {
	logger.Info().Msgf("try to connect to mongo server: %s", clientName)
	opts := options.Client().ApplyURI(connStr)
	client, err := mongo.NewClient(opts)
	if err != nil {
		return nil, err
	}
	connTimeout := 2 * time.Second
	if *opts.ConnectTimeout != 0 {
		connTimeout = *opts.ConnectTimeout
	}
	ctx, cf := context.WithTimeout(context.Background(), connTimeout)
	err = client.Connect(ctx)
	cf()
	if err != nil {
		return nil, err
	}
	// 使用ping来测试连接是否成功
	pingCtx, pingCf := context.WithTimeout(context.Background(), connTimeout)
	defer pingCf()
	err = client.Ping(pingCtx, readpref.Primary())
	if err != nil {
		return nil, err
	}
	logger.Info().Msgf("successfully connect to mongo server: %s", clientName)

	return client, nil
}

// 配置项
type mongoConf struct {
	ClientsConfig map[string]*clientConfig `yaml:"clients" json:"clients" toml:"clients"`
}

func (cfg *mongoConf) ConfigName() string {
	return "mongo"
}

type clientConfig struct {
	ConnectionString string `yaml:"connectionString" json:"connectionString"`
}
