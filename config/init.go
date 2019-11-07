package config

// ext app
type ext struct {
	Sync struct {
		Cfg []struct {
			Name         string `toml:"name"`
			Desc         string `toml:"desc"`
			FromRepo     string `toml:"from_repo"`
			ToRep        string `toml:"to_rep"`
			TimeInterval string `toml:"time_interval"`
		} `toml:"cfg"`
	} `toml:"sync"`
}
