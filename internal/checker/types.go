package checker

import "context"

type Result struct {
	Name       string `json:"name"`
	Current    string `json:"current"`
	Latest     string `json:"latest"`
	IsOutdated bool   `json:"is_outdated"`
}

type Checker interface {
	Check(ctx context.Context) (Result, error)
}
