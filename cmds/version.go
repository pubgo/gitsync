package cmds

import (
	"fmt"
	"github.com/pubgo/g/xcmds"
	"github.com/pubgo/gitsync/version"
	"github.com/spf13/cobra"
	"runtime"
)

func init() {
	xcmds.AddCommand(&cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Version:", version.Version)
			fmt.Println("GitHash:", version.CommitV)
			fmt.Println("BuildDate:", version.BuildV)
			fmt.Println("GoVersion:", runtime.Version())
			fmt.Println("GOOS:", runtime.GOOS)
			fmt.Println("GOARCH:", runtime.GOARCH)
		},
	})
}
