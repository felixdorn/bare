package rules

import (
	"fmt"

	"github.com/felixdorn/bare/core/domain/linter"
)

func init() {
	singleIncomingLink := &linter.SiteRule{
		ID:       "single-incoming-link",
		Name:     "Has only one followed internal linking URL",
		Severity: linter.High,
		Category: linter.Links,
		Tag:      linter.Opportunity,
	}
	singleIncomingLink.Check = func(pages []linter.SiteLintInput) []linter.SiteLintResult {
		// Build reverse index: target URL -> set of unique source URLs (followed only)
		incomingLinks := make(map[string]map[string]bool)

		// Also track all known page URLs
		pageURLs := make(map[string]bool)
		for _, page := range pages {
			pageURLs[page.URL] = true
		}

		for _, page := range pages {
			for _, link := range page.InternalLinks {
				// Only count followed links
				if !link.IsFollow {
					continue
				}

				// Only count links to pages we've crawled
				if !pageURLs[link.TargetURL] {
					continue
				}

				if incomingLinks[link.TargetURL] == nil {
					incomingLinks[link.TargetURL] = make(map[string]bool)
				}
				incomingLinks[link.TargetURL][page.URL] = true
			}
		}

		var results []linter.SiteLintResult

		for targetURL, sourceURLs := range incomingLinks {
			if len(sourceURLs) == 1 {
				// Get the single source URL for evidence
				var sourceURL string
				for url := range sourceURLs {
					sourceURL = url
					break
				}

				results = append(results, linter.SiteLintResult{
					URL: targetURL,
					Lints: []linter.Lint{
						singleIncomingLink.Emit(fmt.Sprintf("only linked from %s", sourceURL)),
					},
				})
			}
		}

		return results
	}
	linter.RegisterSiteRule(singleIncomingLink)
}
