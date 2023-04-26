package config

import (
	"errors"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

var defaultLoader configLoader

func init() {
	cl, err := newConfigLoader()
	if err != nil {
		log.Fatal().Msg("can't create config loader")
		return
	}
	defaultLoader = cl
}

// 出现在配置文件中的配置项，应当实现此接口
type Configurable interface {
	GetConfigName() string
}

// 获取某项配置
type configLoader interface {
	GetConfig(inf Configurable) error
}

var (
	ErrCfgItemNotFound = errors.New("config item not found")
)

// get parsed config by name
func GetConfig(inf Configurable) error {
	return defaultLoader.GetConfig(inf)
}

const (
	CONF_PROFILE = "CONF_PROFILE" // environment variable decide which profile to load
)

// create config loader instance according to some environment variables
func newConfigLoader() (configLoader, error) {
	s, b := os.LookupEnv(CONF_PROFILE)
	if !b {
		log.Warn().Msgf("environment variable \"CONF_PROFILE\" not set, app.yaml will be used")
	}
	deployEnv := strings.ToLower(strings.TrimSpace(s))
	return newLocalConfigLoader(deployEnv)
}
