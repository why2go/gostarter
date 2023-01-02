package gormstarter

import (
	"errors"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/why2go/gostarter/config"
	mylogger "github.com/why2go/gostarter/gormstarter/logger"
	_ "github.com/why2go/gostarter/zerologstarter"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

// 支持多数据源配置，示例如下：
// gorm:
//   db0:
//     dbType: mysql
//     dsn: root:root@tcp(127.0.0.1:3307)/sakila?charset=utf8mb4&parseTime=True&loc=Local
//     connMaxIdleTime: 10m # Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
//     connMaxLifetime: 20m
//     maxIdleConns: 5
//     maxOpenConns: 20
//     logger:
//       logMode: info
//       ignoreErrRecordNotFound: true
//       slowThresholdMS: 200

var (
	clients    map[string]*gorm.DB
	zerologger = log.With().Str("ltag", "gormStarter").Logger()
)

func init() {
	cfg := &gormConfig{}
	err := config.GetConfig(cfg)
	if err != nil {
		zerologger.Fatal().Err(err).Msg("load gorm config failed")
		return
	}
	if len(cfg.Dbs) == 0 {
		zerologger.Fatal().Msg("no data source config found")
	}
	clients = make(map[string]*gorm.DB)
	for srcName, srcCfg := range cfg.Dbs {
		clients[srcName] = newGormDB(srcCfg)
	}
}

var (
	ErrSourceNotFound = errors.New("source not found")
)

// 支持多个数据源配置，使用配置的数据源名字来获取数据源client，参数srcName大小写敏感
func GetDbBySourceName(srcName string) (*gorm.DB, error) {
	if src, ok := clients[srcName]; ok {
		return src, nil
	} else {
		return nil, ErrSourceNotFound
	}
}

type gormConfig struct {
	Dbs map[string]*dataSourceConfig
}

type dataSourceConfig struct {
	DBType          string                 `yaml:"dbType" json:"dbType"`
	DSN             string                 `yaml:"dsn" json:"dsn"`
	ConnMaxIdleTime string                 `yaml:"connMaxIdleTime" json:"connMaxIdleTime"`
	ConnMaxLifetime string                 `yaml:"connMaxLifetime" json:"connMaxLifeTime"`
	MaxIdleConns    *int                   `yaml:"maxIdleConns" json:"maxIdleConns"`
	MaxOpenConns    *int                   `yaml:"maxOpenConns" json:"maxOpenConns"`
	Logger          *mylogger.LoggerConfig `yaml:"logger" json:"logger"`
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

func newGormDB(cfg *dataSourceConfig) *gorm.DB {
	if cfg == nil {
		log.Fatal().Msg("gorm config is nil")
		return nil
	}
	if len(cfg.DBType) == 0 {
		cfg.DBType = dbTypeMysql
	}
	zerologger.Info().Msgf("trying to connect to %s server...", cfg.DBType)
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
		zerologger.Fatal().Msgf("unsupport db type: %s", cfg.DBType)
		return nil
	}
	db, err := gorm.Open(dialector, &gorm.Config{Logger: mylogger.NewZeroLogger(cfg.Logger)})
	if err != nil {
		zerologger.Fatal().Err(err).Msgf("can't connect to %s server", cfg.DBType)
		return nil
	}
	sqlDb, err := db.DB()
	if err != nil {
		zerologger.Fatal().Err(err).Msgf("can't connect to %s server", cfg.DBType)
		return nil
	}
	connMaxIdleTime := 30 * time.Minute // 默认30min
	var connMaxLifetime time.Duration   // 默认没有限制
	maxOpenConns := 10                  // 默认10个活跃连接
	maxIdleConns := 1                   // 默认1个空闲连接
	if len(cfg.ConnMaxIdleTime) != 0 {
		d, err := time.ParseDuration(cfg.ConnMaxIdleTime)
		if err != nil {
			zerologger.Fatal().Err(err).Msgf("invalid duration expression: %s", cfg.ConnMaxIdleTime)
			return nil
		}
		connMaxIdleTime = d
	}
	if len(cfg.ConnMaxLifetime) != 0 {
		d, err := time.ParseDuration(cfg.ConnMaxLifetime)
		if err != nil {
			zerologger.Fatal().Err(err).Msgf("invalid duration expression: %s", cfg.ConnMaxLifetime)
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
	zerologger.Info().Msgf("successfully connect to %s server!", cfg.DBType)
	return db
}
