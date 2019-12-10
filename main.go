package main

import (
	"github.com/pubgo/g/xerror"
	"github.com/pubgo/gitsync/cmds"
)

func main() {
	xerror.Exit(cmds.Execute(), "cmd error")
}
