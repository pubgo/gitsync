package main

import (
	"fmt"
	"github.com/pubgo/g/xerror"
	"github.com/pubgo/gitsync/cmds"
)

func main() {
	defer xerror.Resp(func(err *xerror.Err) {
		fmt.Println(err.P())
	})
	xerror.Panic(cmds.Execute("$PWD"))
}
