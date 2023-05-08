package grpcstarter

import (
	"fmt"
	"net"
	"time"

	config "github.com/why2go/gostarter/config"
	"github.com/why2go/gostarter/grpcstarter/interceptor"
	_ "github.com/why2go/gostarter/zerologstarter"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

var (
	logger = log.With().Str("ltag", "grpcStarter").Logger()
)

func init() {
	cfg := &grpcConf{}
	err := config.GetConfig(cfg)
	if err != nil && err != config.ErrNoConfigItemFound {
		logger.Fatal().Err(err).Msg("load grpc server config failed")
		return
	}

	grpcSrvInstance = newGrpcServer(cfg)
}

type grpcConf struct {
	Host            string `yaml:"host" json:"host"`
	Port            uint16 `yaml:"port" json:"port"`
	ConnTimeoutMS   uint32 `yaml:"connTimeoutMS" json:"connTimeoutMS"`
	WriteBufferSize uint32 `yaml:"writeBufferSize" json:"writeBufferSize"`
	ReadBufferSize  uint32 `yaml:"readBufferSize" json:"readBufferSize"`
	Logger          struct {
		SkipMethods []string `yaml:"skipMethods" json:"skipMethods"`
	} `yaml:"logger" json:"logger"`
	// 暂时废弃interceptors
	Interceptors []string `yaml:"interceptors" json:"interceptors"` // incoming or outgoing
}

func (cfg *grpcConf) ConfigName() string {
	return "grpc"
}

const (
	defaultPort = uint16(8081)
)

var (
	grpcSrvInstance *grpcServer
)

type grpcServer struct {
	host       string
	port       uint16
	opts       []grpc.ServerOption
	grpcServer *grpc.Server
}

func newGrpcServer(cfg *grpcConf) *grpcServer {
	srv := &grpcServer{host: cfg.Host}
	if cfg.Port == 0 {
		srv.port = defaultPort
	} else {
		srv.port = cfg.Port
	}
	if cfg.WriteBufferSize != 0 {
		srv.opts = append(srv.opts, grpc.WriteBufferSize(int(cfg.WriteBufferSize)))
	}
	if cfg.ReadBufferSize != 0 {
		srv.opts = append(srv.opts, grpc.ReadBufferSize(int(cfg.ReadBufferSize)))
	}
	if cfg.ConnTimeoutMS != 0 {
		srv.opts = append(srv.opts,
			grpc.ConnectionTimeout(time.Duration(cfg.ConnTimeoutMS)*time.Millisecond))
	}
	usi := defaultChainedInterceptors(interceptor.NewUnaryConf(cfg.Logger.SkipMethods))
	srv.opts = append(srv.opts,
		grpc.ChainUnaryInterceptor(usi...))
	srv.grpcServer = grpc.NewServer(srv.opts...)
	return srv
}

// 注册grpc service实现
func RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	grpcSrvInstance.grpcServer.RegisterService(sd, ss)
}

func StartGrpcServer() {
	addr := fmt.Sprintf("%s:%d", grpcSrvInstance.host, grpcSrvInstance.port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal().Err(err).Msg("grpc listen error")
		return
	}
	// 不要阻塞主线程
	go func() {
		err = grpcSrvInstance.grpcServer.Serve(lis)
		if err != nil {
			logger.Fatal().Err(err).Msg("grpc serve error")
			return
		}
	}()
}

func StopGrpcServer() {
	logger.Info().Msg("shutting down grpc server...")
	grpcSrvInstance.grpcServer.GracefulStop()
	logger.Info().Msg("grpc server is closed")
}

// var (
// 	availableInts = map[string]func(*interceptor.UnaryConf) grpc.UnaryServerInterceptor{
// 		"incoming": interceptor.UnaryIncomingInterceptor,
// 		"outgoing": interceptor.UnaryOutgoingInterceptor,
// 	}
// )

// func getChainedInterceptors(intNames []string, conf *interceptor.UnaryConf) []grpc.UnaryServerInterceptor {
// 	var ints []grpc.UnaryServerInterceptor
// 	for i := range intNames {
// 		if intNames[i] == "incoming" {
// 			t := availableInts["incoming"](conf)
// 			ints = append([]grpc.UnaryServerInterceptor{t}, ints...)
// 			continue
// 		}
// 		if f, ok := availableInts[intNames[i]]; ok {
// 			ints = append(ints, f(conf))
// 		}
// 	}
// 	return ints
// }

func defaultChainedInterceptors(conf *interceptor.UnaryConf) []grpc.UnaryServerInterceptor {
	ints := []grpc.UnaryServerInterceptor{
		interceptor.UnaryIncomingInterceptor(conf),
		interceptor.UnaryOutgoingInterceptor(conf),
	}
	return ints
}
