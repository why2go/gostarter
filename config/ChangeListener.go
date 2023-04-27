package config

import "sync"

type Changeable interface {
	OnChange(cfg interface{})
}

// 用于管理配置变更
var (
	listeners     = make(map[interface{}]struct{})
	listenersLock sync.Mutex
)

func RegisterConfigChangeListener(ls Changeable) {
	listeners[ls] = struct{}{}
}

func RemoveConfigChangeListener(ls Changeable) {
	delete(listeners, ls)
}
