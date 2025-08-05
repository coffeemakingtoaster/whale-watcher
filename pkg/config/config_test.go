package config_test

import (
	"reflect"
	"testing"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/config"
)

var yamlConfig = `
  github:
    pat: hello world
    username: test
`

func TestNewYamlConfig(t *testing.T) {
	expected := config.Config{
		Github: config.GithubConfig{
			PAT:      "hello world",
			Username: "test",
		},
	}

	actual := config.LoadConfigFromData([]byte(yamlConfig))
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("config mismatch: Expected %v Got %v", expected, actual)
	}
}

func TestNewYamlFromENVConfig(t *testing.T) {
	expected := config.Config{
		Github: config.GithubConfig{
			PAT:      "env",
			Username: "envtest",
		},
	}

	t.Setenv("WHALE_WATCHER_GITHUB_PAT", "env")
	t.Setenv("WHALE_WATCHER_GITHUB_USER_NAME", "envtest")

	actual := config.LoadConfigFromData([]byte{})
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("config mismatch: Expected %v Got %v", expected, actual)
	}
}

func TestNewYamlConfigEnvOverride(t *testing.T) {
	expected := config.Config{
		Github: config.GithubConfig{
			PAT:      "byebye world",
			Username: "test",
		},
	}

	t.Setenv("WHALE_WATCHER_GITHUB_PAT", "byebye world")

	actual := config.LoadConfigFromData([]byte(yamlConfig))
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("config mismatch: Expected %v Got %v", expected, actual)
	}
}

func TestConfigDisallowedCMD(t *testing.T) {
	cfg := config.Config{
		TargetList: "os,fs",
	}
	if cfg.AllowsTarget("cmd") {
		t.Errorf("config mismatch: Expected cmd target to be disallowed but was allowed")
	}
	if !cfg.AllowsTarget("fs") {
		t.Errorf("config mismatch: Expected fs target to be allowed but was disallowed")
	}
	if !cfg.AllowsTarget("os") {
		t.Errorf("config mismatch: Expected os target to be allowed but was disallowed")
	}
}

func TestConfigAllAllowedImpl(t *testing.T) {
	cfg := config.Config{
		TargetList: "",
	}
	if !cfg.AllowsTarget("cmd") {
		t.Errorf("config mismatch: Expected cmd target to be allowed but was disallowed")
	}
	if !cfg.AllowsTarget("fs") {
		t.Errorf("config mismatch: Expected fs target to be allowed but was disallowed")
	}
	if !cfg.AllowsTarget("os") {
		t.Errorf("config mismatch: Expected os target to be allowed but was disallowed")
	}
}

func TestConfigAllAllowedExpl(t *testing.T) {
	cfg := config.Config{
		TargetList: "cmd,os,fs",
	}
	if !cfg.AllowsTarget("cmd") {
		t.Errorf("config mismatch: Expected cmd target to be allowed but was disallowed")
	}
	if !cfg.AllowsTarget("fs") {
		t.Errorf("config mismatch: Expected fs target to be allowed but was disallowed")
	}
	if !cfg.AllowsTarget("os") {
		t.Errorf("config mismatch: Expected os target to be allowed but was disallowed")
	}
}

func TestConfigExplEnv(t *testing.T) {
	t.Setenv("WHALE_WATCHER_TARGET_LIST", "os,cmd")
	cfg := config.LoadConfigFromData([]byte(yamlConfig))
	if !cfg.AllowsTarget("cmd") {
		t.Errorf("config mismatch: Expected cmd target to be allowed but was disallowed")
	}
	if cfg.AllowsTarget("fs") {
		t.Errorf("config mismatch: Expected fs target to be disallowed but was allowed")
	}
	if !cfg.AllowsTarget("os") {
		t.Errorf("config mismatch: Expected os target to be allowed but was disallowed")
	}
}
