package rules

import (
	"fmt"

	"github.com/felixdorn/bare/core/domain/linter"
)

func init() {
	// Rule: 5XX URL in XML Sitemap
	serverErrorInSitemap := &linter.SiteRule{
		ID:       "sitemap-has-5xx-url",
		Name:     "URL in XML sitemap returns server error (5XX)",
		Severity: linter.Critical,
		Category: linter.XMLSitemaps,
		Tag:      linter.Issue,
	}
	serverErrorInSitemap.Check = func(pages []linter.SiteLintInput) []linter.SiteLintResult {
		var results []linter.SiteLintResult

		for _, page := range pages {
			if page.InSitemap && page.StatusCode >= 500 && page.StatusCode < 600 {
				results = append(results, linter.SiteLintResult{
					URL: page.URL,
					Lints: []linter.Lint{
						serverErrorInSitemap.Emit(fmt.Sprintf("HTTP %d", page.StatusCode)),
					},
				})
			}
		}

		return results
	}
	linter.RegisterSiteRule(serverErrorInSitemap)
}
