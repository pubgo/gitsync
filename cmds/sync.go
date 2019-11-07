package cmds

import (
	"github.com/pubgo/g/xcmds"
	"github.com/pubgo/g/xerror"
	"github.com/spf13/cobra"
)

func init() {
	xcmds.AddCommand(&cobra.Command{
		Use:     "sync",
		Aliases: []string{"s"},
		Short:   "sync startup",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer xerror.RespErr(&err)

			return
		},
	})
}
