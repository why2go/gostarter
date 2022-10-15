package zaplogstarter

import (
	"log"
	"strings"
	"time"

	"github.com/why2go/gostarter/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var ZapConfig *zap.Config

func init() {
	cfg := &zapConfig{}
	err := config.GetConfig(cfg)
	if err != nil && err != config.ErrCfgItemNotFound {
		log.Fatal("load zaplog config failed")
		return
	}
	ZapConfig = cfg.buildZapProductionConfig()
}

type zapConfig struct {
	Level             string              `json:"level" yaml:"level"`
	DisableCaller     bool                `json:"disableCaller" yaml:"disableCaller"`
	DisableStacktrace bool                `json:"disableStacktrace" yaml:"disableStacktrace"`
	Sampling          *zap.SamplingConfig `json:"sampling" yaml:"sampling"`
	Encoding          string              `json:"encoding" yaml:"encoding"`
	EncoderConfig     struct {
		MessageKey     string `json:"messageKey" yaml:"messageKey"`
		LevelKey       string `json:"levelKey" yaml:"levelKey"`
		TimeKey        string `json:"timeKey" yaml:"timeKey"`
		NameKey        string `json:"nameKey" yaml:"nameKey"`
		CallerKey      string `json:"callerKey" yaml:"callerKey"`
		StacktraceKey  string `json:"stacktraceKey" yaml:"stacktraceKey"`
		LineEnding     string `json:"lineEnding" yaml:"lineEnding"`
		EncodeTime     string `json:"timeEncoder" yaml:"timeEncoder"`
		EncodeDuration string `json:"durationEncoder" yaml:"durationEncoder"`
		EncodeCaller   string `json:"callerEncoder" yaml:"callerEncoder"`
	} `json:"encoderConfig" yaml:"encoderConfig"`
	OutputPaths      []string `json:"outputPaths" yaml:"outputPaths"`
	ErrorOutputPaths []string `json:"errorOutputPaths" yaml:"errorOutputPaths"`
	// customize
	TimeFormat   string `yaml:"timeFormat" json:"timeFormat"`
	DurationUnit string `yaml:"durationUnit" json:"durationUnit"`
}

func (cfg *zapConfig) GetConfigName() string {
	return "zaplog"
}

func (cfg *zapConfig) buildZapProductionConfig() *zap.Config {
	pcfg := zap.NewProductionConfig()
	// level
	level := &zap.AtomicLevel{}
	level.UnmarshalText([]byte(cfg.Level))
	pcfg.Level = *level

	pcfg.DisableCaller = cfg.DisableCaller
	pcfg.DisableStacktrace = cfg.DisableStacktrace

	if cfg.Sampling != nil {
		pcfg.Sampling = cfg.Sampling
	}
	// encoding
	switch cfg.Encoding {
	case "json":
		pcfg.Encoding = "json"
	case "console":
		pcfg.Encoding = "console"
	default:
		pcfg.Encoding = "json"
	}

	if len(cfg.TimeFormat) != 0 {
		switch strings.ToLower(cfg.TimeFormat) {
		case "rfc3339":
			pcfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		case "rfc3339utc":
			pcfg.EncoderConfig.EncodeTime = RFC3339UTCTimeEncoder
		case "rfc3339nano":
			pcfg.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
		case "rfc3339nanoutc":
			pcfg.EncoderConfig.EncodeTime = RFC3339NanoUTCTimeEncoder
		case "epoch":
			pcfg.EncoderConfig.EncodeTime = zapcore.EpochTimeEncoder
		case "epochmillis":
			pcfg.EncoderConfig.EncodeTime = zapcore.EpochMillisTimeEncoder
		case "epochnanos":
			pcfg.EncoderConfig.EncodeTime = zapcore.EpochNanosTimeEncoder
		case "iso8601":
			pcfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		default:
			pcfg.EncoderConfig.EncodeTime = RFC3339UTCTimeEncoder
		}
	}

	if len(cfg.DurationUnit) != 0 {
		switch strings.ToLower(cfg.DurationUnit) {
		case "nanos":
			pcfg.EncoderConfig.EncodeDuration = zapcore.NanosDurationEncoder
		case "millis":
			pcfg.EncoderConfig.EncodeDuration = MillisDurationEncoder
		case "seconds":
			pcfg.EncoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
		case "string":
			pcfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
		default:
			pcfg.EncoderConfig.EncodeDuration = MillisDurationEncoder
		}
	}

	var callerEncoder zapcore.CallerEncoder
	callerEncoder.UnmarshalText([]byte(cfg.EncoderConfig.EncodeCaller))
	pcfg.EncoderConfig.EncodeCaller = callerEncoder

	if len(cfg.OutputPaths) != 0 {
		pcfg.OutputPaths = cfg.OutputPaths
	}
	if len(cfg.ErrorOutputPaths) != 0 {
		pcfg.ErrorOutputPaths = cfg.ErrorOutputPaths
	}

	return &pcfg
}

func RFC3339UTCTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.UTC().Format(time.RFC3339))
}

func RFC3339NanoUTCTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.UTC().Format(time.RFC3339Nano))
}

func MillisDurationEncoder(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendFloat64(float64(d) / float64(time.Millisecond))
}
