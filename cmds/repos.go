package cmds

import (
	"github.com/pubgo/g/pkg/fileutil"
	"github.com/pubgo/g/xerror"
	"github.com/rs/zerolog/log"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type repo struct {
	RepoDir      string
	TimeInterval int `toml:"time_interval"`

	FromUserPass []string `toml:"from_user_pass"`
	FromRepo     string   `toml:"from_repo"`
	FromBranch   string   `toml:"from_branch"`

	ToUserPass []string `toml:"to_user_pass"`
	ToRepo     string   `toml:"to_repo"`
	ToBranch   string   `toml:"to_branch"`

	commits    []*object.Commit
	curDate    string
	lastCommit *object.Commit

	mutex *sync.RWMutex
}

// 根据git地址获取域名, 用来当做remote origin 名字
func (t *repo) getRepoDomain(repo string) string {
	return xerror.PanicErr(url.Parse(repo)).(*url.URL).Hostname()
}

// 根据git地址获取git repo名字
func (t *repo) getRepoName(repo string) string {
	_us := strings.Split(repo, "/")
	_name := _us[len(_us)-1]
	if strings.HasSuffix(_name, ".git") {
		return _name[:len(_name)-4]
	}
	return _name
}

// 添加 git remote origin
func (t *repo) remoteAdd() (err error) {
	defer xerror.RespErr(&err)

	_repoDir := filepath.Join(t.RepoDir, t.getRepoName(t.FromRepo))
	r := xerror.PanicErr(git.PlainOpen(_repoDir)).(*git.Repository)

	_name := t.getRepoDomain(t.ToRepo)
	xerror.PanicErr(r.CreateRemote(&config.RemoteConfig{
		Name: _name,
		URLs: []string{t.ToRepo},
	}))

	// 检查是否添加成功
	_ok := false
	for _, r := range xerror.PanicErr(r.Remotes()).([]*git.Remote) {
		if r.Config().Name == _name {
			_ok = true
		}
	}

	xerror.PanicT(!_ok, "remote(%s,%s)添加失败", _name, t.ToRepo)

	log.Info().Msg("remoteAdd ok")
	return

}

// 检查日期是否改变
// 重启或者时间增加一天，时期都会改变
func (t *repo) isDateChanged() bool {
	// 获取指定天数之前的那天的日期
	_curDate := time.Now().Add(-time.Duration(t.TimeInterval) * time.Hour * 24).Format("2006-01-02")
	if _curDate != t.curDate {
		// 日期改变初始化
		t.curDate = _curDate
		t.commits = t.commits[:0]
		return true
	}

	return false
}

// 检查仓库是否存在，不存在
func (t *repo) pull() (err error) {
	defer xerror.RespErr(&err)

	_repoDir := filepath.Join(t.RepoDir, t.getRepoName(t.FromRepo))
	r := xerror.PanicErr(git.PlainOpen(_repoDir)).(*git.Repository)
	w := xerror.PanicErr(r.Worktree()).(*git.Worktree)
	if err := w.Pull(&git.PullOptions{
		Auth: &http.BasicAuth{
			Username: t.FromUserPass[0],
			Password: t.FromUserPass[1],
		},
		Force:         true,
		SingleBranch:  true,
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName(t.FromBranch),
		Progress:      os.Stdout,
	}); err != nil && err != git.NoErrAlreadyUpToDate && !strings.Contains(err.Error(), "non-fast-forward update") {
		xerror.PanicM(err, "git pull failed")
	}

	log.Info().Msg("pull ok")
	return
}

// 检查仓库是否存在, 不存在就clone
func (t *repo) clone() (err error) {
	defer xerror.RespErr(&err)

	_repoDir := filepath.Join(t.RepoDir, t.getRepoName(t.FromRepo))
	if !fileutil.CheckNotExist(_repoDir) {
		return
	}

	xerror.PanicErr(git.PlainClone(_repoDir, false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: t.FromUserPass[0],
			Password: t.FromUserPass[1],
		},
		URL:           t.FromRepo,
		SingleBranch:  true,
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName(t.FromBranch),
		Progress:      os.Stdout,
	}))

	// 添加远程url失败
	xerror.PanicM(t.remoteAdd(), "git remote origin add failed")

	log.Info().Msg("clone ok")
	return
}

// 得到当天的提交的所有的commit
func (t *repo) handleCommit() (err error) {
	defer xerror.RespErr(&err)

	xerror.PanicM(t.pull(), "handle git pull failed")

	_repoDir := filepath.Join(t.RepoDir, t.getRepoName(t.FromRepo))
	r := xerror.PanicErr(git.PlainOpen(_repoDir)).(*git.Repository)

	cIter := xerror.PanicErr(r.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})).(object.CommitIter)
	defer cIter.Close()

	xerror.PanicM(cIter.ForEach(func(c *object.Commit) error {
		if c.Committer.When.Format("2006-01-02") == t.curDate {
			t.commits = append(t.commits, c)
		}

		// 获取指定当天commit之前的一个commit
		if c.Committer.When.Before(xerror.PanicErr(time.Parse("2006-01-02", t.curDate)).(time.Time)) {
			if t.lastCommit == nil {
				t.lastCommit = c
				return xerror.ErrDone
			}
		}
		return nil
	}), "git commit iter failed")

	// 按照时间从小到大 commit 排序
	sort.Slice(t.commits, func(i, j int) bool {
		return t.commits[i].Committer.When.Before(t.commits[j].Committer.When)
	})

	log.Info().Msg("handleCommit ok")
	return
}

func (t *repo) commitAndPush() (err error) {
	defer xerror.RespErr(&err)

	var _curCommit *object.Commit = nil
	_now := time.Now().Add(-time.Duration(t.TimeInterval) * time.Hour * 24)
	for _, c := range t.commits {
		//fmt.Println(c.Committer.When.String())
		// 距离commit在两分钟之内，就提交了
		if c.Committer.When.Sub(_now) < 2*time.Minute {
			_curCommit = c
			break
		}
	}

	// 定时检查更新
	if _curCommit == nil || _curCommit == t.lastCommit {
		//fmt.Println("2006-01-02", "没有更新")
		return
	}

	_repoDir := filepath.Join(t.RepoDir, t.getRepoName(t.FromRepo))
	r := xerror.PanicErr(git.PlainOpen(_repoDir)).(*git.Repository)
	w := xerror.PanicErr(r.Worktree()).(*git.Worktree)
	xerror.PanicM(t.pull(), "git pull failed")

	xerror.PanicM(w.Reset(&git.ResetOptions{
		Commit: _curCommit.Hash,
		Mode:   git.HardReset,
	}), "git reset failed")

	xerror.PanicM(w.Reset(&git.ResetOptions{
		Commit: t.lastCommit.Hash,
		Mode:   git.SoftReset,
	}), "git reset failed")

	// 提交commit
	xerror.PanicErr(w.Commit(_curCommit.Message, &git.CommitOptions{
		All: true,
		Author: &object.Signature{
			Name:  t.ToUserPass[0],
			Email: t.ToUserPass[2],
			When:  time.Now(),
		},
	}))

	if err := r.Push(&git.PushOptions{
		RemoteName: t.getRepoDomain(t.ToRepo),
		Auth: &http.BasicAuth{
			Username: t.ToUserPass[0],
			Password: t.ToUserPass[1],
		},
		Progress: os.Stdout,
	}); err != nil && err != git.NoErrAlreadyUpToDate && !strings.Contains(err.Error(), "non-fast-forward update") {
		xerror.Panic(err)
	}

	// 把提交的commit删除，不然需要合并新提交的commit
	xerror.PanicM(w.Reset(&git.ResetOptions{
		Commit: t.lastCommit.Hash,
		Mode:   git.HardReset,
	}), "git last commit reset failed")

	// 最后把lastCommit提前一位
	t.lastCommit = _curCommit

	log.Info().Msg("commitAndPush ok")
	return
}

func (t *repo) run() {
	defer xerror.Debug()

	xerror.Panic(t.clone())

	// 时间更加一天，自动的更新信息
	if t.isDateChanged() {
		xerror.Panic(t.handleCommit())
	}

	xerror.Panic(t.commitAndPush())
	log.Info().Msg("handle over")
}
