package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/rs/zerolog/log"
)

const (
	// nacos服务器地址
	NACOS_SERVER_ADDR = "NACOS_SERVER_ADDR"
	// nacos服务器连接端口
	NACOS_SERVER_PORT = "NACOS_SERVER_PORT"
	// 需要设置的nacos配置相关的环境变量
	NACOS_NAMESPACE_ID  = "NACOS_NAMESPACE_ID"
	NACOS_APP_NAME      = "NACOS_APP_NAME"
	NACOS_ENDPOINT      = "NACOS_ENDPOINT"
	NACOS_REGION_ID     = "NACOS_REGION_ID"
	NACOS_ACCESS_KEY    = "NACOS_ACCESS_KEY"
	NACOS_SECRET_KEY    = "NACOS_SECRET_KEY"
	NACOS_OPEN_KMS      = "NACOS_OPEN_KMS"
	NACOS_USERNAME      = "NACOS_USERNAME"
	NACOS_PASSWORD      = "NACOS_PASSWORD"
	NACOS_GROUP         = "NACOS_GROUP"
	NACOS_DATA_IDS      = "NACOS_DATA_IDS"
	NACOS_CONFIG_FORMAT = "NACOS_CONFIG_FORMAT"
)

type nacosConfigLoader struct {
	isLoaded     bool
	knownFormats []string
	configClient config_client.IConfigClient
	// dataId ==> configItemName ==> configDetail
	rawCfgEntries map[string]interface{}
	parser        configParser
	configFormat  string
	envVars       envVars
	rwlock        sync.RWMutex
	configItems   map[string][]string // dataId ==> array of config names
}

type envVars struct {
	nacosServerAddr   string
	nacosServerPort   uint16
	nacosNamespaceId  string
	nacosAppName      string
	nacosEndpoint     string
	nacosRegionId     string
	nacosOpenKms      bool
	nacosAccessKey    string
	nacosSecretKey    string
	nacosUsername     string
	nacosPassword     string
	nacosGroup        string
	nacosDataIds      []string
	nacosConfigFormat string
}

func newNacosConfigLoader() (configLoader, error) {
	loader := &nacosConfigLoader{
		knownFormats:  []string{"yml", "yaml", "json"},
		rawCfgEntries: make(map[string]interface{}),
		configItems:   make(map[string][]string),
	}
	evs, err := loader.lookupEnvVars()
	if err != nil {
		return nil, err
	}
	configClient, err := clients.NewConfigClient(vo.NacosClientParam{
		ServerConfigs: []constant.ServerConfig{
			{IpAddr: evs.nacosServerAddr, Port: uint64(evs.nacosServerPort)},
		},
		ClientConfig: constant.NewClientConfig(
			constant.WithNamespaceId(evs.nacosNamespaceId),
			constant.WithNotLoadCacheAtStart(true),
			constant.WithAccessKey(evs.nacosAccessKey),
			constant.WithSecretKey(evs.nacosSecretKey),
			constant.WithOpenKMS(evs.nacosOpenKms),
			constant.WithUsername(evs.nacosUsername),
			constant.WithPassword(evs.nacosPassword),
			constant.WithAppName(evs.nacosAppName),
			constant.WithRegionId(evs.nacosRegionId),
			constant.WithEndpoint(evs.nacosEndpoint),
		),
	})

	switch evs.nacosConfigFormat {
	case "yml", "yaml":
		loader.parser = &yamlParser{}
	case "json":
		loader.parser = &jsonParser{}
	default:
		return nil, errors.New("unsupported config file format")
	}

	loader.envVars = *evs
	loader.configClient = configClient

	loader.syncRemoteConfig(loader.envVars.nacosGroup, loader.envVars.nacosDataIds)

	for _, dataId := range loader.envVars.nacosDataIds {
		loader.configClient.ListenConfig(vo.ConfigParam{
			DataId:   dataId,
			Group:    loader.envVars.nacosGroup,
			OnChange: loader.onConfigChange,
		})
	}

	return loader, nil
}

func (loader *nacosConfigLoader) lookupEnvVars() (*envVars, error) {
	var exists bool
	var envVar string
	ans := &envVars{}
	envVar, exists = os.LookupEnv(NACOS_SERVER_ADDR)
	if exists {
		ans.nacosServerAddr = envVar
	} else {
		return nil, errors.New("no nacos server address found")
	}
	envVar, exists = os.LookupEnv(NACOS_SERVER_PORT)
	if exists {
		port, err := strconv.Atoi(envVar)
		if err != nil {
			return nil, err
		}
		ans.nacosServerPort = uint16(port)
	} else {
		return nil, errors.New("no nocos server port found")
	}
	ans.nacosNamespaceId, _ = os.LookupEnv(NACOS_NAMESPACE_ID)
	ans.nacosAppName, _ = os.LookupEnv(NACOS_APP_NAME)
	ans.nacosEndpoint, _ = os.LookupEnv(NACOS_ENDPOINT)
	ans.nacosAccessKey, _ = os.LookupEnv(NACOS_ACCESS_KEY)
	ans.nacosSecretKey, _ = os.LookupEnv(NACOS_SECRET_KEY)
	envVar, exists = os.LookupEnv(NACOS_OPEN_KMS)
	if exists {
		envVar = strings.ToLower(envVar)
		if envVar == "true" || envVar == "1" {
			ans.nacosOpenKms = true
		}
	}
	ans.nacosUsername, _ = os.LookupEnv(NACOS_USERNAME)
	ans.nacosPassword, _ = os.LookupEnv(NACOS_PASSWORD)
	envVar, exists = os.LookupEnv(NACOS_GROUP)
	if exists {
		ans.nacosGroup = envVar
	} else {
		return nil, errors.New("no nacos group found")
	}
	envVar, exists = os.LookupEnv(NACOS_DATA_IDS)
	if exists {
		ids := strings.Split(envVar, ",")
		for _, v := range ids {
			ans.nacosDataIds = append(ans.nacosDataIds, strings.TrimSpace(v))
		}
	} else {
		return nil, errors.New("no nacos data ids found")
	}

	format, exists := os.LookupEnv(NACOS_CONFIG_FORMAT)
	if exists {
		var found bool
		for _, v := range loader.knownFormats {
			if v == format {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unsupported config file format: %s", format)
		}
		ans.nacosConfigFormat = format
	} else {
		return nil, errors.New("no config file format found")
	}

	return ans, nil
}

func (loader *nacosConfigLoader) syncRemoteConfig(group string, dataIds []string) error {
	for _, dataId := range dataIds {
		configStr, err := loader.configClient.GetConfig(vo.ConfigParam{
			DataId: dataId,
			Group:  group,
		})
		if err != nil {
			return err
		}
		rawCfgEntries := make(map[string]interface{})
		err = loader.parser.Unmarshal([]byte(configStr), rawCfgEntries)
		if err != nil {
			return err
		}
		var cfgNames []string
		for k, v := range rawCfgEntries {
			loader.rawCfgEntries[k] = v
			cfgNames = append(cfgNames, k)
		}
		loader.configItems[dataId] = cfgNames
	}
	return nil
}

func (loader *nacosConfigLoader) GetConfig(inf Configurable) error {
	loader.rwlock.RLock()
	defer loader.rwlock.RUnlock()
	configName := inf.GetConfigName()
	loadedEntries := loader.rawCfgEntries
	if entry, ok := loadedEntries[configName]; ok {
		var err error
		b, err := loader.parser.Marshal(entry)
		if err != nil {
			return err
		}
		err = loader.parser.Unmarshal(b, inf)
		if err != nil {
			return err
		}
	} else {
		return ErrCfgItemNotFound
	}
	return nil
}

func (loader *nacosConfigLoader) onConfigChange(namespace, group, dataId, data string) {
	// namesapce用于区别不同的微服务项目集群
	// namespace是nacos默认权限管理的唯一资源粒度
	// 因此要考虑资源权限力度来组织应用的配置
	// group可以设置为不同的开发环境，比如dev，qa，test，prod等
	// dataId用于指定特定的应用，比如user-app-config, user-source-config
	loader.rwlock.Lock()
	defer loader.rwlock.Unlock()
	var err error
	rawCfgEntries := make(map[string]interface{})
	err = loader.parser.Unmarshal([]byte(data), rawCfgEntries)
	if err != nil {
		log.Error().Err(err).Str("ltag", "nacosListener").
			Msgf("unmarshal remote config failed")
		return
	}
	if cfgNames, ok := loader.configItems[dataId]; ok {
		for _, name := range cfgNames {
			delete(loader.configItems, name)
		}
	}
	var newNames []string
	for k, v := range rawCfgEntries {
		loader.rawCfgEntries[k] = v
		newNames = append(newNames, k)
	}
	loader.configItems[dataId] = newNames

	// 通知各个配置变更
	for _, name := range newNames {
		listenerMgr.handleChangeEvent(name)
	}

	log.Info().Interface("newData", rawCfgEntries).Msg("receive new data")
}
