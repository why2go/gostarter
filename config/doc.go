// 用于加载应用的配置。
//
// 支持的配置文件格式为yaml和json，相应支持的文件后缀为yml，yaml，json。
//
// 配置文件的查找路径为当前工作目录下的：resource/cfg/。
//
// 您可以为不同的开发环境配置不同的配置文件，配合环境变量 CONF_PROFILE 来指定特定场景使用的配置文件。
//
// 比如, CONF_PROFILE=prod, 则会使用 resource/cfg/app-prod.yaml 或者 resource/cfg/app-prod.json.
//
// 如果没有指定 CONF_PROFILE 环境变量，则会使用 resource/cfg/app.yaml 或者 resource/cfg/app.json。
//
// 可以使用如下方式快速为您的应用增加一项配置，下面的示例展示了如何为
//
//		  import (
//			   "github.com/rs/zerolog/log"
//			   "github.com/why2go/gostarter/config"
//		  )
//
//		  type HttpServer struct {
//			   Host string `yaml:"host" json:"host"`
//			   Port uint16 `yaml:"port" json:"port"`
//		  }
//
//		  func (HttpServer) GetConfigName() string {
//			   return "httpServer"  // 对应配置文件中配置项最外层名字
//		  }
//
//	  func init() {
//	    cfg := &HttpServer{}
//	    err := config.GetConfig(cfg)
//		   if err != nil {
//	      log.Fatal().Err(err).Msg("load http server configuration failed")
//		   }
//	    // apply loaded cfg
//	    // ...
//	  }
//
// 此时您可以在 resource/cfg/app-prod.yaml 增加如下配置项：
//
//	  httpServer:
//	    host: localhost
//		port: 8080
package config
