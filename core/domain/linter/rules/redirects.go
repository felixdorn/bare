package rules

import (
	"github.com/felixdorn/bare/core/domain/linter"
)

func init() {
	redirectsToSelf := &linter.Rule{
		ID:       "redirects-to-self",
		Name:     "Internal URL redirects back to itself",
		Severity: linter.High,
		Category: linter.Redirects,
		Tag:      linter.Issue,
	}
	redirectsToSelf.Check = func(ctx *linter.Context) []linter.Lint {
		if len(ctx.RedirectChain) == 0 {
			return nil
		}

		pageURL := ctx.URL.String()
		for _, redirect := range ctx.RedirectChain {
			if redirect.URL == pageURL {
				return []linter.Lint{redirectsToSelf.Emit(redirect.URL)}
			}
		}
		return nil
	}
	linter.Register(redirectsToSelf)
}
