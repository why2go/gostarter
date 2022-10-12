package gormstarter

import (
	golog "log"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/why2go/gostarter/config"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// todo: 替换gorm的日志记录

var (
	Client     *gorm.DB
	zerologger = log.With().Str("ltag", "gormStarter").Logger()
)

func init() {
	cfg := &gormConfig{}
	err := config.GetConfig(cfg)
	if err != nil {
		zerologger.Fatal().Msg("load gorm config failed")
		return
	}
	Client = newGormDB(cfg)
}

type gormConfig struct {
	DBType                string `yaml:"dbType"`
	DSN                   string `yaml:"dsn"`
	LogMode               string `yaml:"logMode"`
	SkipErrRecordNotFound bool   `yaml:"skipErrRecordNotFound"`
	ConnMaxIdleTime       string `yaml:"connMaxIdleTime"`
	ConnMaxLifetime       string `yaml:"connMaxLifetime"`
	MaxIdleConns          *int   `yaml:"maxIdleConns"`
	MaxOpenConns          *int   `yaml:"maxOpenConns"`
	SlowThreshold         string `yaml:"slowThreshold"`
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

func newGormDB(cfg *gormConfig) *gorm.DB {
	if cfg == nil {
		log.Fatal().Msg("gorm config is nil")
		return nil
	}
	if len(cfg.DBType) == 0 {
		cfg.DBType = dbTypeMysql
	}
	zerologger.Info().Msgf("trying to connect to %s server...", cfg.DBType)
	// 设置logMode
	var logMode logger.LogLevel
	switch strings.ToLower(cfg.LogMode) {
	case "silent":
		logMode = logger.Silent
	case "error":
		logMode = logger.Error
	case "warn":
		logMode = logger.Warn
	case "info":
		logMode = logger.Info
	default:
		logMode = logger.Silent
	}
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
	// 慢查询日志预支
	SlowThreshold := 500 * time.Millisecond
	if len(cfg.SlowThreshold) != 0 {
		d, err := time.ParseDuration(cfg.SlowThreshold)
		if err != nil {
			zerologger.Fatal().Err(err).Msgf("invalid duration expression: %s", cfg.SlowThreshold)
			return nil
		}
		SlowThreshold = d
	}
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.New(golog.New(os.Stdout, "\r\n", golog.LstdFlags), logger.Config{
			SlowThreshold:             SlowThreshold,
			LogLevel:                  logMode,
			IgnoreRecordNotFoundError: cfg.SkipErrRecordNotFound,
			Colorful:                  false,
		}),
	})
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
