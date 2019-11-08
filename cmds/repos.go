package cmds

import (
	"fmt"
	"github.com/pubgo/g/pkg/fileutil"
	"github.com/pubgo/g/xerror"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

//1. 启动的时候，检查，代码是否拉取，没有的那么，开始拉取代码，拉取之后的，并设置另一个remote origin 标记O1， 然后更新代码到最新
//2. 获取两个月之前的改天的所有的需要提交的commit，并获取id，时间和msg
//3. 获取距离两个月之前而当time最近的那一次commit的信息 标记为C1
//4. git reset--hard C1.id
//5. git reset--soft C1.id 的上一个CID
//6. git commit -m "C1.msg"
//6. git push O1 O1/branch

type repo struct {
	RepoDir      string
	TimeInterval int
	RepoName     string

	FromRepo   string
	FromBranch string
	ToRepo     string
	ToBranch   string

	commits []*object.Commit
	CurDate string

	mutex *sync.RWMutex
}

func (t *repo) getRepoDomain() string {
	_url := xerror.PanicErr(url.Parse(t.ToRepo)).(*url.URL)
	return _url.Hostname()
}

func (t *repo) getRepoName() string {
	_us := strings.Split(t.ToRepo, "/")
	_name := _us[len(_us)-1]
	if strings.HasSuffix(_name, ".git") {
		return _name[:len(_name)-4]
	}
	return _name
}

// 添加 remote
func (t *repo) remoteAdd() (err error) {
	defer xerror.RespErr(&err)

	_repoDir := filepath.Join(t.RepoDir, t.RepoName)
	r := xerror.PanicErr(git.PlainOpen(_repoDir)).(*git.Repository)

	_name := t.getRepoDomain()
	xerror.PanicErr(r.CreateRemote(&config.RemoteConfig{
		Name: _name,
		URLs: []string{t.ToRepo},
	}))

	_ok := false
	for _, r := range xerror.PanicErr(r.Remotes()).([]*git.Remote) {
		if r.Config().Name == _name {
			_ok = true
		}
	}

	xerror.PanicT(!_ok, "remote(%s,%s)添加失败", _name, t.ToRepo)

	return

}

// 检查日期是否改变
func (t *repo) isDateChanged() bool {
	if t.CurDate == "" {
		return true
	}

	_curDate := time.Now().Add(-time.Duration(t.TimeInterval) * time.Hour * 24).Format("2006-01-02")
	if _curDate != t.CurDate {
		t.CurDate = _curDate
		return true
	}

	return false
}

// 检查仓库是否存在，不存在
func (t *repo) pull() (err error) {
	defer xerror.RespErr(&err)

	_repoDir := filepath.Join(t.RepoDir, t.RepoName)
	r := xerror.PanicErr(git.PlainOpen(_repoDir)).(*git.Repository)
	w := xerror.PanicErr(r.Worktree()).(*git.Worktree)
	xerror.PanicM(w.Pull(&git.PullOptions{
		Auth: &http.BasicAuth{
			Username: "",
			Password: "",
		},
		SingleBranch:  true,
		RemoteName:    "origin",
		ReferenceName: plumbing.ReferenceName(t.FromBranch),
		Progress:      os.Stdout,
	}), "git pull failed")

	return
}

// 检查仓库是否存在, 不存在就clone
func (t *repo) clone() (err error) {
	defer xerror.RespErr(&err)

	_repoDir := filepath.Join(t.RepoDir, t.RepoName)
	if !fileutil.CheckNotExist(_repoDir) {
		return
	}

	xerror.PanicErr(git.PlainClone(_repoDir, false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: "",
			Password: "",
		},
		URL:           t.FromRepo,
		SingleBranch:  true,
		RemoteName:    "origin",
		ReferenceName: plumbing.ReferenceName(t.FromBranch),
		Progress:      os.Stdout,
	}))

	return
}

// 得到当天的提交的所有的commit
func (t *repo) handleCommit() (err error) {
	defer xerror.RespErr(&err)

	_repoDir := filepath.Join(t.RepoDir, t.RepoName)
	r := xerror.PanicErr(git.PlainOpen(_repoDir)).(*git.Repository)

	cIter := xerror.PanicErr(r.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
	})).(object.CommitIter)
	defer cIter.Close()

	xerror.PanicM(cIter.ForEach(func(c *object.Commit) error {
		fmt.Println(c)

		c.Committer.When.Format("2006-01-02")

		return nil
	}), "commit iter error")

	return
}

func (t *repo) commitAndPush() (err error) {
	defer xerror.RespErr(&err)

	for _, c := range t.commits {
	}

	//w.Reset(&git.ResetOptions{
	//	Commit: "",
	//	Mode:   "",
	//})

	_repoDir := filepath.Join(t.RepoDir, t.RepoName)
	r := xerror.PanicErr(git.PlainOpen(_repoDir)).(*git.Repository)

	w := xerror.PanicErr(r.Worktree()).(*git.Worktree)
	xerror.PanicErr(w.Commit("example go-git commit", &git.CommitOptions{
		All: true,
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
			When:  time.Now(),
		},
	}))

	xerror.Panic(r.Push(&git.PushOptions{}))

	return
}

func (t *repo) run() {
	defer xerror.Debug()

	xerror.Panic(t.clone())

	// 时间改变
	if t.isDateChanged() {
		xerror.Panic(t.pull())
		xerror.Panic(t.handleCommit())
	}

	xerror.Panic(t.commitAndPush())
}
