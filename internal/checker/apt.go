package checker

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
)

type AptChecker struct {
	packageName string
}

func NewAptChecker(packageName string) *AptChecker {
	return &AptChecker{
		packageName: packageName,
	}
}

func (c *AptChecker) Check(ctx context.Context) (Result, error) {
	cmd := exec.CommandContext(ctx, "apt-cache", "policy", c.packageName)
	cmd.Env = append(os.Environ(), "LANG=C")

	output, err := cmd.Output()
	if err != nil {
		return Result{}, err
	}

	installed, candidate, err := parseAptPolicy(bytes.NewReader(output))
	if err != nil {
		return Result{}, err
	}

	return Result{
		Name:       c.packageName,
		Current:    installed,
		Latest:     candidate,
		IsOutdated: installed != candidate,
	}, nil
}

func parseAptPolicy(r io.Reader) (installed, candidate string, err error) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.ReplaceAll(scanner.Text(), " ", "")

		if after, found := strings.CutPrefix(line, "Installed:"); found {
			if installed != "" {
				return "", "", errors.New("Failed to parse apt policy output: multiple Installed lines found")
			}
			installed = after
		}

		if after, found := strings.CutPrefix(line, "Candidate:"); found {
			if candidate != "" {
				return "", "", errors.New("Failed to parse apt policy output: multiple Candidate lines found")
			}
			candidate = after
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "", err
	}

	if installed == "" || candidate == "" {
		return "", "", errors.New("Failed to parse apt policy output: missing Installed or Candidate version")
	}

	return installed, candidate, nil
}
