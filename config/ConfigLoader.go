package config

import (
	"bytes"
	"errors"
	"os"
	"regexp"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// 加载本地配置文件，配合环境变量CONFIG_PROFILE的值，来确定从那个文件中读取配置
// 寻找路径形式为："./resource/app-[${CONFIG_PROFILE}].[yml|yaml|json]"
// 如果没有设定 CONFIG_PROFILE，则默认使用 "./resource/app.[yml|yaml|json]"
// 对于同一个CONFIG_PROFILE值，应当只存在一个对应的配置文件，如果存在不同后缀的配置文件，
// 例如，同时存在 ./resource/app.yml, ./resource/app.yaml, ./resource/app.json三个文件
// 则只会使用其中一个文件，这往往会产生令人疑惑的结果
// 配置文件中可以使用类似 ${ENV_VAR} 的形式来表示，使用环境变量获取此值，如果没有设置，则默认为空
// 具体匹配的正则表达式为： `\$\{\s*[a-zA-Z_][a-zA-Z0-9_]*\s*\}`

const (
	CONFIG_PROFILE = "CONFIG_PROFILE"
)

type ConfigLoader interface {
	load() ([]byte, *fileFormat, error)
}

type localConfigLoader struct {
	cfgFileDir     string
	fileNamePrefix string
	confProfile    string
	knownFormats   []*fileFormat
	envVarsRegex   *regexp.Regexp
}

func newLocalConfigLoader() (*localConfigLoader, error) {
	var confProfile string
	s, b := os.LookupEnv(CONFIG_PROFILE)
	if !b {
		log.Warn().Msgf("environment variable \"CONFIG_PROFILE\" not set, app.yaml or app.json will be used")
	}
	confProfile = strings.ToLower(strings.TrimSpace(s))
	reg, _ := regexp.Compile(`\$\{\s*[a-zA-Z_][a-zA-Z0-9_]*\s*\}`)
	return &localConfigLoader{
		cfgFileDir:     "./resource/",
		fileNamePrefix: "app",
		confProfile:    confProfile,
		knownFormats:   allSupportedFileFormats,
		envVarsRegex:   reg,
	}, nil
}

func (loader *localConfigLoader) load() ([]byte, *fileFormat, error) {
	oldGlobalLevel := zerolog.GlobalLevel()
	defer func() {
		zerolog.SetGlobalLevel(oldGlobalLevel)
	}()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	var rawCfg []byte
	var err error
	formats := loader.knownFormats
	var pathPrefix string
	if len(loader.confProfile) == 0 {
		pathPrefix = loader.cfgFileDir + loader.fileNamePrefix + "."
	} else {
		pathPrefix = loader.cfgFileDir + loader.fileNamePrefix + "-" + loader.confProfile + "."
	}

	var foundFormat *fileFormat = nil
out:
	for i := range formats {
		for _, suffix := range formats[i].fileSuffix {
			filepath := pathPrefix + suffix
			rawCfg, err = os.ReadFile(filepath)
			if err != nil {
				log.Trace().Err(err).Msgf("can't open config file: %s", filepath)
			} else {
				foundFormat = formats[i]
				break out
			}
		}
	}
	if foundFormat == nil {
		return nil, nil, errors.New("no config file found")
	}
	// replace all env vars expression with their values
	replacedRawCfg := loader.envVarsRegex.ReplaceAllFunc(rawCfg, func(b []byte) []byte {
		val, set := os.LookupEnv(string(bytes.TrimSpace(b[2 : len(b)-1])))
		if !set {
			log.Warn().Msgf("env variable \"%s\" not set", string(bytes.TrimSpace(b[2:len(b)-1])))
		}
		return bytes.TrimSpace([]byte(val))
	})
	return replacedRawCfg, foundFormat, nil
}
