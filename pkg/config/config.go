package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const envPrefix = "WHALE_WATCHER_"

var configPath = "./config.yaml"

var lock = &sync.Mutex{}

var config *Config

func init() {
	configPathEnv := os.Getenv(fmt.Sprintf("%sCONFiG_PATH", envPrefix))
	if len(configPathEnv) != 0 {
		configPath = configPathEnv
	}
}

func SetConfigPath(path string) {
	configPath = path
}

func LoadConfigFromData(data []byte) Config {
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		log.Error().Err(err).Msg("Could not parse config")
	}
	return handleEnvOverrides()
}

func loadConfigFromFile(configPath string) Config {
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Warn().Err(err).Msgf("Could not read config file %s", configPath)
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

func GetConfig() Config {
	if config == nil {
		lock.Lock()
		loadedConfig := loadConfigFromFile(configPath)
		config = &loadedConfig
		lock.Unlock()
	}
	return *config
}
