package config

import (
	"errors"
	"strings"
)

type TargetConfig struct {
	RepositoryURL  string `yaml:"repository" env:"REPOSITORY_URL"`
	DockerfilePath string `yaml:"dockerfile" env:"DOCKERFILE"`
	Image          string `yaml:"image" env:"IMAGE"`
	Branch         string `yaml:"branch" env:"BRANCH"`
	OciPath        string `yaml:"ocipath" env:"OCI_PATH"`
	DockerPath     string `yaml:"dockerpath" env:"DOCKER_PATH"`
	Insecure       bool   `yaml:"insecure" env:"INSECURE"`
}

func (tc *TargetConfig) Validate() error {
	return nil
}

type GithubConfig struct {
	PAT      string `yaml:"pat" env:"PAT"`
	Username string `yaml:"username" env:"USER_NAME"`
}

func (gc *GithubConfig) Validate() error {
	if len(gc.PAT)+len(gc.Username) == 0 {
		return nil
	}
	if gc.PAT == "" {
		return errors.New("PAT must be set!")
	}
	if gc.Username == "" {
		return errors.New("Username must be set!")
	}

	return nil
}

type GiteaConfig struct {
	Username    string `yaml:"username" env:"USER_NAME"`
	Password    string `yaml:"password" env:"PASSWORD"`
	InstanceUrl string `yaml:"instance_url" env:"INSTANCE_URL"`
}

func (gc *GiteaConfig) Validate() error {
	if gc.Password == "" {
		return errors.New("PAT must be set!")
	}
	if gc.Username == "" {
		return errors.New("Username must be set!")
	}
	if gc.InstanceUrl == "" {
		return errors.New("Instanceurl must be set!")
	}
	return nil
}

type BaseImageCacheConfig struct {
	BaseImages    []string `yaml:"base_images" env:"BASE_IMAGES"`
	CacheLocation string   `yaml:"cache_location" env:"CACHE_LOCATION"`
}

func (bicg *BaseImageCacheConfig) Validate() error {
	if len(bicg.BaseImages) > 0 && len(bicg.CacheLocation) == 0 {
		return errors.New("Cache location must be provided")
	}
	return nil
}

type Config struct {
	Github         GithubConfig         `yaml:"github" envPrefix:"GITHUB_"`
	Gitea          GiteaConfig          `yaml:"gitea" envPrefix:"GITEA_"`
	Target         TargetConfig         `yaml:"target" envPrefix:"TARGET_"`
	BaseImageCache BaseImageCacheConfig `yaml:"base_image_cache" envPrefix:"BASE_IMAGE_CACHE"`
	TargetList     string               `yaml:"target_list" env:"TARGET_LIST"`
	LogLevel       int                  `yaml:"log_level" env:"LOG_LEVEL"`
	DocsURL        string               `yaml:"docs_url" env:"DOCS_URL"`
	NoFix          bool                 `yaml:"no_fix" env:"NO_FIX"`
}

func (c *Config) Validate() error {
	err := c.Github.Validate()
	if err != nil {
		return err
	}
	err = c.Target.Validate()
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) AllowsTarget(target string) bool {
	if len(c.TargetList) == 0 {
		return true
	}
	return strings.Contains(c.TargetList, target)
}
