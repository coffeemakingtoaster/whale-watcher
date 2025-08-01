package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const envPrefix = "WHALE_WATCHER_"

var configPath = "./config.yaml"

var lock = &sync.Mutex{}

var config *Config

func init() {
	configPathEnv := os.Getenv(fmt.Sprintf("%sCONFIG_PATH", envPrefix))
	if len(configPathEnv) != 0 {
		log.Warn().Msgf("Config path specified in env! Updating to %s", configPathEnv)
		SetConfigPath(configPathEnv)
	}
	GetConfig()
	zerolog.SetGlobalLevel(zerolog.Level(config.LogLevel))
}

func SetConfigPath(path string) {
	configPath = path
}

func LoadConfigFromData(data []byte) Config {
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		log.Error().Err(err).Msg("Could not parse config, initialising empty and trusting env fallback")
	}
	return handleEnvOverrides()
}

func loadConfigFromFile(configPath string) Config {
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Warn().Err(err).Msgf("Could not read config file %s", configPath)
		config = &Config{}
		return handleEnvOverrides()
	}
	return LoadConfigFromData(data)
}

func handleEnvOverrides() Config {
	err := env.ParseWithOptions(config, env.Options{Prefix: envPrefix})
	if err != nil {
		log.Warn().Err(err).Msg("Failed to read env variables for config overrides")
	}
	return *config
}

func ShouldInteractWithVSC() bool {
	cfg := GetConfig()
	return len(cfg.Github.PAT) > 0 && len(cfg.Github.Username) > 0
}

func GetConfig() Config {
	if config == nil {
		lock.Lock()
		loadedConfig := loadConfigFromFile(configPath)
		config = &loadedConfig
		err := config.Validate()
		if err != nil {
			log.Error().Err(err).Msg("Invalid config!")
		}
		lock.Unlock()
	}
	return *config
}
