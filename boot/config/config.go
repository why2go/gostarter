package config

import (
	"github.com/rs/zerolog/log"
	"github.com/why2go/gostarter/config"
)

var (
	_AppConfig = &AppConf{}
)

func init() {
	err := config.RegisterConfig(AppConf{})
	if err != nil {
		log.Fatal().Err(err).Str("ltag", "boot").Msg("register app config failed")
		return
	}
}

type AppConf struct {
	AppName     string `yaml:"name" json:"name"`
	Author      string `yaml:"author" json:"author"`
	Version     string `yaml:"version" json:"version"`
	ChangeLog   string `yaml:"changeLog" json:"changeLog"`
	Description string `yaml:"description" json:"description"`
}

func (AppConf) Prefix() string {
	return "app"
}

func GetAppConfig() (*AppConf, error) {
	err := config.GetConfig(_AppConfig)
	if err != nil {
		return nil, err
	}
	return _AppConfig, nil
}
