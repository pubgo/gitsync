package cmds

import (
	"github.com/pubgo/g/logs"
	"github.com/pubgo/g/xinit"
	"github.com/rs/zerolog/log"
)

var _logPkg = func() (string, string) { return "pkg", "gitsync" }
var logger = logs.DebugLog(_logPkg())

func init() {
	xinit.Init(func() error {
		logger = log.With().Str(_logPkg()).Logger()
		return nil
	})
}
