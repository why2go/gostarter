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

type Configurable interface {
	GetConfigName() string
}

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
	// environment variable decide which profile to load
	CONF_PROFILE = "CONF_PROFILE"
)

// create config loader instance according to some environment variables
func newConfigLoader() (configLoader, error) {
	deployEnv := strings.ToLower(os.Getenv(CONF_PROFILE))
	return newLocalConfigLoader(deployEnv)
}
