package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

var validConfigYAML = `mqtt:
  broker: "test-broker"
  port: 1883
  username: "test-user"
targets:
  apt_packages:
    - "package1"
    - "package2"
run_interval: 11h`

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Config
		wantErr bool
	}{
		{
			name:  "Valid config",
			input: validConfigYAML,
			want: &Config{
				Mqtt: MqttConfig{
					Broker:   "test-broker",
					Port:     1883,
					Username: "test-user",
				},
				Targets: TargetConfig{
					AptPackages: []string{"package1", "package2"},
				},
				RunInterval: 11 * time.Hour,
			},
			wantErr: false,
		},
		{
			name:    "Invalid config",
			input:   `invalid_yaml`,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Empty config",
			input:   ``,
			want:    nil,
			wantErr: true,
		},
		{
			name: "Too short run_interval",
			input: `mqtt:
  broker: "test-broker"
  port: 1883
  username: "test-user"
targets:
  apt_packages:
    - "package1"
    - "package2"
run_interval: 30m`,
			want:    nil,
			wantErr: true,
		},
		{
			name: "Missing fields",
			input: `mqtt:
  broker: "test-broker"
targets:
  apt_packages:
    - "package1"
    - "package2"
run_interval: 10h`,
			want:    nil,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Parse(strings.NewReader(test.input))
			if (err != nil) != test.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !test.wantErr && !reflect.DeepEqual(got, test.want) {
				t.Errorf("Parse() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestLoadEnvs(t *testing.T) {
	tests := []struct {
		name    string
		envs    map[string]string
		want    RequiredEnvVars
		wantErr bool
	}{
		{
			name: "All required envs set",
			envs: map[string]string{
				"MQTT_PASSWORD": "test-password",
			},
			want: RequiredEnvVars{
				mqttPassword: "test-password",
			},
			wantErr: false,
		},
		{
			name: "Missing required envs",
			envs: map[string]string{},
			want: RequiredEnvVars{
				mqttPassword: "",
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for key, value := range test.envs {
				t.Setenv(key, value)
			}
			got, err := LoadEnvs()
			if (err != nil) != test.wantErr {
				t.Errorf("LoadEnvs() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !test.wantErr && !reflect.DeepEqual(got, test.want) {
				t.Errorf("LoadEnvs() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGetConfig(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "config.yaml")
	os.WriteFile(tmpFile, []byte(validConfigYAML), 0644)

	tests := []struct {
		name    string
		path    string
		envs    map[string]string
		want    *Config
		wantErr bool
	}{
		{
			name:    "Empty path",
			path:    ``,
			envs:    map[string]string{},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Path does not exist",
			path:    `/non/existent/path`,
			envs:    map[string]string{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Valid config file",
			path: tmpFile,
			envs: map[string]string{"MQTT_PASSWORD": "test-password"},
			want: &Config{
				Mqtt: MqttConfig{
					Broker:   "test-broker",
					Port:     1883,
					Username: "test-user",
					Password: "test-password",
				},
				Targets: TargetConfig{
					AptPackages: []string{"package1", "package2"},
				},
				RunInterval: 11 * time.Hour,
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for key, value := range test.envs {
				t.Setenv(key, value)
			}

			got, err := GetConfig(test.path)
			if (err != nil) != test.wantErr {
				t.Errorf("GetConfig() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !test.wantErr && !reflect.DeepEqual(got, test.want) {
				t.Errorf("GetConfig() = %v, want %v", got, test.want)
			}

			for key, value := range test.envs {
				if os.Getenv(key) != value {
					t.Errorf("Environment variable %s = %v, want %v", key, os.Getenv(key), value)
				}
			}
		})
	}
}
