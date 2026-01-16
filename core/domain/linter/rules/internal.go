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

	// Site rule: Internal URL failed to crawl (timeout or other error)
	brokenCrawlError := &linter.SiteRule{
		ID:       "broken-internal-url-crawl-error",
		Name:     "Internal URL failed to crawl",
		Severity: linter.High,
		Category: linter.Internal,
		Tag:      linter.Issue,
	}
	brokenCrawlError.Check = func(pages []linter.SiteLintInput) []linter.SiteLintResult {
		var results []linter.SiteLintResult

		for _, page := range pages {
			if page.CrawlError != "" {
				evidence := "Crawl error"
				if page.IsTimeout {
					evidence = "Request timed out"
				}
				results = append(results, linter.SiteLintResult{
					URL:   page.URL,
					Lints: []linter.Lint{brokenCrawlError.Emit(evidence)},
				})
			}
		}

		return results
	}
	linter.RegisterSiteRule(brokenCrawlError)
}
