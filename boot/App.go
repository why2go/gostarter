package boot

import (
	"context"

	"fmt"
	"os/signal"
	"reflect"
	"syscall"

	"github.com/why2go/gostarter/boot/config"
	_ "github.com/why2go/gostarter/zerologstarter"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// 规定应用的启动过程，包括配置加载，执行启动函数，执行清理函数

var (
	logger      zerolog.Logger
	appInstance *app
)

func init() {
	logger = log.With().Str("ltag", "boot").Logger()
	appConfig, err := config.GetAppConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("load app config failed")
		return
	}
	appInstance = newApp(appConfig)
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

func newApp(cfg *config.AppConf) *app {
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
