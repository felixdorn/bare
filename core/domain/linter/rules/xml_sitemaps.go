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

	// Rule: Noindex URL in XML Sitemap
	noindexInSitemap := &linter.SiteRule{
		ID:       "sitemap-has-noindex-url",
		Name:     "URL in XML sitemap is noindex",
		Severity: linter.Critical,
		Category: linter.XMLSitemaps,
		Tag:      linter.Issue,
	}
	noindexInSitemap.Check = func(pages []linter.SiteLintInput) []linter.SiteLintResult {
		var results []linter.SiteLintResult

		for _, page := range pages {
			if page.InSitemap && page.IsNoindex {
				results = append(results, linter.SiteLintResult{
					URL:   page.URL,
					Lints: []linter.Lint{noindexInSitemap.Emit("")},
				})
			}
		}

		return results
	}
	linter.RegisterSiteRule(noindexInSitemap)

	// Rule: 4XX URL in XML Sitemap
	notFoundInSitemap := &linter.SiteRule{
		ID:       "sitemap-has-4xx-url",
		Name:     "URL in XML sitemap returns not found (4XX)",
		Severity: linter.Critical,
		Category: linter.XMLSitemaps,
		Tag:      linter.Issue,
	}
	notFoundInSitemap.Check = func(pages []linter.SiteLintInput) []linter.SiteLintResult {
		var results []linter.SiteLintResult

		for _, page := range pages {
			if page.InSitemap && page.StatusCode >= 400 && page.StatusCode < 500 {
				results = append(results, linter.SiteLintResult{
					URL: page.URL,
					Lints: []linter.Lint{
						notFoundInSitemap.Emit(fmt.Sprintf("HTTP %d", page.StatusCode)),
					},
				})
			}
		}

		return results
	}
	linter.RegisterSiteRule(notFoundInSitemap)

	// Rule: 3XX URL in XML Sitemap
	redirectInSitemap := &linter.SiteRule{
		ID:       "sitemap-has-3xx-url",
		Name:     "URL in XML sitemap returns redirect (3XX)",
		Severity: linter.Medium,
		Category: linter.XMLSitemaps,
		Tag:      linter.Issue,
	}
	redirectInSitemap.Check = func(pages []linter.SiteLintInput) []linter.SiteLintResult {
		var results []linter.SiteLintResult

		for _, page := range pages {
			if page.InSitemap && page.StatusCode >= 300 && page.StatusCode < 400 {
				results = append(results, linter.SiteLintResult{
					URL: page.URL,
					Lints: []linter.Lint{
						redirectInSitemap.Emit(fmt.Sprintf("HTTP %d", page.StatusCode)),
					},
				})
			}
		}

		return results
	}
	linter.RegisterSiteRule(redirectInSitemap)

	// Rule: Canonicalized URL in XML Sitemap
	canonicalizedInSitemap := &linter.SiteRule{
		ID:       "sitemap-has-canonicalized-url",
		Name:     "URL in XML sitemap has non-self-referencing canonical",
		Severity: linter.High,
		Category: linter.XMLSitemaps,
		Tag:      linter.Issue,
	}
	canonicalizedInSitemap.Check = func(pages []linter.SiteLintInput) []linter.SiteLintResult {
		var results []linter.SiteLintResult

		for _, page := range pages {
			// Check if in sitemap and has a canonical that points elsewhere
			if page.InSitemap && page.Canonical != "" && page.Canonical != page.URL {
				results = append(results, linter.SiteLintResult{
					URL: page.URL,
					Lints: []linter.Lint{
						canonicalizedInSitemap.Emit(page.Canonical),
					},
				})
			}
		}

		return results
	}
	linter.RegisterSiteRule(canonicalizedInSitemap)

	// Rule: Disallowed URL in XML Sitemap
	disallowedInSitemap := &linter.SiteRule{
		ID:       "sitemap-has-disallowed-url",
		Name:     "URL in XML sitemap is disallowed by robots.txt",
		Severity: linter.High,
		Category: linter.XMLSitemaps,
		Tag:      linter.Issue,
	}
	disallowedInSitemap.Check = func(pages []linter.SiteLintInput) []linter.SiteLintResult {
		var results []linter.SiteLintResult

		for _, page := range pages {
			if page.InSitemap && page.IsDisallowed {
				results = append(results, linter.SiteLintResult{
					URL:   page.URL,
					Lints: []linter.Lint{disallowedInSitemap.Emit("")},
				})
			}
		}

		return results
	}
	linter.RegisterSiteRule(disallowedInSitemap)

	// Rule: Timeout URL in XML Sitemap
	timeoutInSitemap := &linter.SiteRule{
		ID:       "sitemap-has-timeout-url",
		Name:     "URL in XML sitemap timed out",
		Severity: linter.Medium,
		Category: linter.XMLSitemaps,
		Tag:      linter.Issue,
	}
	timeoutInSitemap.Check = func(pages []linter.SiteLintInput) []linter.SiteLintResult {
		var results []linter.SiteLintResult

		for _, page := range pages {
			if page.InSitemap && page.IsTimeout {
				results = append(results, linter.SiteLintResult{
					URL:   page.URL,
					Lints: []linter.Lint{timeoutInSitemap.Emit("")},
				})
			}
		}

		return results
	}
	linter.RegisterSiteRule(timeoutInSitemap)
}
