package checker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// GitHubChecker compares a local binary's version against the latest GitHub release.
type GitHubChecker struct {
	owner   string
	name    string
	cmd     string
	cmdArgs []string
	client  *http.Client
}

// NewGitHubChecker creates a GitHubChecker for the specified repository and local command.
// If client is nil, a default HTTP client with a 10-second timeout is used.
func NewGitHubChecker(owner, name, cmd string, cmdArgs []string, client *http.Client) *GitHubChecker {
	if client == nil {
		client = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	return &GitHubChecker{
		owner:   owner,
		name:    name,
		cmd:     cmd,
		cmdArgs: cmdArgs,
		client:  client,
	}
}

type githubApiBody struct {
	TagName string `json:"tag_name"`
}

// Check fetches the latest GitHub release tag and compares it with the local binary's version.
func (c *GitHubChecker) Check(ctx context.Context) (Result, error) {
	latest, err := getLatestGitHubRelease(ctx, c.client, c.owner, c.name, "https://api.github.com")
	if err != nil {
		return Result{}, err
	}

	current, err := getCurrentVersion(ctx, c.cmd, c.cmdArgs)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Name:       c.owner + "/" + c.name,
		Current:    current,
		Latest:     latest,
		IsOutdated: current != latest,
		ReleaseURL: fmt.Sprintf("https://github.com/%s/%s/releases/tag/%s", c.owner, c.name, latest),
	}, nil
}

func getCurrentVersion(ctx context.Context, cmd string, cmdArgs []string) (string, error) {
	command := exec.CommandContext(ctx, cmd, cmdArgs...)
	command.Env = append(os.Environ(), "LANG=C")

	output, err := command.CombinedOutput()
	if err != nil {
		return "", err
	}

	tag, err := parseCmdVersion(string(output))
	if err != nil {
		return "", err
	}

	return tag, nil
}

func getLatestGitHubRelease(ctx context.Context, client *http.Client, owner, repo, baseUrl string) (string, error) {
	baseUrl = strings.TrimSuffix(baseUrl, "/")
	apiEndpoint := fmt.Sprintf("%v/repos/%v/%v/releases/latest", baseUrl, owner, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", apiEndpoint, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("Failed to fetch latest release: " + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	latest, err := parseGitHubRelease(bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	return latest, nil
}

func parseGitHubRelease(body io.Reader) (tag string, err error) {
	var parsedBody githubApiBody
	err = json.NewDecoder(body).Decode(&parsedBody)
	tag = parsedBody.TagName
	if err != nil {
		return "", err
	}
	if tag == "" {
		return "", errors.New("Failed to parse latest release: missing tag_name")
	}
	return tag, nil
}

func parseCmdVersion(version string) (tag string, err error) {
	tag = regexp.MustCompile(`v?(\d+\.\d+\.\d+)`).FindString(version)

	if tag == "" {
		return "", errors.New("Failed to parse current version from output: " + version)
	}

	return tag, nil
}
