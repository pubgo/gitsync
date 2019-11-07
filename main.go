package main

import (
	"github.com/pubgo/g/xerror"
	"github.com/pubgo/gitsync/cmds"
	"os"
)

func main() {
	defer xerror.Debug()
	xerror.Panic(cmds.Execute("GS", os.ExpandEnv("$PWD")))
}
