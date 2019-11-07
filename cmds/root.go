package cmds

import (
	"fmt"
	"github.com/pubgo/g/logs"
	"github.com/pubgo/g/xcmds"
	"github.com/pubgo/gitsync/config"
	"github.com/pubgo/gitsync/version"
	"github.com/spf13/cobra"
)

const Service = "gitsync"

// Execute exec
var Execute = xcmds.Init(func(cmd *cobra.Command) {
	cmd.Use = Service
	cmd.Version = version.Version
}, func() {
	_l := logs.Default()
	_l.Version = version.Version
	_l.Service = Service
	_l.Init()

	//xcmds.InitLog()

	cfg := config.Default()
	fmt.Println(cfg.Ext)
})
