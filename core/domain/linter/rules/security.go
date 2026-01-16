package rules

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/felixdorn/bare/core/domain/linter"
	"github.com/felixdorn/bare/core/domain/url"
)

func init() {
	mixedContent := &linter.Rule{
		ID:       "mixed-content",
		Name:     "HTTPS page loads HTTP resources (mixed content)",
		Severity: linter.Critical,
		Category: linter.Security,
		Tag:      linter.Issue,
	}
	mixedContent.Check = func(ctx *linter.Context) []linter.Lint {
		// Only applies to HTTPS pages
		if ctx.URL.Scheme != "https" {
			return nil
		}

		var lints []linter.Lint
		seen := make(map[string]bool)

		checkURL := func(rawURL string) {
			if rawURL == "" {
				return
			}

			rawURL = strings.TrimSpace(rawURL)

			// Skip data URLs and other non-network schemes
			if strings.HasPrefix(strings.ToLower(rawURL), "data:") ||
				strings.HasPrefix(strings.ToLower(rawURL), "javascript:") ||
				strings.HasPrefix(strings.ToLower(rawURL), "blob:") {
				return
			}

			// Resolve relative URLs against page URL
			parsed, err := url.Parse(rawURL)
			if err != nil {
				return
			}

			resolved := ctx.URL.ResolveReference(parsed)

			// Check if the resolved URL uses HTTP
			if resolved.Scheme == "http" {
				urlStr := resolved.String()
				if !seen[urlStr] {
					seen[urlStr] = true
					lints = append(lints, mixedContent.Emit(urlStr))
				}
			}
		}

		// Check images
		ctx.Doc.Find("img[src]").Each(func(i int, s *goquery.Selection) {
			if src, exists := s.Attr("src"); exists {
				checkURL(src)
			}
		})

		// Check image srcset
		ctx.Doc.Find("img[srcset], source[srcset]").Each(func(i int, s *goquery.Selection) {
			if srcset, exists := s.Attr("srcset"); exists {
				// srcset format: "url size, url size, ..."
				for _, entry := range strings.Split(srcset, ",") {
					parts := strings.Fields(strings.TrimSpace(entry))
					if len(parts) > 0 {
						checkURL(parts[0])
					}
				}
			}
		})

		// Check scripts
		ctx.Doc.Find("script[src]").Each(func(i int, s *goquery.Selection) {
			if src, exists := s.Attr("src"); exists {
				checkURL(src)
			}
		})

		// Check stylesheets and other link elements
		ctx.Doc.Find("link[href]").Each(func(i int, s *goquery.Selection) {
			if href, exists := s.Attr("href"); exists {
				checkURL(href)
			}
		})

		// Check video and audio sources
		ctx.Doc.Find("video[src], audio[src], source[src]").Each(func(i int, s *goquery.Selection) {
			if src, exists := s.Attr("src"); exists {
				checkURL(src)
			}
		})

		// Check video poster
		ctx.Doc.Find("video[poster]").Each(func(i int, s *goquery.Selection) {
			if poster, exists := s.Attr("poster"); exists {
				checkURL(poster)
			}
		})

		// Check iframes
		ctx.Doc.Find("iframe[src]").Each(func(i int, s *goquery.Selection) {
			if src, exists := s.Attr("src"); exists {
				checkURL(src)
			}
		})

		// Check object data
		ctx.Doc.Find("object[data]").Each(func(i int, s *goquery.Selection) {
			if data, exists := s.Attr("data"); exists {
				checkURL(data)
			}
		})

		// Check embed
		ctx.Doc.Find("embed[src]").Each(func(i int, s *goquery.Selection) {
			if src, exists := s.Attr("src"); exists {
				checkURL(src)
			}
		})

		return lints
	}
	linter.Register(mixedContent)

	internalHTTPURL := &linter.Rule{
		ID:       "internal-http-url",
		Name:     "Internal URL uses insecure HTTP protocol",
		Severity: linter.Critical,
		Category: linter.Security,
		Tag:      linter.Issue,
	}
	internalHTTPURL.Check = func(ctx *linter.Context) []linter.Lint {
		// Only trigger for HTTP URLs that return 200
		if ctx.URL.Scheme == "http" && ctx.StatusCode == 200 {
			return []linter.Lint{internalHTTPURL.Emit("")}
		}
		return nil
	}
	linter.Register(internalHTTPURL)

	httpsLinksToHTTP := &linter.Rule{
		ID:       "https-links-to-http",
		Name:     "HTTPS page links to internal HTTP URL",
		Severity: linter.High,
		Category: linter.Security,
		Tag:      linter.Issue,
	}
	httpsLinksToHTTP.Check = func(ctx *linter.Context) []linter.Lint {
		// Only applies to HTTPS pages
		if ctx.URL.Scheme != "https" {
			return nil
		}

		var lints []linter.Lint
		seen := make(map[string]bool)

		ctx.Doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if !exists || href == "" {
				return
			}

			href = strings.TrimSpace(href)
			hrefLower := strings.ToLower(href)

			// Skip non-HTTP schemes
			if strings.HasPrefix(hrefLower, "javascript:") ||
				strings.HasPrefix(hrefLower, "mailto:") ||
				strings.HasPrefix(hrefLower, "tel:") ||
				strings.HasPrefix(hrefLower, "data:") ||
				strings.HasPrefix(href, "#") {
				return
			}

			// Parse and resolve the URL
			parsed, err := url.Parse(href)
			if err != nil {
				return
			}

			resolved := ctx.URL.ResolveReference(parsed)

			// Check if it's an internal HTTP link
			if resolved.Scheme == "http" && resolved.IsInternal(ctx.URL) {
				urlStr := resolved.String()
				if !seen[urlStr] {
					seen[urlStr] = true
					lints = append(lints, httpsLinksToHTTP.Emit(urlStr))
				}
			}
		})

		return lints
	}
	linter.Register(httpsLinksToHTTP)

	httpsFormToHTTP := &linter.Rule{
		ID:       "https-form-to-http",
		Name:     "HTTPS page contains form posting to HTTP",
		Severity: linter.High,
		Category: linter.Security,
		Tag:      linter.Issue,
	}
	httpsFormToHTTP.Check = func(ctx *linter.Context) []linter.Lint {
		// Only applies to HTTPS pages
		if ctx.URL.Scheme != "https" {
			return nil
		}

		var lints []linter.Lint
		seen := make(map[string]bool)

		ctx.Doc.Find("form[action]").Each(func(i int, s *goquery.Selection) {
			action, exists := s.Attr("action")
			if !exists || action == "" {
				return
			}

			action = strings.TrimSpace(action)

			// Skip javascript: actions
			if strings.HasPrefix(strings.ToLower(action), "javascript:") {
				return
			}

			// Parse and resolve the URL
			parsed, err := url.Parse(action)
			if err != nil {
				return
			}

			resolved := ctx.URL.ResolveReference(parsed)

			// Check if the form posts to HTTP
			if resolved.Scheme == "http" {
				urlStr := resolved.String()
				if !seen[urlStr] {
					seen[urlStr] = true
					lints = append(lints, httpsFormToHTTP.Emit(urlStr))
				}
			}
		})

		return lints
	}
	linter.Register(httpsFormToHTTP)

	protocolRelativeURI := &linter.Rule{
		ID:       "protocol-relative-uri",
		Name:     "Loads resources using protocol relative URI",
		Severity: linter.High,
		Category: linter.Security,
		Tag:      linter.Issue,
	}
	protocolRelativeURI.Check = func(ctx *linter.Context) []linter.Lint {
		var lints []linter.Lint
		seen := make(map[string]bool)

		checkProtocolRelative := func(rawURL string) {
			if rawURL == "" {
				return
			}

			rawURL = strings.TrimSpace(rawURL)

			// Check if URL starts with // (protocol-relative)
			if strings.HasPrefix(rawURL, "//") {
				if !seen[rawURL] {
					seen[rawURL] = true
					lints = append(lints, protocolRelativeURI.Emit(rawURL))
				}
			}
		}

		// Check scripts
		ctx.Doc.Find("script[src]").Each(func(i int, s *goquery.Selection) {
			if src, exists := s.Attr("src"); exists {
				checkProtocolRelative(src)
			}
		})

		// Check stylesheets and other link elements
		ctx.Doc.Find("link[href]").Each(func(i int, s *goquery.Selection) {
			if href, exists := s.Attr("href"); exists {
				checkProtocolRelative(href)
			}
		})

		// Check images
		ctx.Doc.Find("img[src]").Each(func(i int, s *goquery.Selection) {
			if src, exists := s.Attr("src"); exists {
				checkProtocolRelative(src)
			}
		})

		// Check image srcset
		ctx.Doc.Find("img[srcset], source[srcset]").Each(func(i int, s *goquery.Selection) {
			if srcset, exists := s.Attr("srcset"); exists {
				for _, entry := range strings.Split(srcset, ",") {
					parts := strings.Fields(strings.TrimSpace(entry))
					if len(parts) > 0 {
						checkProtocolRelative(parts[0])
					}
				}
			}
		})

		// Check video and audio sources
		ctx.Doc.Find("video[src], audio[src], source[src]").Each(func(i int, s *goquery.Selection) {
			if src, exists := s.Attr("src"); exists {
				checkProtocolRelative(src)
			}
		})

		// Check video poster
		ctx.Doc.Find("video[poster]").Each(func(i int, s *goquery.Selection) {
			if poster, exists := s.Attr("poster"); exists {
				checkProtocolRelative(poster)
			}
		})

		// Check iframes
		ctx.Doc.Find("iframe[src]").Each(func(i int, s *goquery.Selection) {
			if src, exists := s.Attr("src"); exists {
				checkProtocolRelative(src)
			}
		})

		// Check object data
		ctx.Doc.Find("object[data]").Each(func(i int, s *goquery.Selection) {
			if data, exists := s.Attr("data"); exists {
				checkProtocolRelative(data)
			}
		})

		// Check embed
		ctx.Doc.Find("embed[src]").Each(func(i int, s *goquery.Selection) {
			if src, exists := s.Attr("src"); exists {
				checkProtocolRelative(src)
			}
		})

		return lints
	}
	linter.Register(protocolRelativeURI)
}
