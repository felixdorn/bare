package rules

import (
	"fmt"

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

	multipleTitles := &linter.Rule{
		ID:       "multiple-titles",
		Name:     "Multiple title tags",
		Severity: linter.High,
		Category: linter.OnPage,
		Tag:      linter.Issue,
	}
	multipleTitles.Check = func(ctx *linter.Context) []linter.Lint {
		count := ctx.Doc.Find("title").Length()
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
