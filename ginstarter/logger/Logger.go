package logger

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var (
	logger = log.With().Str("ltag", "ginLogger").Logger()
)

const (
	RequestIdKey       = "request-id"
	RequestIdHeaderKey = "x-request-id"
)

type LoggerConfig struct {
	SkipPaths []string
}

// LoggerWithConfig instance a Logger middleware with config.
func LoggerWithConfig(conf LoggerConfig) gin.HandlerFunc {
	notlogged := conf.SkipPaths

	var skip map[string]struct{}

	if length := len(notlogged); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range notlogged {
			skip[path] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		requestId := genRequestId()

		_, skipped := skip[path]

		c.Set(RequestIdKey, requestId)
		c.Header(RequestIdHeaderKey, requestId)

		// Log only if path is not being skipped
		if !skipped {
			if len(raw) != 0 {
				path = path + "?" + raw
			}
			e := logger.Log().
				Str("peer", c.ClientIP()).
				Str("method", c.Request.Method).
				Str("path", path).
				Str("requestId", requestId)
			e.Send()
		}

		// Process request
		c.Next()

		// Log only if path is not being skipped
		if !skipped {
			if len(raw) != 0 {
				path = path + "?" + raw
			}
			e := logger.Log().
				Str("peer", c.ClientIP()).
				Str("method", c.Request.Method).
				Str("path", path).
				Str("requestId", requestId).
				Int("statusCode", c.Writer.Status()).
				TimeDiff("latency", time.Now(), start).
				Int("bodySize", c.Writer.Size())
			if len(c.Errors.ByType(gin.ErrorTypePrivate).String()) != 0 {
				e = e.Str("errMsg", c.Errors.ByType(gin.ErrorTypePrivate).String())
			}
			e.Send()
		}
	}
}

func genRequestId() string {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		logger.Fatal().Err(err).Msg("generate request id failed")
	}
	return hex.EncodeToString(buf)
}
