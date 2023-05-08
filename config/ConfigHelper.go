package config

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

// 出现在配置文件中的每项配置，都要实现此接口
// ConfigName()方法，返回配置项的前缀，为保证唯一性，建议使用域名+路径组合
// ConfigName()方法的返回值要满足如下正则约束：`^([A-Za-z0-9._~-]+\/)*[A-Za-z0-9._~-]+$`
// ConfigName()方法不应该返回空字符串，且字符长度不超过512
type Configurable interface {
	ConfigName() string
}

// 获取一项配置内容，c必须是一个指针
func GetConfig(c Configurable) error {
	return defaultHelper.getConfig(c)
}

var defaultHelper *configHelper

func init() {
	ch, err := newConfigHelper()
	if err != nil {
		log.Fatal().Err(err).Msg("load config file failed")
		return
	}
	defaultHelper = ch
}

type configHelper struct {
	configNameRegexp   *regexp.Regexp         // configName应该满足的命名规则
	rawConfigData      []byte                 // 加载的配置文件内容
	rawConfigEntries   map[string]interface{} // 通过configName来列出各个配置内容条目
	fileFormat         *fileFormat            // 配置文件格式
	cachedParsedConfig map[string]interface{} // 用于快速检索已经被解析的配置项
}

func newConfigHelper() (*configHelper, error) {
	l, err := newLocalConfigLoader()
	if err != nil {
		return nil, err
	}
	rawConfig, fileFormat, err := l.load()
	if err != nil {
		return nil, err
	}
	entries := make(map[string]interface{})
	err = fileFormat.parser.Unmarshal(rawConfig, &entries)
	if err != nil {
		return nil, err
	}
	helper := &configHelper{
		configNameRegexp:   regexp.MustCompile(`^([A-Za-z0-9._~-]+\/)*[A-Za-z0-9._~-]+$`),
		rawConfigData:      rawConfig,
		rawConfigEntries:   entries,
		fileFormat:         fileFormat,
		cachedParsedConfig: make(map[string]interface{}),
	}
	return helper, nil
}

var (
	ErrMalformedConfigName = errors.New("malformed config name")
	ErrNoConfigItemFound   = errors.New("no config item found")
)

// 传入一个指针，修改这个指针的值
func (helper *configHelper) getConfig(c Configurable) error {
	configName := strings.TrimSpace(c.ConfigName())
	if configName == "" || len(configName) > 512 ||
		!helper.configNameRegexp.MatchString(configName) {
		return ErrMalformedConfigName
	}
	rv := reflect.ValueOf(c)
	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf("non-pointer param, type: %s", reflect.TypeOf(c).String())
	}
	ctyp := reflect.TypeOf(c)
	if parsedConf, exists := helper.cachedParsedConfig[configName]; exists {
		ptyp := reflect.TypeOf(parsedConf)
		if ctyp != ptyp {
			return fmt.Errorf(`conflict config name, their type are "%s" and "%s"`, ctyp, ptyp)
		} else {
			rv.Elem().Set(reflect.ValueOf(parsedConf).Elem())
		}
	} else {
		if inf, exists := helper.rawConfigEntries[configName]; exists {
			parser := helper.fileFormat.parser
			b, _ := parser.Marshal(inf)
			v := reflect.New(ctyp.Elem())
			err := parser.Unmarshal(b, v.Interface())
			if err != nil {
				return err
			}
			helper.cachedParsedConfig[configName] = v.Interface()
			rv.Elem().Set(v.Elem())
		} else {
			return ErrNoConfigItemFound
		}
	}
	return nil
}
