package cmds

import (
	"github.com/pubgo/g/xcmds"
	"github.com/pubgo/g/xcmds/xcmd_ss"
	"github.com/pubgo/g/xerror"
	"github.com/pubgo/gitsync/version"
)

const Service = "gitsync"

// Execute exec
var Execute = xcmds.Init("GS", func(cmd *xcmds.Command) {
	defer xerror.Assert()

	cmd.Use = Service
	cmd.Version = version.Version

	// 添加加密命令
	xcmd_ss.Init()
})
