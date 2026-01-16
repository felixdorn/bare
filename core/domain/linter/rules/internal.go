package rules

import (
	"fmt"

	"github.com/felixdorn/bare/core/domain/linter"
)

func init() {
	brokenURL := &linter.Rule{
		ID:       "broken-internal-url",
		Name:     "Broken internal URL",
		Severity: linter.High,
		Category: linter.Internal,
		Tag:      linter.Issue,
	}
	brokenURL.Check = func(ctx *linter.Context) []linter.Lint {
		// A URL is broken if it returns 4XX or 5XX status
		if ctx.StatusCode >= 400 {
			return []linter.Lint{brokenURL.Emit(fmt.Sprintf("HTTP %d", ctx.StatusCode))}
		}
		return nil
	}
	linter.Register(brokenURL)
}
