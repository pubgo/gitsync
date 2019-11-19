package cmds

import (
	"fmt"
	"github.com/pubgo/g/gotry"
	"github.com/pubgo/g/pkg/fileutil"
	"github.com/pubgo/g/xerror"
	"github.com/rs/zerolog/log"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

func newRepo() repo {
	return repo{mutex: new(sync.Mutex)}
}

type repo struct {
	RepoDir      string
	TimeInterval int `toml:"time_interval"`
	TimeOffset   int `toml:"time_offset"`

	FromUserPass []string `toml:"from_user_pass"`
	FromRepo     string   `toml:"from_repo"`
	FromBranch   string   `toml:"from_branch"`

	ToUserPass []string `toml:"to_user_pass"`
	ToRepo     string   `toml:"to_repo"`
	ToBranch   string   `toml:"to_branch"`

	commits    []*object.Commit
	curDate    string
	lastCommit *object.Commit

	mutex *sync.Mutex
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
// deprecated
func (t *repo) _remoteAdd() (err error) {
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

	log.Info().Str("repo", t.getRepoName(t.FromRepo)).Msg("remoteAdd ok")
	return

}

// 检查日期是否改变
// 重启或者时间增加一天，时期都会改变
func (t *repo) isDateChanged() bool {
	// 获取指定天数之前的那天的日期
	_curDate := time.Now().Add(time.Duration(t.TimeOffset) * time.Hour * 24).Format("2006-01-02")
	if _curDate != t.curDate {
		// 日期改变初始化
		t.curDate = _curDate
		t.commits = t.commits[:0]
		return true
	}
	return false
}

// 检查仓库是否存在，不存在
func (t *repo) pullFrom() (err error) {
	defer xerror.RespErr(&err)

	_repoDir := filepath.Join(t.RepoDir, t.getRepoName(t.FromRepo))
	r := xerror.PanicErr(git.PlainOpen(_repoDir + "_from")).(*git.Repository)
	w := xerror.PanicErr(r.Worktree()).(*git.Worktree)
	gotry.RetryAt(time.Second*10, func(i int) {
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
		}); err != nil &&
			err != git.NoErrAlreadyUpToDate &&
			!strings.Contains(err.Error(), "non-fast-forward update") &&
			!strings.Contains(err.Error(), "worktree contains unstaged") {
			xerror.PanicM(err, "git pull failed")
		}
	})

	log.Info().Str("repo_from", t.getRepoName(t.FromRepo)).Msg("pull ok")
	return
}

func (t *repo) pullTo() (err error) {
	defer xerror.RespErr(&err)

	_repoDir := filepath.Join(t.RepoDir, t.getRepoName(t.ToRepo))
	r := xerror.PanicErr(git.PlainOpen(_repoDir + "_to")).(*git.Repository)
	w := xerror.PanicErr(r.Worktree()).(*git.Worktree)
	gotry.RetryAt(time.Second*10, func(i int) {
		if err := w.Pull(&git.PullOptions{
			Auth: &http.BasicAuth{
				Username: t.ToUserPass[0],
				Password: t.ToUserPass[1],
			},
			Force:         true,
			SingleBranch:  true,
			RemoteName:    "origin",
			ReferenceName: plumbing.NewBranchReferenceName(t.ToBranch),
			Progress:      os.Stdout,
		}); err != nil &&
			err != git.NoErrAlreadyUpToDate &&
			!strings.Contains(err.Error(), "non-fast-forward update") &&
			!strings.Contains(err.Error(), "worktree contains unstaged") {
			xerror.PanicM(err, "git pull failed")
		}
	})

	log.Info().Str("repo_to", t.getRepoName(t.ToRepo)).Msg("pull ok")
	return
}

// 检查仓库是否存在, 不存在就clone
func (t *repo) clone() (err error) {
	defer xerror.RespErr(&err)

	_repoDir := filepath.Join(t.RepoDir, t.getRepoName(t.FromRepo))
	gotry.RetryAt(time.Minute, func(i int) {
		repoDirFrom := _repoDir + "_from"

		if !fileutil.CheckNotExist(repoDirFrom) {
			return
		}

		xerror.PanicErr(git.PlainClone(repoDirFrom, false, &git.CloneOptions{
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
		log.Info().Msgf("clone %s ok", repoDirFrom)
		xerror.Panic(t.pullFrom())
	})

	gotry.RetryAt(time.Minute, func(i int) {
		repoDirTo := _repoDir + "_to"

		if !fileutil.CheckNotExist(repoDirTo) {
			return
		}

		if _, err := git.PlainClone(repoDirTo, false, &git.CloneOptions{
			Auth: &http.BasicAuth{
				Username: t.ToUserPass[0],
				Password: t.ToUserPass[1],
			},
			URL:           t.ToRepo,
			SingleBranch:  true,
			RemoteName:    "origin",
			ReferenceName: plumbing.NewBranchReferenceName(t.ToBranch),
			Progress:      os.Stdout,
		}); err != nil && err != transport.ErrEmptyRemoteRepository {
			fmt.Println(err.Error())
			xerror.Panic(err)
		}

		log.Info().Msgf("clone %s ok", repoDirTo)

		xerror.Panic(t.pullTo())
	})

	return
}

// 得到当天的提交的所有的commit
func (t *repo) handleCommit() (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(t.pullFrom())

	_repoDir := filepath.Join(t.RepoDir, t.getRepoName(t.FromRepo))
	rFrom := xerror.PanicErr(git.PlainOpen(_repoDir + "_from")).(*git.Repository)

	cIter := xerror.PanicErr(rFrom.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})).(object.CommitIter)
	defer cIter.Close()

	_now := time.Now().Add(time.Duration(t.TimeOffset) * time.Hour * 24)
	xerror.PanicM(cIter.ForEach(func(c *object.Commit) error {
		//fmt.Println(c.Committer.When.String())
		//&& c.Committer.When.After(_now)
		if c.Committer.When.Format("2006-01-02") == t.curDate {
			t.commits = append(t.commits, c)
		}

		// 获取指定当天commit之前的一个commit
		if c.Committer.When.Before(_now) {
			// 如果github仓库就一次commit，那么，就是第一次提交，把所有代码提交了
			if t.isFirstTime() {
				xerror.PanicM(t._commitAndPush(c), "git commit error")
				log.Info().Str("repo", t.getRepoName(t.FromRepo)).Msg("git check ok")
			}
			return xerror.ErrDone
		}

		return nil
	}), "git commit iter failed")

	// 按照时间从小到大 commit 排序
	sort.Slice(t.commits, func(i, j int) bool {
		return t.commits[i].Committer.When.Before(t.commits[j].Committer.When)
	})

	for i, c := range t.commits {
		fmt.Println(i, "today commit: ", c.Committer.When.String(), c.Hash.String())
	}

	log.Info().Str("repo", t.getRepoName(t.FromRepo)).Msg("handleCommit ok")

	xerror.PanicM(t.getLastCommitFromNewRepo(), "get lastCommit error")

	return
}

func (t *repo) getLastCommitFromNewRepo() (err error) {
	defer xerror.RespErr(&err)

	xerror.Panic(t.pullTo())

	_repoDir := filepath.Join(t.RepoDir, t.getRepoName(t.FromRepo))
	tFrom := xerror.PanicErr(git.PlainOpen(_repoDir + "_to")).(*git.Repository)

	cIter := xerror.PanicErr(tFrom.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})).(object.CommitIter)
	defer cIter.Close()

	t.lastCommit = xerror.PanicErr(cIter.Next()).(*object.Commit)
	t.lastCommit.Committer.When = t.lastCommit.Committer.When.Add(time.Duration(t.TimeOffset) * time.Hour * 24)
	log.Info().Str("repo", t.getRepoName(t.FromRepo)).Msg("lastCommit ok")
	return
}

func (t *repo) isFirstTime() bool {
	xerror.Panic(t.pullTo())

	_repoDir := filepath.Join(t.RepoDir, t.getRepoName(t.FromRepo))
	tFrom := xerror.PanicErr(git.PlainOpen(_repoDir + "_to")).(*git.Repository)

	cIter := xerror.PanicErr(tFrom.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})).(object.CommitIter)
	defer cIter.Close()

	i := 0
	xerror.Panic(cIter.ForEach(func(_ *object.Commit) error {
		i++
		return nil
	}))

	return i == 1
}

func (t *repo) _commitAndPush(c *object.Commit) (err error) {
	defer xerror.RespErr(&err)

	_repoDir := filepath.Join(t.RepoDir, t.getRepoName(t.FromRepo))
	rFrom := xerror.PanicErr(git.PlainOpen(_repoDir + "_from")).(*git.Repository)
	wFrom := xerror.PanicErr(rFrom.Worktree()).(*git.Worktree)
	xerror.PanicM(t.pullFrom(), "git pull failed")

	xerror.PanicM(wFrom.Reset(&git.ResetOptions{
		Commit: c.Hash,
		Mode:   git.HardReset,
	}), "git reset failed")

	for _, f := range xerror.PanicErr(ioutil.ReadDir(_repoDir + "_from")).([]os.FileInfo) {
		if f.Name() == ".git" || f.Name() == ".DS_Store" {
			continue
		}

		_cmd := exec.Command("cp", "-Rf", fmt.Sprintf(`%s%s`, _repoDir, "_from/"+f.Name()), _repoDir+"_to/"+f.Name())
		_dd, err := _cmd.CombinedOutput()
		xerror.PanicM(err, "copy from repo %s", _dd)
	}

	rTo := xerror.PanicErr(git.PlainOpen(_repoDir + "_to")).(*git.Repository)
	wTo := xerror.PanicErr(rTo.Worktree()).(*git.Worktree)
	fmt.Println(_repoDir + "_to")

	_status := xerror.PanicErr(wTo.Status()).(git.Status)
	for k := range _status {
		xerror.PanicErr(wTo.Add(k))
	}

	// 提交commit
	xerror.PanicErr(wTo.Commit(c.Message, &git.CommitOptions{
		All: true,
		Author: &object.Signature{
			Name:  c.Author.Name,
			Email: c.Author.Email,
			When:  time.Now(),
		},
	}))

	if err := rTo.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: t.ToUserPass[0],
			Password: t.ToUserPass[1],
		},
		Progress: os.Stdout,
		RefSpecs: []config.RefSpec{"+" + config.DefaultPushRefSpec},
	}); err != nil && err != git.NoErrAlreadyUpToDate && !strings.Contains(err.Error(), "non-fast-forward update") {
		xerror.PanicM(err, "git 仓库 %s push failed", t.ToRepo)
	}

	// 最后把lastCommit提前一位
	t.lastCommit = c

	log.Info().Str("repo", t.getRepoName(t.FromRepo)).Msg("commitAndPush ok")

	xerror.PanicM(t.pullFrom(), "git pull failed")
	return
}

func (t *repo) commitAndPush() (err error) {
	defer xerror.RespErr(&err)

	var _curCommit *object.Commit = nil
	_now := time.Now().Add(time.Duration(t.TimeOffset) * time.Hour * 24)
	// 距离commit在TimeInterval分钟之内，就提交了
	_timeInterval := math.Abs((time.Duration(t.TimeInterval) * time.Minute).Seconds())
	for _, c := range t.commits {
		if math.Abs(c.Committer.When.Sub(_now).Seconds()) < _timeInterval {
			//fmt.Println("ok", c.Committer.When.String(), t.lastCommit.Committer.When.String())
			if c.Committer.When.Sub(t.lastCommit.Committer.When).Seconds() <= 0 {
				continue
			}

			_curCommit = c
			break
		}
	}

	// 定时检查更新
	if _curCommit == nil {
		log.Info().Msg("没有更新")
		return
	}

	xerror.PanicM(t._commitAndPush(_curCommit), "git commit error")
	return
}

func (t *repo) run() {
	defer xerror.Resp(func(err *xerror.Err) {
		fmt.Println(err.P())
		os.Exit(-1)
	})

	defer t.mutex.Unlock()
	t.mutex.Lock()

	log.Info().Msgf("startup %s", t.getRepoName(t.FromRepo))

	xerror.Panic(t.clone())

	// 时间更加一天，自动的更新信息
	if t.isDateChanged() {
		xerror.Panic(t.handleCommit())
	}

	xerror.Panic(t.commitAndPush())
	log.Info().Msgf("over %s", t.getRepoName(t.FromRepo))
	fmt.Print("\n\n")
}
