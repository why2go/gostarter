package config

import (
	"reflect"
	"sync"

	"github.com/rs/zerolog/log"
)

type Changeable interface {
	Configurable
	OnChange(cfg interface{})
}

var (
	listenerMgr = newListenerManager()
)

// 用于管理配置变更
type listenerManager struct {
	lk        sync.Mutex
	listeners map[string]Changeable
}

func newListenerManager() *listenerManager {
	l := &listenerManager{
		lk:        sync.Mutex{},
		listeners: make(map[string]Changeable),
	}
	return l
}

func RegisterConfigChangeListener(ls Changeable) {
	listenerMgr.lk.Lock()
	defer listenerMgr.lk.Unlock()
	listenerMgr.listeners[ls.GetConfigName()] = ls
}

func RemoveConfigChangeListener(ls Changeable) {
	listenerMgr.lk.Lock()
	defer listenerMgr.lk.Unlock()
	delete(listenerMgr.listeners, ls.GetConfigName())
}

func (mgr *listenerManager) handleChangeEvent(cfgName string) {
	if c, ok := mgr.listeners[cfgName]; ok {
		typ := reflect.TypeOf(c)
		for typ.Kind() == reflect.Pointer {
			typ = typ.Elem()
		}
		obj := reflect.New(typ).Interface().(Configurable)
		err := GetConfig(obj)
		if err != nil {
			log.Error().Err(err).Msg("handle change event failed")
			return
		}
		c.OnChange(obj)
	}
}
