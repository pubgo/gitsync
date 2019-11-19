package cmds

import (
	"github.com/pubgo/g/pkg/encoding/cryptoutil"
	"github.com/pubgo/g/pkg/fileutil"
	"github.com/pubgo/g/xcmds"
	"github.com/pubgo/g/xconfig/xconfig_instance"
	"github.com/pubgo/g/xerror"
	"github.com/pubgo/gitsync/config"
	"github.com/spf13/cobra"
	"os"
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

			config.Default().Cfg.AppSecret = os.ExpandEnv(config.Default().Cfg.AppSecret)

			_cfg := config.Default().Ext.Sync

			if _cfg.RepoDir == "" {
				_cfg.RepoDir = repos
			}

			// 检查拉取代码的目录是否存在, 不存在创建
			_repoDir := filepath.Join(xconfig_instance.HomeDir(), _cfg.RepoDir)
			xerror.PanicM(fileutil.IsNotExistMkDir(_repoDir), "%s目录创建失败", _repoDir)

			var _repos []*repo
			for _, cfg := range _cfg.Cfg {
				if cfg.TimeOffset == 0 {
					cfg.TimeOffset = 7
					if _cfg.TimeOffset != 0 {
						cfg.TimeOffset = _cfg.TimeOffset
					}
				}

				if cfg.TimeInterval <= 0 {
					cfg.TimeInterval = 2
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
					cfg.FromUserPass = append(cfg.FromUserPass, _cfg.FromUserPass...)
				}

				if len(_cfg.ToUserPass) != 0 {
					cfg.ToUserPass = append(cfg.ToUserPass, _cfg.ToUserPass...)
				}

				xerror.PanicT(cfg.FromRepo == "" || cfg.ToRepo == "", "git repo error(from:%s, to:%s)", cfg.FromRepo, cfg.ToRepo)
				xerror.PanicT(len(cfg.FromUserPass) != 3 && len(cfg.ToUserPass) != 3, "git repo username, password and email is not set")

				var _repo = newRepo()
				_repo.RepoDir = _repoDir
				_repo.TimeInterval = cfg.TimeInterval
				_repo.TimeOffset = cfg.TimeOffset

				_repo.FromRepo = cfg.FromRepo
				_repo.FromBranch = cfg.FromBranch
				_repo.FromUserPass = cfg.FromUserPass
				if _repo.FromUserPass[1] == "" {
					_repo.FromUserPass[1] = os.Getenv("from_user_pass")
				}
				_repo.FromUserPass[1] = string(cryptoutil.MyXorDecrypt(os.ExpandEnv(_repo.FromUserPass[1]), []byte(config.Default().Cfg.AppSecret)))

				_repo.ToRepo = cfg.ToRepo
				_repo.ToBranch = cfg.ToBranch
				_repo.ToUserPass = cfg.ToUserPass
				if _repo.ToUserPass[1] == "" {
					_repo.ToUserPass[1] = os.Getenv("to_user_pass")
				}
				_repo.ToUserPass[1] = string(cryptoutil.MyXorDecrypt(os.ExpandEnv(_repo.ToUserPass[1]), []byte(config.Default().Cfg.AppSecret)))
				_repos = append(_repos, &_repo)
			}

			for {
				for _, repo := range _repos {
					repo.run()
				}
				time.Sleep(time.Minute)
			}
		},
	})
}
