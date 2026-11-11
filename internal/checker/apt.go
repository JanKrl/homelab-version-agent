package checker

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

type AptChecker struct {
	packageName string
}

func (c *AptChecker) Check(r io.Reader) (Result, error) {
	// TODO
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
