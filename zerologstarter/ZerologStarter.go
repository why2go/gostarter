package zerologstarter

import (
	"time"

	config "github.com/why2go/gostarter/config"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	var err error
	cfg := &zerologConf{}
	err = config.GetConfig(cfg)
	if err != nil {
		if err == config.ErrCfgItemNotFound {
			cfg.GlobalLevel = zerolog.LevelInfoValue
		} else {
			log.Fatal().Err(err).Msg("init zerolog failed")
			return
		}
	}
	// 设置GolbalLevel
	level, err := zerolog.ParseLevel(cfg.GlobalLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = time.RFC3339
}

type zerologConf struct {
	GlobalLevel string `yaml:"globalLevel" json:"globalLevel"`
	// 可以预定义一些hooks，比如告警hook等
	// Hooks []string `yaml:"hooks" json:"hooks"`
}

func (cfg *zerologConf) GetConfigName() string {
	return "zerolog"
}
