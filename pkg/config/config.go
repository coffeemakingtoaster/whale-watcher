package config

import (
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const envPrefix = "WHALE_WATCHER_"

func LoadConfigFromData(data []byte) *Config {
	var config Config
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		log.Error().Err(err).Msg("Could not parse config")
	}
	return handleEnvOverrides(&config)
}

func LoadConfigFromFile(configPath string) *Config {
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Warn().Err(err).Msgf("Could not read config file %s", configPath)
		var config Config
		return handleEnvOverrides(&config)
	}
	return LoadConfigFromData(data)
}

func handleEnvOverrides(config *Config) *Config {
	err := env.ParseWithOptions(config, env.Options{Prefix: envPrefix})
	if err != nil {
		log.Warn().Err(err).Msg("Failed to read env variables for config overrides")
	}
	return config
}
