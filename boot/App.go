package boot

import (
	"context"

	"fmt"
	"os/signal"
	"reflect"
	"syscall"

	config "github.com/why2go/gostarter/config"
	_ "github.com/why2go/gostarter/zerologstarter"

	"github.com/rs/zerolog/log"
)

// 规定应用的启动过程，包括配置加载，执行启动函数，执行清理函数

const (
	defaultAppName     = "github.com/why2go/gostarter"
	defaultVersion     = "v0.0.1"
	defaultDescription = "this is a demo app"
)

var (
	logger      = log.With().Str("ltag", "boot").Logger()
	appInstance *app
)

func init() {
	cfg := &appConf{}
	err := config.GetConfig(cfg)
	if err != nil {
		if err == config.ErrCfgItemNotFound {
			cfg.AppName = defaultAppName
			cfg.Version = defaultVersion
			cfg.Description = defaultDescription
		} else {
			logger.Fatal().Err(err).Msg("load app config failed")
			return
		}
	}
	appInstance = newApp(cfg)
}

type appConf struct {
	AppName     string `yaml:"name" json:"name"`
	Author      string `yaml:"author" json:"author"`
	Version     string `yaml:"version" json:"version"`
	ChangeLog   string `yaml:"changeLog" json:"changeLog"`
	Description string `yaml:"description" json:"description"`
}

func (cfg *appConf) GetConfigName() string {
	return "app"
}

type app struct {
	name        string
	author      string
	version     string
	changeLog   string
	description string
	starters    []interface{}
	sweepers    []interface{}
}

func newApp(cfg *appConf) *app {
	app := &app{
		name:        cfg.AppName,
		author:      cfg.Author,
		version:     cfg.Version,
		changeLog:   cfg.ChangeLog,
		description: cfg.Description,
	}
	return app
}

// staters shall be "func()" or "func() error"
func AddStarters(starters ...interface{}) {
	for _, f := range starters {
		switch f.(type) {
		case func():
		case func() error:
		default:
			err := fmt.Errorf("invalid starter: %s", reflect.TypeOf(f))
			logger.Fatal().Err(err).Send()
		}
	}
	appInstance.starters = append(appInstance.starters, starters...)
}

func AddSweepers(sweepers ...interface{}) {
	for _, f := range sweepers {
		switch f.(type) {
		case func():
		case func() error:
		default:
			err := fmt.Errorf("invalid sweeper: %s", reflect.TypeOf(f))
			logger.Fatal().Err(err).Send()
		}
	}
	appInstance.sweepers = append(appInstance.sweepers, sweepers...)
}

func Run() {
	startup()
	shutdown()
}

func GetAppName() string {
	return appInstance.name
}

func GetAppVersion() string {
	return appInstance.version
}

func GetAppDescription() string {
	return appInstance.description
}

func GetAppAuthor() string {
	return appInstance.author
}

func GetAppChangeLog() string {
	return appInstance.changeLog
}

func startup() {
	logger.Info().Msgf("[%s] is starting", GetAppName())

	for _, f := range appInstance.starters {
		switch v := f.(type) {
		case func():
			v()
		case func() error:
			err := v()
			if err != nil {
				logger.Fatal().Err(err).Send()
			}
		default:
			err := fmt.Errorf("invalid starter: %s", reflect.TypeOf(f))
			logger.Fatal().Err(err).Send()
		}
	}

	logger.Info().Msgf("successfully start [%s]!", GetAppName())
}

func shutdown() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGSTOP, syscall.SIGINT, syscall.SIGTERM)

	<-ctx.Done()
	stop()

	logger.Info().Msgf("[%s] is sweeping", GetAppName())

	for _, f := range appInstance.sweepers {
		switch v := f.(type) {
		case func():
			v()
		case func() error:
			err := v()
			logger.Error().Err(err).Msg("app sweep error") // don't panic, go on sweeping
		default:
			err := fmt.Errorf("invalid sweeper: %s", reflect.TypeOf(f))
			logger.Fatal().Err(err).Send()
		}
	}

	fmt.Printf("\n===== press Ctrl+C again to force quit =====\n\n")
	logger.Info().Msgf("successfully stop [%s]!", GetAppName())
}
