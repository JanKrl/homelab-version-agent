package publisher

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/JanKrl/homelab-version-agent/internal/checker"
	"github.com/JanKrl/homelab-version-agent/internal/config"
)

func TestConstructMessage(t *testing.T) {
	result := checker.Result{
		Name:       "curl",
		Current:    "1.2.3",
		Latest:     "1.2.4",
		IsOutdated: true,
	}

	payloadBytes, err := constructMessage(result, "test-host")
	if err != nil {
		t.Errorf("Function returned unexpected error: %v", err)
	}
	var data map[string]any
	err = json.Unmarshal(payloadBytes, &data)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if data["host"] != "test-host" {
		t.Errorf("Invalid host, expected 'test-host', got '%v'", data["host"])
	}
	if data["name"] != "curl" {
		t.Errorf("Invalid name, expected 'curl', got '%v'", data["name"])
		return
	}
	if data["current"] != "1.2.3" {
		t.Errorf("Invalid current, expected '1.2.3', got '%v'", data["current"])
		return
	}
	if data["latest"] != "1.2.4" {
		t.Errorf("Invalid latest, expected '1.2.4', got '%v'", data["latest"])
		return
	}
	if data["is_outdated"] != true {
		t.Errorf("Invalid is_outdated, expected 'true', got '%v'", data["is_outdated"])
		return
	}

	timeStr, ok := data["last_checked_at"].(string)
	_, err = time.Parse(time.RFC3339, timeStr)
	if !ok || (err != nil) {
		t.Errorf("Invalid last_checked_at, received '%v', ok '%v', err '%v'", data["last_checked_at"], ok, err)
		return
	}
}

func TestSanitizeID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "curl", want: "curl"},
		{input: "pocket-id/pocket-id", want: "pocket-id__pocket-id"},
		{input: "gtsteffaniak/filebrowser", want: "gtsteffaniak__filebrowser"},
		{input: "libssl3.0", want: "libssl3__0"},
		{input: "my app", want: "my__app"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if got := sanitizeID(test.input); got != test.want {
				t.Errorf("sanitizeID(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestAvailabilityTopic(t *testing.T) {
	got := availabilityTopic("homelab/versions", "my-lxc")
	want := "homelab/versions/my-lxc/status"
	if got != want {
		t.Errorf("availabilityTopic() = %q, want %q", got, want)
	}
}

func TestConstructDiscoveryMessage(t *testing.T) {
	t.Run("GitHub binary with release_url", func(t *testing.T) {
		result := checker.Result{
			Name:       "pocket-id/pocket-id",
			Current:    "v1.2.3",
			Latest:     "v1.3.0",
			IsOutdated: true,
			ReleaseURL: "https://github.com/pocket-id/pocket-id/releases/tag/v1.3.0",
		}

		topic, payloadBytes, err := constructDiscoveryMessage(result, "homelab/versions", "my-lxc")
		if err != nil {
			t.Fatalf("constructDiscoveryMessage() returned unexpected error: %v", err)
		}

		wantTopic := "homeassistant/update/my-lxc/pocket-id__pocket-id/config"
		if topic != wantTopic {
			t.Errorf("topic = %q, want %q", topic, wantTopic)
		}

		var data map[string]any
		if err := json.Unmarshal(payloadBytes, &data); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		expectedFields := map[string]string{
			"name":                    "pocket-id/pocket-id",
			"unique_id":               "my-lxc___pocket-id__pocket-id",
			"state_topic":             "homelab/versions/my-lxc/pocket-id/pocket-id",
			"value_template":          "{{ value_json.current }}",
			"latest_version_topic":    "homelab/versions/my-lxc/pocket-id/pocket-id",
			"latest_version_template": "{{ value_json.latest }}",
			"release_url":             "https://github.com/pocket-id/pocket-id/releases/tag/v1.3.0",
			"availability_topic":      "homelab/versions/my-lxc/status",
			"payload_available":       "online",
			"payload_not_available":   "offline",
		}

		for key, want := range expectedFields {
			if got, ok := data[key].(string); !ok || got != want {
				t.Errorf("data[%q] = %v, want %q", key, data[key], want)
			}
		}

		device, ok := data["device"].(map[string]any)
		if !ok {
			t.Fatalf("data[\"device\"] is missing or not an object: %T", data["device"])
		}

		if device["name"] != "LXC my-lxc" {
			t.Errorf("device[\"name\"] = %v, want %q", device["name"], "LXC my-lxc")
		}
		if device["model"] != "Proxmox LXC Container" {
			t.Errorf("device[\"model\"] = %v, want %q", device["model"], "Proxmox LXC Container")
		}
		if device["manufacturer"] != "Homelab" {
			t.Errorf("device[\"manufacturer\"] = %v, want %q", device["manufacturer"], "Homelab")
		}

		wantIdentifiers := []any{"homelab_lxc_my-lxc"}
		if !reflect.DeepEqual(device["identifiers"], wantIdentifiers) {
			t.Errorf("device[\"identifiers\"] = %v, want %v", device["identifiers"], wantIdentifiers)
		}
	})

	t.Run("APT package omits empty release_url", func(t *testing.T) {
		result := checker.Result{
			Name:       "curl",
			Current:    "8.5.0",
			Latest:     "8.5.0",
			IsOutdated: false,
		}

		topic, payloadBytes, err := constructDiscoveryMessage(result, "homelab/versions", "my-lxc")
		if err != nil {
			t.Fatalf("constructDiscoveryMessage() returned unexpected error: %v", err)
		}

		wantTopic := "homeassistant/update/my-lxc/curl/config"
		if topic != wantTopic {
			t.Errorf("topic = %q, want %q", topic, wantTopic)
		}

		var data map[string]any
		if err := json.Unmarshal(payloadBytes, &data); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if _, exists := data["release_url"]; exists {
			t.Errorf("expected release_url to be omitted when empty, got %v", data["release_url"])
		}
		if data["unique_id"] != "my-lxc___curl" {
			t.Errorf("unique_id = %v, want %q", data["unique_id"], "my-lxc___curl")
		}
	})
}

func TestBuildClientOptions(t *testing.T) {
	mqttCfg := config.MqttConfig{
		Broker:   "192.168.1.4",
		Port:     1883,
		Topic:    "homelab/versions",
		Username: "agent",
		Password: "secret",
	}

	opts := buildClientOptions(mqttCfg, "my-lxc")

	if !opts.AutoReconnect {
		t.Error("expected AutoReconnect to be true")
	}
	if !opts.WillEnabled {
		t.Fatal("expected WillEnabled to be true")
	}
	if opts.WillTopic != "homelab/versions/my-lxc/status" {
		t.Errorf("WillTopic = %q, want %q", opts.WillTopic, "homelab/versions/my-lxc/status")
	}
	if string(opts.WillPayload) != "offline" {
		t.Errorf("WillPayload = %q, want %q", string(opts.WillPayload), "offline")
	}
	if opts.WillQos != 1 {
		t.Errorf("WillQos = %d, want 1", opts.WillQos)
	}
	if !opts.WillRetained {
		t.Error("expected WillRetained to be true")
	}
	if opts.OnConnect == nil {
		t.Error("expected OnConnect handler to be set for publishing online status")
	}
}
