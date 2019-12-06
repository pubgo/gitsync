package cmds

import (
	"github.com/pubgo/g/logs"
	"github.com/pubgo/g/xdi"
	"github.com/rs/zerolog"
)

var logger = logs.DebugLog("pkg", "gitsync")

func init() {
	xdi.InitInvoke(func(log zerolog.Logger) {
		logger = log.With().Str("pkg", "gitsync").Logger()
	})
}
