package ginstarter

import (
	"github.com/gin-gonic/gin"
)

func GetRequestIdFromGinContext(ctx *gin.Context) string {
	v := ctx.Value("request_id")
	if reqId, ok := v.(string); ok {
		return reqId
	} else {
		return ""
	}
}
