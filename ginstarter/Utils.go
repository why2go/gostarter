package ginstarter

import (
	"github.com/gin-gonic/gin"
	ginlogger "github.com/why2go/gostarter/ginstarter/logger"
)

func GetRequestIdFromGinContext(ctx *gin.Context) string {
	v := ctx.Value(ginlogger.RequestIdKey)
	if reqId, ok := v.(string); ok {
		return reqId
	} else {
		return ""
	}
}
