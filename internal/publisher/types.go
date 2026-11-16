// Package publisher handles publishing version check results to external systems.
package publisher

import (
	"context"

	"github.com/JanKrl/homelab-version-agent/internal/checker"
)

// Publisher defines the interface for sending version check results and closing the connection.
type Publisher interface {
	Publish(ctx context.Context, results []checker.Result) error
	Close()
}
