package config

import (
	"github.com/pubgo/g/xconfig"
	"github.com/pubgo/g/xdi"
	"github.com/pubgo/g/xerror"
)

// Config
type Config struct {
	Cfg *xconfig.Config
	Ext ext `toml:"ext"`
}

func init() {
	xdi.InitProvide(func(config *xconfig.Config) (_cfg *Config, err error) {
		defer xerror.RespErr(&err)
		_cfg = &Config{Cfg: config}
		xerror.PanicM(xconfig.ExtDecode(&_cfg.Ext), "init config error")
		return
	})
}
