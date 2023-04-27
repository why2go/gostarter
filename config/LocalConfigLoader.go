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

// load configs from "./resource/cfg/app-[confProfile].[yml|yaml|json]"

const (
	// 当使用本地配置时，CONF_PROFILE 用来确定使用哪个环境的配置文件，如dev, qa, test, prod等
	// 寻找路径形式为："./resource/cfg/app-[${CONF_PROFILE}].[yml|yaml|json]"
	// 如果没有设定 CONF_PROFILE，则默认使用 "./resource/cfg/app.[yml|yaml|json]"
	CONF_PROFILE = "CONF_PROFILE"
)

type localConfigLoader struct {
	cfgFileDir     string
	fileNamePrefix string
	confProfile    string
	knownFormats   []string
	rawCfgEntries  map[string]interface{} // fixme: not efficient
	isLoaded       bool
	parser         configParser
	envVarsRegex   *regexp.Regexp
}

func newLocalConfigLoader() (configLoader, error) {
	var confProfile string
	s, b := os.LookupEnv(CONF_PROFILE)
	if !b {
		log.Warn().Msgf("environment variable \"CONF_PROFILE\" not set, app.yaml will be used")
	}
	confProfile = strings.ToLower(strings.TrimSpace(s))
	reg, _ := regexp.Compile(`\$\{\s*[a-zA-Z_][a-zA-Z0-9_]*\s*\}`)
	return &localConfigLoader{
		cfgFileDir:     "./resource/cfg/",
		fileNamePrefix: "app",
		confProfile:    confProfile,
		knownFormats:   []string{"yml", "yaml", "json"},
		rawCfgEntries:  make(map[string]interface{}),
		isLoaded:       false,
		envVarsRegex:   reg,
	}, nil
}

func (loader *localConfigLoader) GetConfig(inf Configurable) error {
	if !loader.isLoaded {
		err := loader.loadConfig()
		if err != nil {
			return err
		}
	}
	cfgName := inf.GetConfigName()

	i, ok := loader.rawCfgEntries[cfgName]
	if !ok {
		return ErrCfgItemNotFound
	}

	out, _ := loader.parser.Marshal(i)

	return loader.parser.Unmarshal(out, inf)
}

func (loader *localConfigLoader) loadConfig() error {
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
	var parser configParser
	for i := range formats {
		filepath := pathPrefix + formats[i]
		rawCfg, err = os.ReadFile(filepath)
		if err != nil {
			log.Trace().Err(err).Msgf("can't open config file: %s", filepath)
		} else {
			switch formats[i] {
			case "yml", "yaml":
				parser = &yamlParser{}
			case "json":
				parser = &jsonParser{}
			}
			break
		}
	}
	if err != nil {
		err = errors.New("can't open any config file")
		return err
	}
	// replace all env vars expression with their values
	replacedRawCfg := loader.envVarsRegex.ReplaceAllFunc(rawCfg, func(b []byte) []byte {
		val, set := os.LookupEnv(string(bytes.TrimSpace(b[2 : len(b)-1])))
		if !set {
			log.Warn().Msgf("env variable \"%s\" not set", string(bytes.TrimSpace(b[2:len(b)-1])))
		}
		return bytes.TrimSpace([]byte(val))
	})
	rawCfgEntries := make(map[string]interface{})
	err = parser.Unmarshal(replacedRawCfg, rawCfgEntries)
	if err != nil {
		return err
	}

	log.Trace().Interface("raw config entries", rawCfgEntries).Msg("")

	loader.rawCfgEntries = rawCfgEntries
	loader.isLoaded = true
	loader.parser = parser
	return nil
}
