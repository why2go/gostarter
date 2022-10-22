package ginstarter

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/why2go/gostarter/config"
	ginlogger "github.com/why2go/gostarter/ginstarter/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var (
	DefaultRouter *gin.Engine
	logger        = log.With().Str("ltag", "ginStarter").Logger()
	httpServer    *ginServer

	defaultListenPort      = uint16(8080)
	defaultShutdownLatency = 5 * time.Minute
)

func init() {
	cfg := &ginConf{}
	err := config.GetConfig(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("load gin conf failed")
		return
	}
	router := newGinRouter(cfg)
	DefaultRouter = router
	gs := newGinServer(router, cfg.Host, cfg.Port)
	httpServer = gs
}

type ginServer struct {
	host   string
	port   string
	server *http.Server
}

type ginConf struct {
	Host   string      `yaml:"host" json:"host"`
	Port   uint16      `yaml:"port" json:"port"`
	Mode   string      `yaml:"mode" json:"mode"`
	Cors   *corsConf   `yaml:"cors" json:"cors"`
	Logger *loggerConf `yaml:"logger" json:"logger"`
}

type corsConf struct {
	Origins []string `yaml:"origins" json:"origins"`
	Methods []string `yaml:"methods" json:"methods"`
	Headers []string `yaml:"headers" json:"headers"`
}

type loggerConf struct {
	SkipPaths []string `yaml:"skipPaths" json:"skipPaths"`
}

func (cfg *ginConf) GetConfigName() string {
	return "gin"
}

func newGinRouter(cfg *ginConf) *gin.Engine {
	setGinMode(cfg.Mode)
	e := gin.New()
	e.Use(gin.Recovery())
	setGinCors(e, cfg.Cors)
	setGinLogger(e, cfg.Logger)
	return e
}

func setGinMode(mode string) {
	switch strings.ToLower(mode) {
	case gin.DebugMode:
		mode = gin.DebugMode
	case gin.TestMode:
		mode = gin.TestMode
	case gin.ReleaseMode:
		mode = gin.ReleaseMode
	default:
		mode = gin.DebugMode
	}
	gin.SetMode(mode)
}

func setGinCors(router *gin.Engine, corsCfg *corsConf) {
	if corsCfg != nil {
		cfg := cors.Config{}
		if corsCfg.Origins != nil {
			cfg.AllowOrigins = corsCfg.Origins
		}
		if corsCfg.Methods != nil {
			cfg.AllowMethods = corsCfg.Methods
		}
		if corsCfg.Headers != nil {
			cfg.AllowHeaders = corsCfg.Headers
		}
		router.Use(cors.New(cfg))
	}
}

func setGinLogger(router *gin.Engine, loggerConf *loggerConf) {
	var skipPaths []string
	if loggerConf != nil {
		skipPaths = loggerConf.SkipPaths
	}
	router.Use(ginlogger.LoggerWithConfig(ginlogger.LoggerConfig{
		SkipPaths: skipPaths,
	}))
}

func newGinServer(mux http.Handler, host string, port uint16) *ginServer {
	svr := &ginServer{
		host: "",
		port: "8080",
		server: &http.Server{
			Handler: mux,
		},
	}
	if port == 0 {
		port = defaultListenPort
	}
	svr.server.Addr = fmt.Sprintf("%s:%d", host, port)
	return svr
}

func StartHttpServer() {
	go func() {
		if err := httpServer.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Err(err).Msgf("http server listen failed")
		}
	}()
}

func StopHttpServer() {
	logger.Info().Msg("shutting down http server...")
	ctx, cf := context.WithTimeout(context.Background(), defaultShutdownLatency)
	defer cf()
	err := httpServer.server.Shutdown(ctx)
	if err != nil {
		logger.Err(err).Msg("http server shutdown error")
		return
	}
	logger.Info().Msg("http server is closed")
}
