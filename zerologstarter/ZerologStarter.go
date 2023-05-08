package zerologstarter

import (
	"strings"
	"time"

	config "github.com/why2go/gostarter/config"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	var err error
	cfg := &zerologConf{}
	err = config.GetConfig(cfg)
	if err != nil {
		if err == config.ErrNoConfigItemFound {
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
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.TimestampFunc = time.Now().UTC
	zerolog.MessageFieldName = "msg"

	switch strings.ToLower(cfg.DurationFieldUnit) {
	case "millis", "ms", "millisecond":
		zerolog.DurationFieldUnit = time.Millisecond
	case "nanos", "ns", "nanosecond":
		zerolog.DurationFieldUnit = time.Nanosecond
	case "micros", "us", "microsecond":
		zerolog.DurationFieldUnit = time.Microsecond
	}

	if cfg.EnableRotation {
		if cfg.Logger == nil {
			log.Fatal().Msgf("init zerolog failed, rotation config can't be nil")
			return
		}
		w := lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize, // megabytes
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,   //days
			Compress:   cfg.Compress, // disabled by default
		}
		log.Logger = zerolog.New(&w).With().Timestamp().Logger()
	}
}

type zerologConf struct {
	GlobalLevel        string `yaml:"globalLevel" json:"globalLevel"`
	DurationFieldUnit  string `yaml:"durationFieldUnit" json:"durationFieldUnit"`
	EnableRotation     bool   `yaml:"enableRotation" json:"enableRotation"`
	*lumberjack.Logger `yaml:"rotationConfig" json:"rotationConfig"`
}

func (cfg *zerologConf) ConfigName() string {
	return "zerolog"
}
