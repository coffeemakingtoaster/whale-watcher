package config

type GithubConfig struct {
	PAT      string `yaml:"pat" env:"PAT"`
	Username string `yaml:"username" env:"USER_NAME"`
}

type Config struct {
	Github GithubConfig `yaml:"github" envPrefix:"GH_"`
}
