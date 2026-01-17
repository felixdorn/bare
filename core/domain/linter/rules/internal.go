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

	// Rule: Query string contains tracking parameters
	trackingParams := &linter.Rule{
		ID:       "tracking-parameters",
		Name:     "Query string contains tracking parameters",
		Severity: linter.Medium,
		Category: linter.Internal,
		Tag:      linter.Issue,
	}

	// Common tracking parameters
	trackingParamList := map[string]bool{
		// UTM (Google Analytics)
		"utm_source":   true,
		"utm_medium":   true,
		"utm_campaign": true,
		"utm_term":     true,
		"utm_content":  true,
		"utm_id":       true,

		// Google Ads
		"gclid":   true,
		"gclsrc":  true,
		"dclid":   true,
		"gbraid":  true,
		"wbraid":  true,

		// Facebook/Meta
		"fbclid":          true,
		"fb_action_ids":   true,
		"fb_action_types": true,
		"fb_source":       true,

		// Microsoft/Bing
		"msclkid": true,

		// HubSpot
		"_hsenc":        true,
		"_hsmi":         true,
		"hsCtaTracking": true,
		"__hstc":        true,
		"__hsfp":        true,
		"__hssc":        true,

		// Mailchimp
		"mc_cid": true,
		"mc_eid": true,

		// Matomo/Piwik
		"pk_campaign": true,
		"pk_kwd":      true,
		"pk_source":   true,
		"pk_medium":   true,
		"pk_content":  true,
		"mtm_campaign": true,
		"mtm_source":   true,
		"mtm_medium":   true,
		"mtm_keyword":  true,
		"mtm_content":  true,

		// Social
		"igshid":  true,
		"twclid":  true,

		// Adobe Analytics
		"s_kwcid": true,

		// Other
		"ref":       true,
		"affiliate": true,
		"trk":       true,
		"clickid":   true,
	}

	trackingParams.Check = func(ctx *linter.Context) []linter.Lint {
		query := ctx.URL.Query()
		if len(query) == 0 {
			return nil
		}

		var found []string
		for param := range query {
			if trackingParamList[param] {
				found = append(found, param)
			}
		}

		if len(found) > 0 {
			return []linter.Lint{trackingParams.Emit(strings.Join(found, ", "))}
		}
		return nil
	}
	linter.Register(trackingParams)

	// Rule: URL contains non-ASCII characters
	nonASCII := &linter.Rule{
		ID:       "non-ascii-url",
		Name:     "URL contains non-ASCII characters",
		Severity: linter.Low,
		Category: linter.Internal,
		Tag:      linter.PotentialIssue,
	}
	nonASCII.Check = func(ctx *linter.Context) []linter.Lint {
		// Check path and query string for non-ASCII characters
		toCheck := ctx.URL.Path
		if ctx.URL.RawQuery != "" {
			toCheck += "?" + ctx.URL.RawQuery
		}

		for _, r := range toCheck {
			if r > 127 {
				return []linter.Lint{nonASCII.Emit(toCheck)}
			}
		}
		return nil
	}
	linter.Register(nonASCII)

	// Rule: URL contains double slash
	doubleSlash := &linter.Rule{
		ID:       "double-slash-url",
		Name:     "URL contains a double slash",
		Severity: linter.Low,
		Category: linter.Internal,
		Tag:      linter.Issue,
	}
	doubleSlash.Check = func(ctx *linter.Context) []linter.Lint {
		path := ctx.URL.Path
		if strings.Contains(path, "//") {
			return []linter.Lint{doubleSlash.Emit(path)}
		}
		return nil
	}
	linter.Register(doubleSlash)

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
