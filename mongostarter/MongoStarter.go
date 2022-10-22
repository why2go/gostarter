package mongostarter

import (
	"context"
	"time"

	config "github.com/why2go/gostarter/config"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	Client *mongoClient
	logger = log.With().Str("ltag", "mongoStarter").Logger()
)

func init() {
	cfg := &mongoConf{}
	err := config.GetConfig(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("get mongo config failed")
		return
	}
	Client = newMongoClient(cfg)
}

type mongoClient = mongo.Client

func newMongoClient(cfg *mongoConf) *mongoClient {
	logger.Info().Msg("connecting to mongo server...")
	opts := options.Client().ApplyURI(cfg.ConnectionString)
	client, err := mongo.NewClient(opts)
	if err != nil {
		logger.Fatal().Err(err).Msg("new mongo client failed")
		return nil
	}
	err = client.Connect(context.TODO())
	if err != nil {
		logger.Fatal().Err(err).Msg("can't connect to mongo server")
		return nil
	}
	// 使用ping来测试连接是否成功
	connTimeout := 2 * time.Second
	if *opts.ConnectTimeout != 0 {
		connTimeout = *opts.ConnectTimeout
	}
	pingCtx, pingCf := context.WithTimeout(context.Background(), connTimeout)
	defer pingCf()
	err = client.Ping(pingCtx, readpref.Primary())
	if err != nil {
		logger.Fatal().Err(err).Msg("can't ping mongo server")
		return nil
	}
	logger.Info().Msg("successfully connect to mongo server!")

	return client
}

// 配置项
type mongoConf struct {
	ConnectionString string `yaml:"connectionString" json:"connectionString"`
}

func (cfg *mongoConf) GetConfigName() string {
	return "mongo"
}
