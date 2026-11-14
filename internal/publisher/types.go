package publisher

import (
	"context"

	"github.com/JanKrl/homelab-version-agent/internal/checker"
)

type Publisher interface {
	Publish(ctx context.Context, results []checker.Result) error
}
