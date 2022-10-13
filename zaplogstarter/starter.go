package zaplogstarter

import (
	"log"
	"strings"

	"github.com/why2go/gostarter/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// todo: 考虑直接使用zap的类型

var ZapConfig *zapConfig

func init() {
	cfg := &zapConfig{Config: zap.NewProductionConfig()}
	err := config.GetConfig(cfg)
	if err != nil && err != config.ErrCfgItemNotFound {
		log.Fatal("load zaplog config failed")
		return
	}
	if len(cfg.TimeFormat) != 0 {
		switch strings.ToLower(cfg.TimeFormat) {
		case "rfc3339":
			cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		case "rfc3339nano":
			cfg.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
		case "epoch":
			cfg.EncoderConfig.EncodeTime = zapcore.EpochTimeEncoder
		case "epochmillis":
			cfg.EncoderConfig.EncodeTime = zapcore.EpochMillisTimeEncoder
		case "epochnanos":
			cfg.EncoderConfig.EncodeTime = zapcore.EpochNanosTimeEncoder
		case "iso8601":
			cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		default:
			cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		}
	}
	if len(cfg.DurationUnit) != 0 {
		switch strings.ToLower(cfg.DurationUnit) {
		case "nanos":
			cfg.EncoderConfig.EncodeDuration = zapcore.NanosDurationEncoder
		case "seconds":
			cfg.EncoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
		case "string":
			cfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
		default:
			cfg.EncoderConfig.EncodeDuration = zapcore.NanosDurationEncoder
		}
	}
	ZapConfig = cfg
}

type zapConfig struct {
	zap.Config
	TimeFormat   string `yaml:"timeFormat" json:"timeFormat"`
	DurationUnit string `yaml:"durationUnit" json:"durationUnit"`
}

func (cfg *zapConfig) GetConfigName() string {
	return "zaplog"
}
