package rules

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/felixdorn/bare/core/domain/linter"
)


func init() {
	localhostLink := &linter.Rule{
		ID:       "localhost-link",
		Name:     "Has link with a URL referencing LocalHost or 127.0.0.1",
		Severity: linter.Critical,
		Category: linter.Links,
		Tag:      linter.Issue,
	}
	localhostLink.Check = func(ctx *linter.Context) []linter.Lint {
		var lints []linter.Lint

		ctx.Doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if !exists {
				return
			}

			hrefLower := strings.ToLower(href)

			// Check for localhost or 127.0.0.1
			if strings.Contains(hrefLower, "://localhost") ||
				strings.Contains(hrefLower, "://127.0.0.1") {
				lints = append(lints, localhostLink.Emit(href))
			}
		})

		return lints
	}
	linter.Register(localhostLink)

	// Matches Windows drive letters like C:\ D:/ etc.
	driveLetterPattern := regexp.MustCompile(`(?i)^[a-z]:[/\\]`)

	localFileLink := &linter.Rule{
		ID:       "local-file-link",
		Name:     "Has link with a URL referencing a local or UNC file path",
		Severity: linter.Critical,
		Category: linter.Links,
		Tag:      linter.Issue,
	}
	localFileLink.Check = func(ctx *linter.Context) []linter.Lint {
		var lints []linter.Lint

		ctx.Doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if !exists {
				return
			}

			// Check for UNC paths (\\server\path)
			if strings.HasPrefix(href, "\\\\") {
				lints = append(lints, localFileLink.Emit(href))
				return
			}

			// Check for file:// protocol
			if strings.HasPrefix(strings.ToLower(href), "file://") {
				lints = append(lints, localFileLink.Emit(href))
				return
			}

			// Check for Windows drive letter paths (C:\, D:/, etc.)
			if driveLetterPattern.MatchString(href) {
				lints = append(lints, localFileLink.Emit(href))
				return
			}
		})

		return lints
	}
	linter.Register(localFileLink)

	whitespaceHref := &linter.Rule{
		ID:       "whitespace-href",
		Name:     "Has a link with whitespace in href attribute",
		Severity: linter.High,
		Category: linter.Links,
		Tag:      linter.Issue,
	}
	whitespaceHref.Check = func(ctx *linter.Context) []linter.Lint {
		var lints []linter.Lint

		ctx.Doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if !exists || href == "" {
				return
			}

			// Check for leading or trailing whitespace
			trimmed := strings.TrimSpace(href)
			if trimmed != href {
				lints = append(lints, whitespaceHref.Emit(href))
			}
		})

		return lints
	}
	linter.Register(whitespaceHref)

	noOutgoingLinks := &linter.Rule{
		ID:       "no-outgoing-links",
		Name:     "Has no outgoing links",
		Severity: linter.High,
		Category: linter.Links,
		Tag:      linter.PotentialIssue,
	}
	noOutgoingLinks.Check = func(ctx *linter.Context) []linter.Lint {
		hasValidLink := false

		ctx.Doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
			if hasValidLink {
				return // Already found one
			}

			href, exists := s.Attr("href")
			if !exists || href == "" {
				return
			}

			href = strings.TrimSpace(href)
			hrefLower := strings.ToLower(href)

			// Skip fragment-only links (#something)
			if strings.HasPrefix(href, "#") {
				return
			}

			// Skip javascript: links
			if strings.HasPrefix(hrefLower, "javascript:") {
				return
			}

			// Skip mailto: links
			if strings.HasPrefix(hrefLower, "mailto:") {
				return
			}

			// Skip tel: links
			if strings.HasPrefix(hrefLower, "tel:") {
				return
			}

			// Skip data: links
			if strings.HasPrefix(hrefLower, "data:") {
				return
			}

			// This is a valid outgoing link
			hasValidLink = true
		})

		if !hasValidLink {
			return []linter.Lint{noOutgoingLinks.Emit("")}
		}
		return nil
	}
	linter.Register(noOutgoingLinks)

	malformedHref := &linter.Rule{
		ID:       "malformed-href",
		Name:     "Has outgoing links with malformed href data",
		Severity: linter.High,
		Category: linter.Links,
		Tag:      linter.Issue,
	}
	malformedHref.Check = func(ctx *linter.Context) []linter.Lint {
		var lints []linter.Lint

		ctx.Doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if !exists || href == "" {
				return
			}

			href = strings.TrimSpace(href)
			hrefLower := strings.ToLower(href)

			// Skip fragment-only links
			if strings.HasPrefix(href, "#") {
				return
			}

			// Skip special schemes
			if strings.HasPrefix(hrefLower, "javascript:") ||
				strings.HasPrefix(hrefLower, "mailto:") ||
				strings.HasPrefix(hrefLower, "tel:") ||
				strings.HasPrefix(hrefLower, "data:") {
				return
			}

			// Try to parse the URL
			parsed, err := url.Parse(href)
			if err != nil {
				lints = append(lints, malformedHref.Emit(href))
				return
			}

			// Check for malformed absolute URLs
			if parsed.Scheme != "" {
				scheme := parsed.Scheme

				// Valid protocols list
				validProtocols := map[string]bool{
					"blob": true, "data": true, "file": true, "ftp": true,
					"http": true, "https": true, "javascript": true, "mailto": true,
					"resource": true, "ssh": true, "tel": true, "urn": true,
					"view-source": true, "ws": true, "wss": true,
				}

				if scheme == "http" || scheme == "https" {
					// http/https must have a host
					if parsed.Host == "" {
						lints = append(lints, malformedHref.Emit(href))
					}
				} else if !validProtocols[scheme] {
					// Unknown scheme - invalid/malformed
					lints = append(lints, malformedHref.Emit(href))
				}
			}
		})

		return lints
	}
	linter.Register(malformedHref)

	nonHTTPProtocol := &linter.Rule{
		ID:       "non-http-protocol",
		Name:     "Has link to a non-HTTP protocol",
		Severity: linter.High,
		Category: linter.Links,
		Tag:      linter.PotentialIssue,
	}
	nonHTTPProtocol.Check = func(ctx *linter.Context) []linter.Lint {
		var lints []linter.Lint

		// Non-HTTP protocols that we flag (valid but unusual for web links)
		nonHTTPSchemes := map[string]bool{
			"ftp": true, "file": true, "ssh": true,
			"ws": true, "wss": true, "blob": true,
			"urn": true, "resource": true, "view-source": true,
		}

		ctx.Doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if !exists || href == "" {
				return
			}

			href = strings.TrimSpace(href)

			parsed, err := url.Parse(href)
			if err != nil {
				return // Handled by malformed-href
			}

			if nonHTTPSchemes[parsed.Scheme] {
				lints = append(lints, nonHTTPProtocol.Emit(href))
			}
		})

		return lints
	}
	linter.Register(nonHTTPProtocol)
}
