package config

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

// 出现在配置文件中的每项配置，都要实现此接口
// Prefix()方法，返回配置项的前缀，为保证唯一性，建议使用域名+路径组合
// Prefix()方法的返回值要满足如下正则约束：`^([A-Za-z0-9._~-]+\/)*[A-Za-z0-9._~-]+$`
// Prefix()方法不应该返回空字符串，且字符长度不超过512
type Configurable interface {
	Prefix() string
}

// 注册配置项，只有在注册成功后，才能通过GetConfig方法获取到配置内容
func RegisterConfig(c Configurable) error {
	return defaultHelper.registerConfig(c)
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
	configs      map[string]Configurable
	prefixRegexp *regexp.Regexp // prefix应该满足的命名规则
	rawConfig    []byte         // 加载的配置文件内容
	fileFormat   *fileFormat    // 配置文件格式
	mu           sync.Mutex
	configType   reflect.Type         // 将所有注册的配置项利用反射构造出一个结构体，以用来解析配置内容
	needParse    bool                 // 是否需要重新解析配置文件
	parsedValue  reflect.Value        // 存储已经解析的配置内容，是指向struct的指针
	fieldMap     map[reflect.Type]int // 用于快速检索配置项是哪个字段
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
	helper := &configHelper{
		configs:      make(map[string]Configurable),
		prefixRegexp: regexp.MustCompile(`^([A-Za-z0-9._~-]+\/)*[A-Za-z0-9._~-]+$`),
		rawConfig:    rawConfig,
		fileFormat:   fileFormat,
		needParse:    true,
	}
	return helper, nil
}

const (
	maxPrefixLength = 512
)

var (
	errMalformedPrefix    = errors.New(`malformed prefix, no longer than 512, and must match "^([A-Za-z0-9._~-]+\/)*[A-Za-z0-9._~-]+$"`)
	errUnregisteredConfig = errors.New(`unregistered config`)
)

func (helper *configHelper) registerConfig(c Configurable) error {
	helper.mu.Lock()
	defer helper.mu.Unlock()
	prefix := strings.TrimSpace(c.Prefix())
	if len(prefix) == 0 ||
		len(prefix) > maxPrefixLength ||
		!helper.prefixRegexp.MatchString(prefix) {
		return errMalformedPrefix
	}
	if preCfg, exists := helper.configs[prefix]; exists {
		return fmt.Errorf("duplicated config prefix: %s, previous config type is: %s",
			prefix, reflect.TypeOf(preCfg))
	}
	helper.configs[prefix] = c
	helper.needParse = true
	return nil
}

// 传入一个指针，修改这个指针的值
func (helper *configHelper) getConfig(c Configurable) error {
	rv := reflect.ValueOf(c)
	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf("non-pointer param, type: %s", reflect.TypeOf(c).String())
	}
	helper.mu.Lock()
	defer helper.mu.Unlock()
	prefix := strings.TrimSpace(c.Prefix())
	if _, exists := helper.configs[prefix]; !exists {
		return errUnregisteredConfig
	}
	if helper.needParse {
		err := helper.parseRawConfig()
		if err != nil {
			return err
		}
	}
	ctyp := reflect.TypeOf(c)
	idx := helper.fieldMap[ctyp]
	v := helper.parsedValue.Elem().Field(idx) // 指针
	rv.Elem().Set(v.Elem())
	return nil
}

func (helper *configHelper) parseRawConfig() error {
	log.Info().Msg("parse raw config")
	var err error
	err = helper.newStructType()
	if err != nil {
		return err
	}
	value := reflect.New(helper.configType)
	inf := value.Interface()
	parser := helper.fileFormat.parser
	err = parser.Unmarshal(helper.rawConfig, inf)
	if err != nil {
		return err
	}
	helper.parsedValue = value
	helper.needParse = false
	return nil
}

func (helper *configHelper) newStructType() error {
	var fields []reflect.StructField
	var idx int
	fieldMap := make(map[reflect.Type]int)
	for _, c := range helper.configs {
		prefix := strings.TrimSpace(c.Prefix())
		ctyp := reflect.TypeOf(c)
		if ctyp.Kind() != reflect.Pointer {
			ctyp = reflect.PointerTo(ctyp)
		}
		if len(prefix) == 0 {
			return fmt.Errorf("type %s has empty prefix", ctyp.String())
		} else { // 如果有配置前缀
			field := reflect.StructField{
				Name: genFieldName(idx),
				Type: ctyp,
				Tag:  genSupportedTags(prefix),
			}
			fields = append(fields, field)
			fieldMap[ctyp] = idx
			idx++
		}
	}
	helper.configType = reflect.StructOf(fields)
	helper.fieldMap = fieldMap
	return nil
}

func genSupportedTags(tagVal string) reflect.StructTag {
	var tag string
	for _, v := range allTagPrefixes {
		tag += fmt.Sprintf(` %s:"%s"`, v, tagVal)
	}
	return reflect.StructTag(tag[1:])
}

func genFieldName(idx int) string {
	return fmt.Sprintf("Field_%d", idx)
}
