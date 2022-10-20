//go:build ignore
// +build ignore

package logger

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type zapLogger struct {
	zlogger                   *zap.SugaredLogger
	LogLevel                  logger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
}

func NewZapLogger(cfg *LoggerConfig) *zapLogger {
	if cfg == nil {
		cfg = getDefaultLoggerConfig()
	}

	zapConfig := zap.NewProductionConfig()
	zapConfig.DisableCaller = true
	// encoding
	switch cfg.Encoding {
	case "json":
		zapConfig.Encoding = "json"
	case "console":
		zapConfig.Encoding = "console"
	default:
		zapConfig.Encoding = "json"
	}
	if len(cfg.OutputPaths) != 0 {
		zapConfig.OutputPaths = cfg.OutputPaths
	}

	switch strings.ToLower(cfg.TimeFormat) {
	case "rfc3339":
		zapConfig.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	case "rfc3339utc":
		zapConfig.EncoderConfig.EncodeTime = RFC3339UTCTimeEncoder
	case "rfc3339nano":
		zapConfig.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	case "rfc3339nanoutc":
		zapConfig.EncoderConfig.EncodeTime = RFC3339NanoUTCTimeEncoder
	case "epoch":
		zapConfig.EncoderConfig.EncodeTime = zapcore.EpochTimeEncoder
	case "epochmillis":
		zapConfig.EncoderConfig.EncodeTime = zapcore.EpochMillisTimeEncoder
	case "epochnanos":
		zapConfig.EncoderConfig.EncodeTime = zapcore.EpochNanosTimeEncoder
	case "iso8601":
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	default:
		zapConfig.EncoderConfig.EncodeTime = RFC3339UTCTimeEncoder
	}

	switch strings.ToLower(cfg.DurationUnit) {
	case "nanos":
		zapConfig.EncoderConfig.EncodeDuration = zapcore.NanosDurationEncoder
	case "millis":
		zapConfig.EncoderConfig.EncodeDuration = MillisDurationEncoder
	case "seconds":
		zapConfig.EncoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	case "string":
		zapConfig.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	default:
		zapConfig.EncoderConfig.EncodeDuration = MillisDurationEncoder
	}

	l, err := zapConfig.Build()
	if err != nil {
		log.Fatal("gorm starter init zap log failed", err)
		return nil
	}

	zlogger := &zapLogger{
		zlogger:                   l.Sugar(),
		LogLevel:                  getLogLevel(cfg.LogMode),
		IgnoreRecordNotFoundError: isIgnoreErrRecordNotFound(cfg.IgnoreErrRecordNotFound),
		SlowThreshold:             getSlowThreshold(cfg.SlowThresholdMS),
	}
	return zlogger
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

// LogMode log mode
func (l *zapLogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info print info
func (l zapLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.zlogger.Infof(msg, data...)
	}
}

// Warn print warn messages
func (l zapLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.zlogger.Warnf(msg, data...)
	}
}

// Error print error messages
func (l zapLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.zlogger.Errorf(msg, data...)
	}
}

// Trace print sql message
func (l zapLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= logger.Error && (!l.IgnoreRecordNotFoundError || !errors.Is(err, gorm.ErrRecordNotFound)):
		sql, rows := fc()
		l.zlogger.Errorw("gormTrace",
			zap.Error(err),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold && l.LogLevel >= logger.Warn:
		sql, rows := fc()
		l.zlogger.Warnw("gormTrace",
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	case l.LogLevel == logger.Info:
		sql, rows := fc()
		l.zlogger.Infow("gormTrace",
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	}
}
