package logger

import (
	"strings"
	"time"

	"gorm.io/gorm/logger"
)

type LoggerConfig struct {
	LogMode                 string `yaml:"logMode" json:"logMode"`
	IgnoreErrRecordNotFound *bool  `yaml:"ignoreErrRecordNotFound" json:"ignoreErrRecordNotFound"`
	SlowThresholdMS         int    `yaml:"slowThresholdMS" json:"slowThresholdMS"`
	// zap log config
	// Encoding     string   `json:"encoding" yaml:"encoding"`
	// OutputPaths  []string `json:"outputPaths" yaml:"outputPaths"`
	// TimeFormat   string   `yaml:"timeFormat" json:"timeFormat"`
	// DurationUnit string   `yaml:"durationUnit" json:"durationUnit"`
}

func getDefaultLoggerConfig() *LoggerConfig {
	var ignored bool = false
	return &LoggerConfig{
		LogMode:                 "info",
		IgnoreErrRecordNotFound: &ignored,
		SlowThresholdMS:         200,
		// Encoding:                "json",
		// TimeFormat:              "rfc3339utc",
	}
}

func getLogLevel(logMod string) logger.LogLevel {
	switch strings.ToLower(logMod) {
	case "trace":
		return logger.Silent - 1
	case "info":
		return logger.Info
	case "silent":
		return logger.Silent
	case "warn":
		return logger.Warn
	case "error":
		return logger.Error
	default:
		return logger.Info
	}
}

func isIgnoreErrRecordNotFound(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func getSlowThreshold(n int) time.Duration {
	if n <= 0 {
		n = 200
	}
	return time.Duration(n) * time.Millisecond
}
