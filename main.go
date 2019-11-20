package main

import (
	"fmt"
	"github.com/pubgo/g/xerror"
	"github.com/pubgo/gitsync/cmds"
	"os/exec"
)

func main() {
	ss_, err := exec.Command("go", "version").Output()
	xerror.Exit(err)()
	fmt.Println(string(ss_))

	// git tag --sort=committerdate | tee | tail -n 1
	// git rev-parse --short=8 HEAD
	// git log -1 | tee | tail -n 1
	xerror.Exit(cmds.Execute("$PWD"))("command error")
}
