package config

import (
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// NamespaceId          string                   // the namespaceId of Nacos.When namespace is public, fill in the blank string here.
// AppName              string                   // the appName
// AppKey               string                   // the client identity information
// Endpoint             string                   // the endpoint for get Nacos server addresses
// RegionId             string                   // the regionId for kms
// AccessKey            string                   // the AccessKey for kms
// SecretKey            string                   // the SecretKey for kms

const (
	// nacos服务器地址
	NACOS_SERVER_ADDR = "NACOS_SERVER_ADDR"
	// nacos服务器连接端口
	NACOS_SERVER_PORT = "NACOS_SERVER_PORT"
	// 需要设置的nacos配置相关的环境变量
	NACOS_NAMESPACE_ID = "NACOS_NAMESPACE_ID"
	NACOS_APP_NAME     = "NACOS_APP_NAME"
	NACOS_APP_KEY      = "NACOS_APP_KEY"
	NACOS_ENDPOINT     = "NACOS_ENDPOINT"
	NACOS_REGION_ID    = "NACOS_REGION_ID"
	NACOS_ACCESS_KEY   = "NACOS_ACCESS_KEY"
	NACOS_SECRET_KEY   = "NACOS_SECRET_KEY"
	NACOS_OPEN_KMS     = "NACOS_OPEN_KMS"
	NACOS_USERNAMES    = "NACOS_USERNAME"
	NACOS_PASSWORD     = "NACOS_PASSWORD"
)

type nacosConfigLoader struct {
	configClient config_client.IConfigClient
}

func newNacosConfigLoader() (configLoader, error) {

	return nil, nil
}

func (loader *nacosConfigLoader) GetConfig(inf Configurable) error {
	configClient, err := clients.NewConfigClient(vo.NacosClientParam{
		ServerConfigs: []constant.ServerConfig{},
		ClientConfig:  &constant.ClientConfig{},
	})

	configClient.GetConfig(vo.ConfigParam{})

	configClient.ListenConfig(vo.ConfigParam{})

	//

	return nil
}
