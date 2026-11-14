package config

import "time"

type MqttConfig struct {
	Broker   string `yaml:"broker"`
	Port     int    `yaml:"port"`
	Topic    string `yaml:"topic"`
	Username string `yaml:"username"`
	Password string
}

type GitHubBinaryConfig struct {
	Owner string `yaml:"owner"`
	Name  string `yaml:"name"`
	Cmd   string `yaml:"cmd"`
}

type TargetConfig struct {
	AptPackages []string             `yaml:"apt_packages"`
	GitHubRepos []GitHubBinaryConfig `yaml:"github_binaries"`
}

type Config struct {
	Mqtt        MqttConfig    `yaml:"mqtt"`
	Targets     TargetConfig  `yaml:"targets"`
	RunInterval time.Duration `yaml:"run_interval"`
}
