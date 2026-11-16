// Package checker provides version checkers for local packages and binaries.
package checker

import "context"

// Result represents the outcome of a version check for a single target.
type Result struct {
	Name       string `json:"name"`
	Current    string `json:"current"`
	Latest     string `json:"latest"`
	IsOutdated bool   `json:"is_outdated"`
	ReleaseURL string `json:"release_url,omitempty"`
}

// Checker defines the interface for checking the current and latest version of a target.
type Checker interface {
	Check(ctx context.Context) (Result, error)
}
