package rules

import (
	"fmt"
	"strings"

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

	// Rule: URL contains uppercase characters
	uppercaseURL := &linter.Rule{
		ID:       "uppercase-url",
		Name:     "URL contains upper case characters",
		Severity: linter.Medium,
		Category: linter.Internal,
		Tag:      linter.PotentialIssue,
	}
	uppercaseURL.Check = func(ctx *linter.Context) []linter.Lint {
		path := ctx.URL.Path
		if strings.ToLower(path) != path {
			return []linter.Lint{uppercaseURL.Emit(path)}
		}
		return nil
	}
	linter.Register(uppercaseURL)

	// Rule: URL contains whitespace
	whitespaceURL := &linter.Rule{
		ID:       "whitespace-url",
		Name:     "URL contains whitespace",
		Severity: linter.Medium,
		Category: linter.Internal,
		Tag:      linter.Issue,
	}
	whitespaceURL.Check = func(ctx *linter.Context) []linter.Lint {
		path := ctx.URL.Path
		rawPath := ctx.URL.RawPath
		if rawPath == "" {
			rawPath = path
		}

		// Check for actual spaces, plus signs (space encoding), or %20
		if strings.Contains(path, " ") ||
			strings.Contains(rawPath, "+") ||
			strings.Contains(strings.ToLower(rawPath), "%20") {
			return []linter.Lint{whitespaceURL.Emit(rawPath)}
		}
		return nil
	}
	linter.Register(whitespaceURL)
}
