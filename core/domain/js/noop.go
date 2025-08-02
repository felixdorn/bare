package js

import (
	"context"

	"github.com/felixdorn/bare/core/domain/url"
)

// Noop is a no-op crawler that does nothing.
type Noop struct{}

// NewNoop creates a new Noop crawler.
func NewNoop() *Noop {
	return &Noop{}
}

// Run does nothing and returns nil.
func (n *Noop) Run(ctx context.Context) ([]*url.URL, error) {
	return nil, nil
}
