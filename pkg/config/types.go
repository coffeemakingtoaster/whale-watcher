package config

import "errors"

type TargetConfig struct {
	RepositoryURL  string `yaml:"repository" env:"REPOSITORY_URL"`
	DockerfilePath string `yaml:"dockerfile" env:"DOCKERFILE_PATH"`
	Image          string `yaml:"image" env:"IMAGE"`
	Branch         string `yaml:"branch" env:"BRANCH"`
	OciPath        string `yaml:"ocipath" env:"OCI_PATH"`
}

func (tc *TargetConfig) Validate() error {
	if tc.RepositoryURL == "" && tc.DockerfilePath == "" {
		return errors.New("RepositoryURL and Dockerfilepath must be set!")
	}

	if tc.Image == "" && tc.OciPath == "" {
		return errors.New("Either image identifier or oci path must be set")
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

type Config struct {
	Github GithubConfig `yaml:"github" envPrefix:"GH_"`
	Target TargetConfig `yaml:"target" envPrefix:"TARGET_"`
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
