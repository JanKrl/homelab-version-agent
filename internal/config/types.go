// Package config handles loading and validating the agent's YAML configuration
// and required environment variables.
package config

import "time"

// MqttConfig holds MQTT broker connection and topic settings.
type MqttConfig struct {
	Broker   string `yaml:"broker"`
	Port     int    `yaml:"port"`
	Topic    string `yaml:"topic"`
	Username string `yaml:"username"`
	Password string
}

// GitHubBinaryConfig defines a local binary and its corresponding GitHub repository.
type GitHubBinaryConfig struct {
	Owner   string   `yaml:"owner"`
	Name    string   `yaml:"name"`
	Cmd     string   `yaml:"cmd"`
	CmdArgs []string `yaml:"cmd_args"`
}

// TargetConfig lists the APT packages and GitHub binaries to monitor.
type TargetConfig struct {
	AptPackages []string             `yaml:"apt_packages"`
	GitHubRepos []GitHubBinaryConfig `yaml:"github_binaries"`
}

// Config represents the top-level agent configuration.
type Config struct {
	Host        string        `yaml:"host"`
	Mqtt        MqttConfig    `yaml:"mqtt"`
	Targets     TargetConfig  `yaml:"targets"`
	RunInterval time.Duration `yaml:"run_interval"`
}
