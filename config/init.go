package config

// ext app
type ext struct {
	Sync struct {
		TimeInterval int      `toml:"time_interval"`
		TimeOffset   int      `toml:"time_offset"`
		RepoDir      string   `toml:"repo_dir"`
		FromBranch   string   `toml:"from_branch"`
		FromUserPass []string `toml:"from_user_pass"`
		ToBranch     string   `toml:"to_branch"`
		ToUserPass   []string `toml:"to_user_pass"`
		Cfg          []struct {
			FromRepo     string   `toml:"from_repo"`
			ToRepo       string   `toml:"to_repo"`
			TimeInterval int      `toml:"time_interval"`
			TimeOffset   int      `toml:"time_offset"`
			RepoDir      string   `toml:"repo_dir"`
			FromBranch   string   `toml:"from_branch"`
			FromUserPass []string `toml:"from_user_pass"`
			ToBranch     string   `toml:"to_branch"`
			ToUserPass   []string `toml:"to_user_pass"`
		} `toml:"cfg"`
	} `toml:"sync"`
}
