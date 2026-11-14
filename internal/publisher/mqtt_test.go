package publisher

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/JanKrl/homelab-version-agent/internal/checker"
)

func TestConstructMessage(t *testing.T) {
	result := checker.Result{
		Name:       "curl",
		Current:    "1.2.3",
		Latest:     "1.2.4",
		IsOutdated: true,
	}

	payloadBytes, err := constructMessage(result)
	if err != nil {
		t.Errorf("Function returned unexpected error: %v", err)
	}
	var data map[string]any
	err = json.Unmarshal(payloadBytes, &data)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
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
