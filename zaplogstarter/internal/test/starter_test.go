package zaplogstarter

import (
	"fmt"
	"testing"
	"time"

	"github.com/why2go/gostarter/zaplogstarter"
	"go.uber.org/zap"
)

func TestZap(t *testing.T) {
	l, err := zaplogstarter.ZapConfig.Build()

	if err != nil {
		fmt.Println(err)
		return
	}

	l.Info("test", zap.String("hello", "world"), zap.Duration("lantency", 14*time.Second))

	sl := l.Sugar()

	sl.Debug(zap.Float64("pi", 3.1415926))
}
