package rules

import (
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
}
