package interceptor

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

const (
	skipAll = "*"
)

type UnaryConf struct {
	methodsSet map[string]struct{}
}

func NewUnaryConf(skippedMethods []string) *UnaryConf {
	cfg := &UnaryConf{
		methodsSet: make(map[string]struct{}),
	}
	for i := range skippedMethods {
		cfg.methodsSet[skippedMethods[i]] = struct{}{}
	}
	return cfg
}

func (cfg *UnaryConf) isAllSkipped() bool {
	_, b := cfg.methodsSet[skipAll]
	return b
}

func (cfg *UnaryConf) isMethodSkipped(method string) bool {
	_, b := cfg.methodsSet[method]
	return b
}

type tsKey string

var (
	startAtKey = tsKey("startAt")
)

func UnaryIncomingInterceptor(conf *UnaryConf) grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		if conf.isAllSkipped() || conf.isMethodSkipped(info.FullMethod) {
			return handler(ctx, req)
		}
		start := time.Now()
		var requestId string = genRequestId()
		// var peerAddr string
		// if p, ok := peer.FromContext(ctx); ok {
		// 	peerAddr = p.Addr.String()
		// }
		var newCtx context.Context
		newCtx = context.WithValue(ctx, "requestId", requestId)
		newCtx = context.WithValue(newCtx, startAtKey, start)
		// log.Log().
		// 	Str("method", info.FullMethod).
		// 	Str("peer", peerAddr).
		// 	Str("startAt", start.Format(time.RFC3339)).
		// 	Interface("request", req).
		// 	Str("requestId", requestId).
		// 	Str("stage", "incoming").
		// 	Send()
		return handler(newCtx, req)
	}
}

func UnaryOutgoingInterceptor(conf *UnaryConf) grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		if conf.isAllSkipped() || conf.isMethodSkipped(info.FullMethod) {
			return handler(ctx, req)
		}
		var requestId string
		if val := ctx.Value("requestId"); val != nil {
			requestId = val.(string)
		}
		var startAt time.Time
		if val := ctx.Value(startAtKey); val != nil {
			startAt = val.(time.Time)
		}
		var peerAddr string
		if p, ok := peer.FromContext(ctx); ok {
			peerAddr = p.Addr.String()
		}
		resp, err = handler(ctx, req)
		event := log.Log().
			Str("method", info.FullMethod).
			Str("peer", peerAddr).
			Interface("request", req)
		if err != nil {
			event = event.Err(err)
		}
		if len(requestId) != 0 {
			event = event.Str("requestId", requestId)
		}
		if !startAt.IsZero() {
			event = event.Dur("latencyMS", time.Since(startAt)).
				Str("startAt", startAt.UTC().Format(time.RFC3339))
		}
		event.Send()
		return
	}
}

func genRequestId() string {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}
