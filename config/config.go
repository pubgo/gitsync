package config

import (
	"github.com/pubgo/g/xconfig"
	"github.com/pubgo/g/xerror"
	"github.com/pubgo/g/xinit"
)

// Config
type Config struct {
	Cfg *xconfig.Config
	Ext ext `toml:"ext"`
}

func init() {
	xinit.InitProvide(func(config *xconfig.Config) *Config {
		defer xerror.Assert()
		_cfg := &Config{Cfg: config}
		xerror.PanicM(xconfig.ExtDecode(&_cfg.Ext), "init config error")
		return _cfg
	})
}
