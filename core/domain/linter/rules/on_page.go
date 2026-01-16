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
}
