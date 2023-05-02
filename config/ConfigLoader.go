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
	ErrUnsupportedConfigSource = errors.New("config source is not supported")
	ErrCfgItemNotFound         = errors.New("config item not found")
)

// get parsed config by name
func GetConfig(inf Configurable) error {
	return defaultLoader.GetConfig(inf)
}

const (
	// 配置文件来源，local表示使用本地配置，nacos表示使用Nacos远程配置，默认使用本地配置
	CONFIG_SOURCE = "CONFIG_SOURCE"
)

const (
	config_source_local = "local"
	config_source_nacos = "nacos"
)

// create config loader instance according to some environment variables
func newConfigLoader() (configLoader, error) {
	configSource := config_source_local
	configSourceVal, b := os.LookupEnv(CONFIG_SOURCE)
	if b {
		configSource = strings.ToLower(strings.TrimSpace(configSourceVal))
	}
	switch configSource {
	case "", config_source_local:
		return newLocalConfigLoader()
	case config_source_nacos:
		return newNacosConfigLoader()
	default:
		return nil, ErrUnsupportedConfigSource
	}
}
