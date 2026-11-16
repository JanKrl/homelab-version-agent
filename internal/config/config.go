package config

import (
	"errors"
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// GetConfig loads and validates configuration from the YAML file at path and
// populates secrets from required environment variables.
func GetConfig(path string) (*Config, error) {
	if path == "" {
		return nil, errors.New("path is empty")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config, err := Parse(file)
	if err != nil {
		return nil, err
	}

	envVars, err := LoadEnvs()
	if err != nil {
		return nil, err
	}

	config.Mqtt.Password = envVars.mqttPassword

	return config, nil
}

// ValidateConfig checks that all required fields in cfg are present and valid.
func ValidateConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("config is nil")
	}
	if cfg.Host == "" {
		return errors.New("host is empty")
	}
	if cfg.Mqtt.Broker == "" {
		return errors.New("mqtt.broker is empty")
	}
	if cfg.Mqtt.Port <= 0 {
		return errors.New("mqtt.port must be greater than 0")
	}
	if cfg.Mqtt.Topic == "" {
		return errors.New("mqtt.topic is empty")
	}
	if cfg.Mqtt.Username == "" {
		return errors.New("mqtt.username is empty")
	}
	if len(cfg.Targets.AptPackages) == 0 && len(cfg.Targets.GitHubRepos) == 0 {
		return errors.New("targets must contain at least one apt package or GitHub repo")
	}
	if cfg.RunInterval < time.Hour*1 {
		return errors.New("run_interval must be greater than 1 hour")
	}

	for _, gitHubTarget := range cfg.Targets.GitHubRepos {
		if gitHubTarget.Owner == "" {
			return errors.New("One or more GitHub targets lacks `owner`")
		}
		if gitHubTarget.Name == "" {
			return errors.New("One or more GitHub targets lacks `name`")
		}
		if gitHubTarget.Cmd == "" {
			return errors.New("One or more GitHub targets lacks `cmd`")
		}
		if len(gitHubTarget.CmdArgs) == 0 {
			return errors.New("One or more GitHub targets lacks `cmd_args`")
		}
	}

	return nil
}

// Parse decodes YAML configuration from r and validates it.
func Parse(r io.Reader) (*Config, error) {
	if r == nil {
		return nil, errors.New("reader is nil")
	}

	var cfg Config
	err := yaml.NewDecoder(r).Decode(&cfg)
	if err != nil {
		return nil, err
	}

	if err := ValidateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// RequiredEnvVars holds configuration values loaded from environment variables.
type RequiredEnvVars struct {
	mqttPassword string
}

// LoadEnvs reads required environment variables, such as MQTT_PASSWORD.
func LoadEnvs() (RequiredEnvVars, error) {
	mqttPassword := os.Getenv("MQTT_PASSWORD")
	if mqttPassword == "" {
		return RequiredEnvVars{}, errors.New("MQTT_PASSWORD environment variable is not set")
	}

	return RequiredEnvVars{
		mqttPassword: mqttPassword,
	}, nil
}
