package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseGithubRelease(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name: "Real output, valid tag",
			input: `{
  "url": "https://api.github.com/repos/JanKrl/ha-kospel-cmi/releases/441297496",
  "assets_url": "https://api.github.com/repos/JanKrl/ha-kospel-cmi/releases/441297496/assets",
  "upload_url": "https://uploads.github.com/repos/JanKrl/ha-kospel-cmi/releases/441297496/assets{?name,label}",
  "html_url": "https://github.com/JanKrl/ha-kospel-cmi/releases/tag/v1.6.0",
  "id": 441297496,
  "author": {
    "login": "JanKrl",
    "id": 89135865,
    "node_id": "MDQ6VXNlcjg5MTM1ODY1",
    "avatar_url": "https://avatars.githubusercontent.com/u/89135865?v=4",
    "gravatar_id": "",
    "url": "https://api.github.com/users/JanKrl",
    "html_url": "https://github.com/JanKrl",
    "followers_url": "https://api.github.com/users/JanKrl/followers",
    "following_url": "https://api.github.com/users/JanKrl/following{/other_user}",
    "gists_url": "https://api.github.com/users/JanKrl/gists{/gist_id}",
    "starred_url": "https://api.github.com/users/JanKrl/starred{/owner}{/repo}",
    "subscriptions_url": "https://api.github.com/users/JanKrl/subscriptions",
    "organizations_url": "https://api.github.com/users/JanKrl/orgs",
    "repos_url": "https://api.github.com/users/JanKrl/repos",
    "events_url": "https://api.github.com/users/JanKrl/events{/privacy}",
    "received_events_url": "https://api.github.com/users/JanKrl/received_events",
    "type": "User",
    "user_view_type": "public",
    "site_admin": false
  },
  "node_id": "RE_kwDORLhZgc4aTapY",
  "tag_name": "v1.6.0",
  "target_commitish": "9b80c3607e0d2cb8566bdf34ee6a30d0098cbe94",
  "name": "v1.6.0",
  "draft": false,
  "immutable": false,
  "prerelease": false,
  "created_at": "2026-11-11T06:07:07Z",
  "updated_at": "2026-11-11T06:19:04Z",
  "published_at": "2026-11-11T06:07:17Z",
  "assets": [

  ],
  "tarball_url": "https://api.github.com/repos/JanKrl/ha-kospel-cmi/tarball/v1.6.0",
  "zipball_url": "https://api.github.com/repos/JanKrl/ha-kospel-cmi/zipball/v1.6.0",
  "body": "## [1.6.0](https://github.com/JanKrl/ha-kospel-cmi/compare/v1.5.0...v1.6.0) (2026-11-11)\n\n\n### Features\n\n* **ha:** add Kospel PPE.4 device support and multi-device architecture ([#156](https://github.com/JanKrl/ha-kospel-cmi/issues/156)) ([0005048](https://github.com/JanKrl/ha-kospel-cmi/commit/000504835f9a4989ead68775357f83788055c842))\n\n\n### Documentation\n\n* document release-please monorepo pitfalls and release protocol ([#150](https://github.com/JanKrl/ha-kospel-cmi/issues/150)) ([d2ca5c9](https://github.com/JanKrl/ha-kospel-cmi/commit/d2ca5c9fb52ed2f7664886dc69d28175a9275a00))\n\n\n---\n☕ **Enjoying this integration?**\nI maintain this open-source integration in my private time. If you'd like to say thanks with a coffee:\n\n[![Buy me a coffee](https://img.shields.io/badge/Kup%20mi%20kaw%C4%99-buycoffee.to-ffdd00.svg?style=for-the-badge&logo=buy-me-a-coffee&logoColor=black)](https://buycoffee.to/jankrl)\n\n*GitHub issues can be created in either Polish or English.*"
}`,
			want:    `v1.6.0`,
			wantErr: false,
		},
		{
			name: "Invalid JSON",
			input: `{
  "url": "https://api.github.com/repos/JanKrl/ha-kospel-cmi/releases/441297496",
  "assets_url": "https://api.github.com/repos/JanKrl/ha-kospel-cmi/releases/441297496/assets",
  "upload_url": "https://uploads.github.com/repos/JanKrl/ha-kospel-cmi/releases/441297496/assets{?name,label}",
  "html_url": "`,
			want:    "",
			wantErr: true,
		},
		{
			name: "Missing tag_name",
			input: `{
  "url": "https://api.github.com/repos/JanKrl/ha-kospel-cmi/releases/441297496",
  "assets_url": "https://api.github.com/repos/JanKrl/ha-kospel-cmi/releases/441297496/assets",
  "upload_url": "https://uploads.github.com/repos/JanKrl/ha-kospel-cmi/releases/441297496/assets{?name,label}"
}`,
			want:    "",
			wantErr: true,
		},
		{
			name:    "Empty input",
			input:   ``,
			want:    "",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseGitHubRelease(strings.NewReader(test.input))
			if (err != nil) != test.wantErr {
				t.Errorf("parseGitHubRelease() error = %v, wantErr %v, got >>%v<<", err, test.wantErr, got)
				return
			}
			if got != test.want {
				t.Errorf("parseGitHubRelease() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGetLatestGitHubRelease(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo/releases/latest" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"tag_name": "v1.2.3"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	t.Run("Valid response", func(t *testing.T) {
		client := &http.Client{}
		tag, err := getLatestGitHubRelease(context.Background(), client, "owner", "repo", ts.URL)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if tag != "v1.2.3" {
			t.Errorf("Expected tag 'v1.2.3', got %v", tag)
		}
	})
}

func TestParseCmdVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "Standard --version output",
			input:   "curl 8.5.0 (x86_64-pc-linux-gnu) libcurl/8.5.0",
			want:    "8.5.0",
			wantErr: false,
		},
		{
			name:    "v-prefixed version",
			input:   "pocket-id version v1.2.3\n",
			want:    "v1.2.3",
			wantErr: false,
		},
		{
			name:    "Filebrowser version output",
			input:   "File Browser v2.31.2/abcdef",
			want:    "v2.31.2",
			wantErr: false,
		},
		{
			name:    "No version in output",
			input:   "unknown output",
			want:    "",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseCmdVersion(test.input)
			if (err != nil) != test.wantErr {
				t.Errorf("parseCmdVersion() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if got != test.want {
				t.Errorf("parseCmdVersion() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGetCurrentVersion(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "fake-cli")
	script := `#!/bin/sh
if [ "$1" = "--version" ]; then
  echo "fake-cli v1.2.3"
  exit 0
fi
if [ "$1" = "version" ] && [ "$2" = "--short" ]; then
  echo "v2.0.1"
  exit 0
fi
if [ "$1" = "version" ]; then
  echo "fake-cli version v1.2.3"
  exit 0
fi
if [ "$1" = "-V" ]; then
  echo "fake-cli v3.4.5" >&2
  exit 0
fi
echo "unknown flag: $1" >&2
exit 1
`
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write fake cli script: %v", err)
	}

	tests := []struct {
		name    string
		cmd     string
		cmdArgs []string
		want    string
		wantErr bool
	}{
		{
			name:    "Standard --version flag",
			cmd:     scriptPath,
			cmdArgs: []string{"--version"},
			want:    "v1.2.3",
			wantErr: false,
		},
		{
			name:    "Subcommand version",
			cmd:     scriptPath,
			cmdArgs: []string{"version"},
			want:    "v1.2.3",
			wantErr: false,
		},
		{
			name:    "Multi-arg cmd_args",
			cmd:     scriptPath,
			cmdArgs: []string{"version", "--short"},
			want:    "v2.0.1",
			wantErr: false,
		},
		{
			name:    "Version printed to stderr",
			cmd:     scriptPath,
			cmdArgs: []string{"-V"},
			want:    "v3.4.5",
			wantErr: false,
		},
		{
			name:    "Command exits with error",
			cmd:     scriptPath,
			cmdArgs: []string{"--invalid"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "Non-existent command",
			cmd:     filepath.Join(tmpDir, "does-not-exist"),
			cmdArgs: []string{"--version"},
			want:    "",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := getCurrentVersion(context.Background(), test.cmd, test.cmdArgs)
			if (err != nil) != test.wantErr {
				t.Errorf("getCurrentVersion() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if got != test.want {
				t.Errorf("getCurrentVersion() = %v, want %v", got, test.want)
			}
		})
	}
}
