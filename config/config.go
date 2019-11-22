package config

import (
	"github.com/pubgo/g/pdd/config"
	"github.com/pubgo/g/xconfig/xconfig_instance"
	"github.com/pubgo/g/xerror"
	"sync"
)

// Config app
type _Config struct {
	Cfg *config.Config
	Ext ext `toml:"ext"`
}

var Default = _init
var Init = _init

var _cfg *_Config
var _once sync.Once

// _init
// global config instance
func _init() *_Config {
	_once.Do(func() {
		_cfg = &_Config{Cfg: xconfig_instance.Default()}
		xerror.Panic(xconfig_instance.ExtDecode(&_cfg.Ext))
	})
	return _cfg
}
