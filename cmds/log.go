package cmds

import (
	"github.com/pubgo/g/logs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var _logPkg = func() (string, string) { return "pkg", "gitsync" }
var logger = logs.DebugLog(_logPkg())

func InitLog(cnt ...zerolog.Context) {
	logger = log.With().Str(_logPkg()).Logger()
	if len(cnt) != 0 {
		logger = cnt[0].Str(_logPkg()).Logger()
	}
}
