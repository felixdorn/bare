package rules

import (
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
}
