package main

import (
	"github.com/pubgo/g/xerror"
	"github.com/pubgo/gitsync/cmds"
	"runtime"
)

func main() {
	// git tag --sort=committerdate | tee | tail -n 1
	// git rev-parse --short=8 HEAD
	// git log -1 | tee | tail -n 1
	runtime.Version()
	xerror.Exit(cmds.Execute("$PWD"))("command error")
}
