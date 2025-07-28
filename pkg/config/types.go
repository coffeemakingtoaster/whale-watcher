package config

import "errors"

type TargetConfig struct {
	RepositoryURL  string `yaml:"repository" env:"REPOSITORY_URL"`
	DockerfilePath string `yaml:"dockerfile" env:"DOCKERFILE_PATH"`
	Image          string `yaml:"image" env:"IMAGE"`
	Branch         string `yaml:"branch" env:"BRANCH"`
	OciPath        string `yaml:"ocipath" env:"OCI_PATH"`
	DockerPath     string `yaml:"dockerpath" env:"DOCKER_PATH"`
	Insecure       bool   `yaml:"insecure" env:"INSECURE"`
}

func (tc *TargetConfig) Validate() error {
	if tc.RepositoryURL == "" && tc.DockerfilePath == "" {
		return errors.New("RepositoryURL and Dockerfilepath must be set!")
	}

	if tc.Image == "" && (tc.OciPath == "" || tc.DockerPath == "") {
		return errors.New("Either image identifier or tar paths must be set")
	}

	if tc.Image != "" && tc.OciPath != "" {
		return errors.New("Only image identifier OR oci path can be set at a time")
	}

	return nil
}

type GithubConfig struct {
	PAT      string `yaml:"pat" env:"PAT"`
	Username string `yaml:"username" env:"USER_NAME"`
}

func (gc *GithubConfig) Validate() error {
	if gc.PAT == "" {
		return errors.New("PAT must be set!")
	}
	if gc.Username == "" {
		return errors.New("Username must be set!")
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
	Target         TargetConfig         `yaml:"target" envPrefix:"TARGET_"`
	BaseImageCache BaseImageCacheConfig `yaml:"base_image_cache" envPrefix:"BASE_IMAGE_CACHE"`
	TargetList     string               `yaml:"target_list" envPrefix:"TARGET_LIST"`
	LogLevel       int                  `yaml:"log_level" envPrefix:"LOG_LEVEL"`
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
