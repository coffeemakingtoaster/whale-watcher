package config

import (
	"errors"

	"github.com/spf13/viper"
)

type TargetConfig struct {
	RepositoryURL  string `mapstructure:"repository" env:"REPOSITORY_URL" desc:"Specify a remote repository. This can be empty for local."`
	DockerfilePath string `mapstructure:"dockerfile" env:"DOCKERFILE" desc:"Specify the dockerfile path. This is either a local path or the location of the Dockerfile in the specified repository"`
	Image          string `mapstructure:"image" env:"IMAGE" desc:"Image speficier to pull the image from a remote registry. This can be empty for local"`
	Branch         string `mapstructure:"branch" env:"BRANCH" desc:"Specify branch that should be pulled for validation. Only needed if remote repo is used"`
	OciPath        string `mapstructure:"ocipath" env:"OCI_PATH" desc:"Specify the location of the oci tar file. Not needed if image is pulled from registry"`
	DockerPath     string `mapstructure:"dockerpath" env:"DOCKER_PATH" desc:"Specify the location of the docker tar file. Not needed if image is pulled from registry"`
	Insecure       bool   `mapstructure:"insecure" env:"INSECURE" desc:"Specify whether the image parsing should be done unsafe (i.e. use http instead of https to communicate with registry)"`
}

type GithubConfig struct {
	PAT      string `mapstructure:"pat" env:"PAT" desc:"Personal access token of account used for creating pr and pushing changes"`
	Username string `mapstructure:"username" env:"USER_NAME" desc:"Username of account used for creating pr and pushing changes"`
}

func ValidateGithub() error {
	if len(viper.GetString("github.pat"))+len(viper.GetString("github.username")) == 0 {
		return nil
	}
	if viper.GetString("github.pat") == "" {
		return errors.New("PAT must be set!")
	}
	if viper.GetString("github.username") == "" {
		return errors.New("Username must be set!")
	}

	return nil
}

type GiteaConfig struct {
	Username    string `mapstructure:"username" env:"USER_NAME" desc:"Username of account used for creating pr and pushing changes"`
	Password    string `mapstructure:"password" env:"PASSWORD" desc:"Password of account used for creating pr and pushing changes"`
	InstanceUrl string `mapstructure:"instance_url" env:"INSTANCE_URL" desc:"URL of the gitea instance"`
}

func ValidateGitea() error {
	if viper.GetString("gitea.password") == "" {
		return errors.New("Password must be set!")
	}
	if viper.GetString("gitea.username") == "" {
		return errors.New("Username must be set!")
	}
	if viper.GetString("gitea.instanceurl") == "" {
		return errors.New("Instanceurl must be set!")
	}
	return nil
}

type Config struct {
	Github     GithubConfig `mapstructure:"github" envPrefix:"GITHUB_"`
	Gitea      GiteaConfig  `mapstructure:"gitea" envPrefix:"GITEA_"`
	Target     TargetConfig `mapstructure:"target" envPrefix:"TARGET_"`
	TargetList string       `mapstructure:"target_list" env:"TARGET_LIST" desc:"List all allowed targets"`
	LogLevel   int          `mapstructure:"log_level" env:"LOG_LEVEL" desc:"Set log level (1-5)"`
	DocsURL    string       `mapstructure:"docs_url" env:"DOCS_URL" desc:"Url pointing to active deployment of policy set documentation"`
	NoFix      bool         `mapstructure:"no_fix" env:"NO_FIX" desc:"Disable the fixing functionality for detected violations"`
}
