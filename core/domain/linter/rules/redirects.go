package rules

import (
	"fmt"

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

	redirectBroken := &linter.Rule{
		ID:       "redirect-broken",
		Name:     "URL redirect broken (4XX or 5XX)",
		Severity: linter.High,
		Category: linter.Redirects,
		Tag:      linter.Issue,
	}
	redirectBroken.Check = func(ctx *linter.Context) []linter.Lint {
		// Only applies if there were redirects
		if len(ctx.RedirectChain) == 0 {
			return nil
		}

		// Check if final status is an error
		if ctx.StatusCode >= 400 {
			return []linter.Lint{redirectBroken.Emit(fmt.Sprintf("redirected to %d", ctx.StatusCode))}
		}
		return nil
	}
	linter.Register(redirectBroken)
}
