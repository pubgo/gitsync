package cmds

import (
	"github.com/jinzhu/gorm"
	"github.com/pubgo/g/xcmds"
	"github.com/pubgo/g/xcmds/migrate"
	"github.com/spf13/cobra"
)

// Execute exec
var Execute = xcmds.Init(func(cmd *cobra.Command) {
	cmd.Use = "gitsync"

	migrate.InitCommand(func(opt *migrate.Options) {
		opt.Db = &gorm.DB{}
	})
})
