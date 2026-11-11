package checker

import "context"

type Result struct {
	Name       string
	Current    string
	Latest     string
	IsOutdated bool
}

type Checker interface {
	Check(ctx context.Context) (Result, error)
}
