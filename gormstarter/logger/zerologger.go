package logger

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type zeroLogger struct {
	logger                    zerolog.Logger
	LogLevel                  logger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
}

func NewZeroLogger(cfg *LoggerConfig) *zeroLogger {
	if cfg == nil {
		cfg = getDefaultLoggerConfig()
	}
	return &zeroLogger{
		logger:                    log.With().Str("ltag", "gormStarter").Logger(),
		LogLevel:                  getLogLevel(cfg.LogMode),
		SlowThreshold:             getSlowThreshold(cfg.SlowThresholdMS),
		IgnoreRecordNotFoundError: isIgnoreErrRecordNotFound(cfg.IgnoreErrRecordNotFound),
	}
}

// LogMode log mode
func (l *zeroLogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info print info
func (l zeroLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.logger.Log().Msgf(msg+", %s", data...)
	}
}

// Warn print warn messages
func (l zeroLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.logger.Log().Msgf(msg+", %s", data...)
	}
}

// Error print error messages
func (l zeroLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.logger.Log().Msgf(msg+", %s", data...)
	}
}

// Trace print sql message
func (l zeroLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= logger.Error && (!l.IgnoreRecordNotFoundError || !errors.Is(err, gorm.ErrRecordNotFound)):
		sql, rows := fc()
		l.logger.Log().
			Err(err).
			Dur("elapsed", elapsed).
			Int64("rows", rows).
			Str("sql", sql).
			Send()
	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold && l.LogLevel >= logger.Warn:
		sql, rows := fc()
		l.logger.Log().
			Dur("elapsed", elapsed).
			Int64("rows", rows).
			Str("sql", sql).
			Send()
	case l.LogLevel == logger.Info:
		sql, rows := fc()
		l.logger.Log().
			Dur("elapsed", elapsed).
			Int64("rows", rows).
			Str("sql", sql).
			Send()
	}
}
