package gormstarter

import (
	"errors"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/why2go/gostarter/config"
	gormLogger "github.com/why2go/gostarter/gormstarter/logger"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

var (
	dataSourceMap map[string]*gorm.DB
	logger        = log.With().Str("ltag", "gormStarter").Logger()
)

func init() {
	var cfg gormConfig
	err := config.GetConfig(&cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("load gorm config failed")
		return
	}
	if len(cfg) == 0 {
		logger.Fatal().Msg("no data source config found")
	}
	dataSourceMap = make(map[string]*gorm.DB)
	for srcName, srcCfg := range cfg {
		dataSourceMap[srcName] = newGormDB(srcName, srcCfg)
	}
}

var (
	ErrSourceNotFound = errors.New("source not found")
)

// 支持多个数据源配置，使用配置的数据源名字来获取数据源client，参数srcName大小写敏感
func GetDbBySourceName(srcName string) (*gorm.DB, error) {
	if src, ok := dataSourceMap[srcName]; ok {
		return src, nil
	} else {
		return nil, ErrSourceNotFound
	}
}

type gormConfig map[string]*dataSourceConfig

type dataSourceConfig struct {
	DBType          string                   `yaml:"dbType" json:"dbType"`
	DSN             string                   `yaml:"dsn" json:"dsn"`
	ConnMaxIdleTime string                   `yaml:"connMaxIdleTime" json:"connMaxIdleTime"`
	ConnMaxLifetime string                   `yaml:"connMaxLifeTime" json:"connMaxLifeTime"`
	MaxIdleConns    *int                     `yaml:"maxIdleConns" json:"maxIdleConns"`
	MaxOpenConns    *int                     `yaml:"maxOpenConns" json:"maxOpenConns"`
	Logger          *gormLogger.LoggerConfig `yaml:"logger" json:"logger"`
}

func (cfg *gormConfig) GetConfigName() string {
	return "gorm"
}

const (
	dbTypeMysql     = "mysql"
	dbTypePostgres  = "postgres"
	dbTypeSqlite    = "sqlite"
	dbTypeSqlServer = "sqlserver"
)

func newGormDB(srcName string, cfg *dataSourceConfig) *gorm.DB {
	if cfg == nil {
		log.Fatal().Msg("gorm config is nil")
		return nil
	}
	if len(cfg.DBType) == 0 {
		cfg.DBType = dbTypeMysql
	}
	logger.Info().Msgf("trying to connect to data source: %s...", srcName)
	// 设置数据库类型
	var dialector gorm.Dialector
	switch strings.ToLower(cfg.DBType) {
	case dbTypeMysql:
		dialector = mysql.Open(cfg.DSN)
	case dbTypePostgres:
		dialector = postgres.Open(cfg.DSN)
	case dbTypeSqlite:
		dialector = sqlite.Open("gorm.db")
	case dbTypeSqlServer:
		dialector = sqlserver.Open(cfg.DSN)
	default:
		logger.Fatal().Msgf("unsupport db type: %s", cfg.DBType)
		return nil
	}
	db, err := gorm.Open(dialector, &gorm.Config{Logger: gormLogger.NewZeroLogger(cfg.Logger)})
	if err != nil {
		logger.Fatal().Err(err).Msgf("can't connect to %s server", cfg.DBType)
		return nil
	}
	sqlDb, err := db.DB()
	if err != nil {
		logger.Fatal().Err(err).Msgf("can't connect to %s server", cfg.DBType)
		return nil
	}
	connMaxIdleTime := 30 * time.Minute // 默认30min
	var connMaxLifetime time.Duration   // 默认没有限制
	maxOpenConns := 10                  // 默认10个活跃连接
	maxIdleConns := 1                   // 默认1个空闲连接
	if len(cfg.ConnMaxIdleTime) != 0 {
		d, err := time.ParseDuration(cfg.ConnMaxIdleTime)
		if err != nil {
			logger.Fatal().Err(err).Msgf("invalid duration expression: %s", cfg.ConnMaxIdleTime)
			return nil
		}
		connMaxIdleTime = d
	}
	if len(cfg.ConnMaxLifetime) != 0 {
		d, err := time.ParseDuration(cfg.ConnMaxLifetime)
		if err != nil {
			logger.Fatal().Err(err).Msgf("invalid duration expression: %s", cfg.ConnMaxLifetime)
			return nil
		}
		connMaxLifetime = d
	}
	if cfg.MaxOpenConns != nil {
		maxOpenConns = *cfg.MaxOpenConns
	}
	if cfg.MaxIdleConns != nil {
		maxIdleConns = *cfg.MaxIdleConns
	}
	sqlDb.SetConnMaxIdleTime(connMaxIdleTime)
	sqlDb.SetConnMaxLifetime(connMaxLifetime)
	sqlDb.SetMaxOpenConns(maxOpenConns)
	sqlDb.SetMaxIdleConns(maxIdleConns)
	logger.Info().Msgf("successfully connect to data source: %s!", srcName)
	return db
}
