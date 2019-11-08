package config

// ext app
type ext struct {
	Sync struct {
		RepoDir string `toml:"repo_dir"`
		Cfg     []struct {
			TimeInterval int    `toml:"time_interval"`
			RepoName     string `toml:"repo_name"`
			FromRepo     string `toml:"from_repo"`
			FromBranch   string `toml:"from_branch"`
			ToRepo       string `toml:"to_repo"`
			ToBranch     string `toml:"to_branch"`
		} `toml:"cfg"`
	} `toml:"sync"`
}

