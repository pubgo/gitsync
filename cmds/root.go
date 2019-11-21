package cmds

import (
	"github.com/pubgo/g/logs"
	"github.com/pubgo/g/pkg"
	"github.com/pubgo/g/xcmds"
	"github.com/pubgo/g/xcmds/ss_cmd"
	"github.com/pubgo/g/xerror"
	"github.com/pubgo/gitsync/config"
	"github.com/pubgo/gitsync/version"
	"github.com/spf13/cobra"
)

const Service = "gitsync"

// Execute exec
var Execute = xcmds.Init("GS", func(cmd *cobra.Command) {
	defer xerror.Assert()

	cmd.Use = Service
	cmd.Version = version.Version

	// 添加加密命令
	ss_cmd.Init()
}, func() (err error) {
	defer xerror.RespErr(&err)

	_l := logs.Default()
	_l.Version = version.Version
	_l.Service = Service
	_l.Init()

	_cfg := config.Default()
	if pkg.IsDebug() {
		logs.P("config", _cfg)
	}

	return
})
