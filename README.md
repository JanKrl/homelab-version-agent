# homelab-version-agent

A lightweight Go agent that periodically checks installed software versions on a host and publishes them to an MQTT broker with Home Assistant MQTT Discovery support.

> **Note:** This is a hobby project in early development. Configuration formats, MQTT payloads, and internal APIs may change.

## Features

- **APT package checks** via `apt-cache policy` (compares installed vs. candidate version)
- **GitHub release checks** for local binaries (runs a local version command and compares against the latest GitHub release tag)
- **MQTT & Home Assistant Discovery** (publishes `update` entities, availability status, and version state)

## Requirements

- Go 1.26+ (to build from source)
- `apt-cache` (if monitoring APT packages)
- An MQTT broker (e.g., Mosquitto / Home Assistant)

## Build

```bash
go build -o agent ./cmd/agent
```

## Configuration

By default, the agent reads `/etc/homelab-agent/config.yaml`. You can override this with the `-config` flag.

Example `config.yaml`:

```yaml
host: "my-host"

mqtt:
  broker: "192.168.1.4"
  port: 1883
  topic: "homelab/versions"
  username: "homelab-version-agent"

run_interval: 10h # Minimum allowed interval is 1h

targets:
  apt_packages:
    - "curl"
  github_binaries:
    - owner: "cli"
      name: "cli"
      cmd: "gh"
      cmd_args: ["--version"]
```

### Environment Variables

- `MQTT_PASSWORD` (required): password for the MQTT user.

## Usage

```bash
MQTT_PASSWORD="your-password" ./agent -config ./config.yaml
```

Available flags:

- `-config` — path to the YAML configuration file (default: `/etc/homelab-agent/config.yaml`)
- `-version` — print the agent version and exit

## Development

Run the test suite:

```bash
go test ./...
```
