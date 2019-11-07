package config

import (
	"github.com/pubgo/g/pdd/config"
	"github.com/pubgo/g/xconfig/xconfig_instance"
	"github.com/pubgo/g/xerror"
	"sync"
)

// Config app
type Config struct {
	Cfg *config.Config
	Ext ext `toml:"ext"`
}

var _cfg *Config
var _once sync.Once

// Default global config instance
func Default() *Config {
	_once.Do(func() {
		_cfg = &Config{Cfg: xconfig_instance.Default()}
		xerror.Panic(xconfig_instance.ExtDecode(&_cfg.Ext))
	})
	return _cfg
}
