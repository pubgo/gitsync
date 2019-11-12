package cmds

import (
	"github.com/pubgo/g/pkg/fileutil"
	"github.com/pubgo/g/xcmds"
	"github.com/pubgo/g/xconfig/xconfig_instance"
	"github.com/pubgo/g/xerror"
	"github.com/pubgo/gitsync/config"
	"github.com/spf13/cobra"
	"path/filepath"
	"time"
)

const repos = "repos"

func init() {
	xcmds.AddCommand(&cobra.Command{
		Use:     "sync",
		Aliases: []string{"s"},
		Short:   "sync startup",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer xerror.RespErr(&err)
			_cfg := config.Default().Ext.Sync

			if _cfg.RepoDir == "" {
				_cfg.RepoDir = repos
			}

			// 检查拉取代码的目录是否存在, 不存在创建
			_repoDir := filepath.Join(xconfig_instance.HomeDir(), _cfg.RepoDir)
			xerror.PanicM(fileutil.IsNotExistMkDir(_repoDir), "%s目录创建失败", _repoDir)

			var _repos []*repo
			for _, cfg := range _cfg.Cfg {
				if cfg.TimeInterval <= 0 {
					cfg.TimeInterval = 7
					if _cfg.TimeInterval > 0 {
						cfg.TimeInterval = _cfg.TimeInterval
					}
				}

				if cfg.FromBranch == "" {
					cfg.FromBranch = "master"
					if _cfg.FromBranch != "" {
						cfg.FromBranch = _cfg.FromBranch
					}
				}

				if cfg.ToBranch == "" {
					cfg.ToBranch = "master"
					if _cfg.ToBranch != "" {
						cfg.ToBranch = _cfg.ToBranch
					}
				}

				if len(_cfg.FromUserPass) != 0 {
					cfg.FromUserPass = _cfg.FromUserPass
				}

				if len(_cfg.ToUserPass) != 0 {
					cfg.ToUserPass = _cfg.ToUserPass
				}

				xerror.PanicT(cfg.FromRepo == "" || cfg.ToRepo == "", "git repo error(from:%s, to:%s)", cfg.FromRepo, cfg.ToRepo)
				xerror.PanicT(len(cfg.FromUserPass) != 3 && len(cfg.ToUserPass) != 3, "git repo username, password and email is not set")

				var _repo repo
				_repo.RepoDir = _repoDir
				_repo.TimeInterval = cfg.TimeInterval

				_repo.FromRepo = cfg.FromRepo
				_repo.FromBranch = cfg.FromBranch
				_repo.FromUserPass = cfg.FromUserPass

				_repo.ToRepo = cfg.ToRepo
				_repo.ToBranch = cfg.ToBranch
				_repo.ToUserPass = cfg.ToUserPass
				_repos = append(_repos, &_repo)
			}

			for {
				for _, repo := range _repos {
					go repo.run()
				}
				time.Sleep(time.Minute)
			}
		},
	})
}
