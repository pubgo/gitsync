package config

// ext app
type ext struct {
	Sync struct {
		RepoDir      string   `toml:"repo_dir"`
		FromBranch   string   `toml:"from_branch"`
		ToBranch     string   `toml:"to_branch"`
		TimeInterval int      `toml:"time_interval"`
		FromUserPass []string `toml:"from_user_pass"`
		ToUserPass   []string `toml:"to_user_pass"`
		Cfg          []struct {
			TimeInterval int      `toml:"time_interval"`
			FromUserPass []string `toml:"from_user_pass"`
			FromRepo     string   `toml:"from_repo"`
			FromBranch   string   `toml:"from_branch"`
			ToUserPass   []string `toml:"to_user_pass"`
			ToRepo       string   `toml:"to_repo"`
			ToBranch     string   `toml:"to_branch"`
		} `toml:"cfg"`
	} `toml:"sync"`
}
