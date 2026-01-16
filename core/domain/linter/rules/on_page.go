package rules

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/felixdorn/bare/core/domain/linter"
)

func init() {
	missingTitle := &linter.Rule{
		ID:       "missing-title",
		Name:     "Title tag is missing",
		Severity: linter.Critical,
		Category: linter.OnPage,
		Tag:      linter.Issue,
	}
	missingTitle.Check = func(ctx *linter.Context) []linter.Lint {
		if ctx.Doc.Find("head title").Length() == 0 {
			return []linter.Lint{missingTitle.Emit("")}
		}
		return nil
	}
	linter.Register(missingTitle)

	emptyTitle := &linter.Rule{
		ID:       "empty-title",
		Name:     "Title tag is empty",
		Severity: linter.Critical,
		Category: linter.OnPage,
		Tag:      linter.Issue,
	}
	emptyTitle.Check = func(ctx *linter.Context) []linter.Lint {
		title := ctx.Doc.Find("head title")
		if title.Length() > 0 && strings.TrimSpace(title.Text()) == "" {
			return []linter.Lint{emptyTitle.Emit("")}
		}
		return nil
	}
	linter.Register(emptyTitle)

	titleOutsideHead := &linter.Rule{
		ID:       "title-outside-head",
		Name:     "Title tag outside of head",
		Severity: linter.Critical,
		Category: linter.OnPage,
		Tag:      linter.Issue,
	}
	titleOutsideHead.Check = func(ctx *linter.Context) []linter.Lint {
		var lints []linter.Lint
		ctx.Doc.Find("body title").Each(func(i int, s *goquery.Selection) {
			// Check if this title is inside an SVG (where it's valid)
			if s.ParentsFiltered("svg").Length() > 0 {
				return
			}
			lints = append(lints, titleOutsideHead.Emit(""))
		})
		return lints
	}
	linter.Register(titleOutsideHead)

	emptyHTML := &linter.Rule{
		ID:       "empty-html",
		Name:     "HTML is missing or empty",
		Severity: linter.Critical,
		Category: linter.OnPage,
		Tag:      linter.Issue,
	}
	emptyHTML.Check = func(ctx *linter.Context) []linter.Lint {
		body := ctx.Doc.Find("body")
		if body.Length() == 0 || strings.TrimSpace(body.Text()) == "" {
			return []linter.Lint{emptyHTML.Emit("")}
		}
		return nil
	}
	linter.Register(emptyHTML)

	metaDescOutsideHead := &linter.Rule{
		ID:       "meta-description-outside-head",
		Name:     "Meta description outside of head",
		Severity: linter.High,
		Category: linter.OnPage,
		Tag:      linter.Issue,
	}
	metaDescOutsideHead.Check = func(ctx *linter.Context) []linter.Lint {
		if ctx.Doc.Find(`body meta[name="description"]`).Length() > 0 {
			return []linter.Lint{metaDescOutsideHead.Emit("")}
		}
		return nil
	}
	linter.Register(metaDescOutsideHead)

	multipleTitles := &linter.Rule{
		ID:       "multiple-titles",
		Name:     "Multiple title tags",
		Severity: linter.High,
		Category: linter.OnPage,
		Tag:      linter.Issue,
	}
	multipleTitles.Check = func(ctx *linter.Context) []linter.Lint {
		count := ctx.Doc.Find("head title").Length()
		if count > 1 {
			return []linter.Lint{multipleTitles.Emit(fmt.Sprintf("Found %d title tags", count))}
		}
		return nil
	}
	linter.Register(multipleTitles)

	missingH1 := &linter.Rule{
		ID:       "missing-h1",
		Name:     "H1 tag is missing",
		Severity: linter.Medium,
		Category: linter.OnPage,
		Tag:      linter.Opportunity,
	}
	missingH1.Check = func(ctx *linter.Context) []linter.Lint {
		if ctx.Doc.Find("h1").Length() == 0 {
			return []linter.Lint{missingH1.Emit("")}
		}
		return nil
	}
	linter.Register(missingH1)

	multipleH1 := &linter.Rule{
		ID:       "multiple-h1",
		Name:     "Multiple H1 tags",
		Severity: linter.Low,
		Category: linter.OnPage,
		Tag:      linter.PotentialIssue,
	}
	multipleH1.Check = func(ctx *linter.Context) []linter.Lint {
		count := ctx.Doc.Find("h1").Length()
		if count > 1 {
			return []linter.Lint{multipleH1.Emit(fmt.Sprintf("Found %d H1 tags", count))}
		}
		return nil
	}
	linter.Register(multipleH1)

	titleDescSame := &linter.Rule{
		ID:       "title-description-same",
		Name:     "Title and meta description are the same",
		Severity: linter.Low,
		Category: linter.OnPage,
		Tag:      linter.PotentialIssue,
	}
	titleDescSame.Check = func(ctx *linter.Context) []linter.Lint {
		title := strings.TrimSpace(ctx.Doc.Find("head title").Text())
		desc, exists := ctx.Doc.Find(`head meta[name="description"]`).Attr("content")
		if !exists {
			return nil
		}
		desc = strings.TrimSpace(desc)
		if title != "" && desc != "" && title == desc {
			return []linter.Lint{titleDescSame.Emit(title)}
		}
		return nil
	}
	linter.Register(titleDescSame)

	missingAlt := &linter.Rule{
		ID:       "missing-alt",
		Name:     "Image with missing alt text",
		Severity: linter.Medium,
		Category: linter.OnPage,
		Tag:      linter.Opportunity,
	}
	missingAlt.Check = func(ctx *linter.Context) []linter.Lint {
		var lints []linter.Lint
		ctx.Doc.Find("img").Each(func(i int, s *goquery.Selection) {
			// Skip images with role="presentation" (decorative)
			if role, exists := s.Attr("role"); exists && role == "presentation" {
				return
			}

			alt, exists := s.Attr("alt")
			if !exists || strings.TrimSpace(alt) == "" {
				src, _ := s.Attr("src")
				lints = append(lints, missingAlt.Emit(src))
			}
		})
		return lints
	}
	linter.Register(missingAlt)

	loremIpsum := &linter.Rule{
		ID:       "lorem-ipsum",
		Name:     "Contains Lorem Ipsum dummy text",
		Severity: linter.Medium,
		Category: linter.OnPage,
		Tag:      linter.Issue,
	}
	loremIpsum.Check = func(ctx *linter.Context) []linter.Lint {
		bodyText := strings.ToLower(ctx.Doc.Find("body").Text())
		if strings.Contains(bodyText, "lorem ipsum") {
			return []linter.Lint{loremIpsum.Emit("")}
		}
		return nil
	}
	linter.Register(loremIpsum)

	shortH1 := &linter.Rule{
		ID:       "short-h1",
		Name:     "H1 length too short",
		Severity: linter.Low,
		Category: linter.OnPage,
		Tag:      linter.Opportunity,
	}
	shortH1.Check = func(ctx *linter.Context) []linter.Lint {
		var lints []linter.Lint
		ctx.Doc.Find("h1").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			words := strings.Fields(text)
			if len(words) > 0 && len(words) < 3 {
				lints = append(lints, shortH1.Emit(fmt.Sprintf("%d words: %s", len(words), text)))
			}
		})
		return lints
	}
	linter.Register(shortH1)

	longH1 := &linter.Rule{
		ID:       "long-h1",
		Name:     "H1 length too long",
		Severity: linter.Low,
		Category: linter.OnPage,
		Tag:      linter.Opportunity,
	}
	longH1.Check = func(ctx *linter.Context) []linter.Lint {
		var lints []linter.Lint
		ctx.Doc.Find("h1").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			words := strings.Fields(text)
			if len(words) > 10 {
				lints = append(lints, longH1.Emit(fmt.Sprintf("%d words: %s", len(words), text)))
			}
		})
		return lints
	}
	linter.Register(longH1)

	shortTitle := &linter.Rule{
		ID:       "short-title",
		Name:     "Title tag length too short",
		Severity: linter.Low,
		Category: linter.OnPage,
		Tag:      linter.Opportunity,
	}
	shortTitle.Check = func(ctx *linter.Context) []linter.Lint {
		title := strings.TrimSpace(ctx.Doc.Find("head title").Text())
		if len(title) > 0 && len(title) < 40 {
			return []linter.Lint{shortTitle.Emit(fmt.Sprintf("%d chars: %s", len(title), title))}
		}
		return nil
	}
	linter.Register(shortTitle)

	longTitle := &linter.Rule{
		ID:       "long-title",
		Name:     "Title tag length too long",
		Severity: linter.Low,
		Category: linter.OnPage,
		Tag:      linter.Opportunity,
	}
	longTitle.Check = func(ctx *linter.Context) []linter.Lint {
		title := strings.TrimSpace(ctx.Doc.Find("head title").Text())
		if len(title) > 60 {
			return []linter.Lint{longTitle.Emit(fmt.Sprintf("%d chars: %s", len(title), title))}
		}
		return nil
	}
	linter.Register(longTitle)

	metaDescEmpty := &linter.Rule{
		ID:       "meta-description-empty",
		Name:     "Meta description is empty",
		Severity: linter.Low,
		Category: linter.OnPage,
		Tag:      linter.PotentialIssue,
	}
	metaDescEmpty.Check = func(ctx *linter.Context) []linter.Lint {
		meta := ctx.Doc.Find(`head meta[name="description"]`)
		if meta.Length() > 0 {
			content, exists := meta.First().Attr("content")
			if exists && strings.TrimSpace(content) == "" {
				return []linter.Lint{metaDescEmpty.Emit("")}
			}
		}
		return nil
	}
	linter.Register(metaDescEmpty)

	multipleMetaDesc := &linter.Rule{
		ID:       "multiple-meta-descriptions",
		Name:     "Multiple meta descriptions",
		Severity: linter.Low,
		Category: linter.OnPage,
		Tag:      linter.Issue,
	}
	multipleMetaDesc.Check = func(ctx *linter.Context) []linter.Lint {
		count := ctx.Doc.Find(`head meta[name="description"]`).Length()
		if count > 1 {
			return []linter.Lint{multipleMetaDesc.Emit(fmt.Sprintf("Found %d meta descriptions", count))}
		}
		return nil
	}
	linter.Register(multipleMetaDesc)

	metaDescTooLong := &linter.Rule{
		ID:       "meta-description-too-long",
		Name:     "Meta description length too long",
		Severity: linter.Low,
		Category: linter.OnPage,
		Tag:      linter.Opportunity,
	}
	metaDescTooLong.Check = func(ctx *linter.Context) []linter.Lint {
		desc, exists := ctx.Doc.Find(`head meta[name="description"]`).Attr("content")
		if exists && len(desc) > 320 {
			return []linter.Lint{metaDescTooLong.Emit(fmt.Sprintf("%d characters", len(desc)))}
		}
		return nil
	}
	linter.Register(metaDescTooLong)

	metaDescTooShort := &linter.Rule{
		ID:       "meta-description-too-short",
		Name:     "Meta description length too short",
		Severity: linter.Low,
		Category: linter.OnPage,
		Tag:      linter.Opportunity,
	}
	metaDescTooShort.Check = func(ctx *linter.Context) []linter.Lint {
		desc, exists := ctx.Doc.Find(`head meta[name="description"]`).Attr("content")
		if exists && len(strings.TrimSpace(desc)) > 0 && len(desc) < 110 {
			return []linter.Lint{metaDescTooShort.Emit(fmt.Sprintf("%d characters", len(desc)))}
		}
		return nil
	}
	linter.Register(metaDescTooShort)
}
