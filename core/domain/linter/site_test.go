package linter_test

import (
	"testing"

	"github.com/felixdorn/bare/core/domain/linter"
	_ "github.com/felixdorn/bare/core/domain/linter/rules"
	"github.com/stretchr/testify/assert"
)

func TestSiteLint_SingleIncomingLink_MultipleSources(t *testing.T) {
	pages := []linter.SiteLintInput{
		{
			URL: "http://example.com/",
			InternalLinks: []linter.SiteLink{
				{TargetURL: "http://example.com/about", IsFollow: true},
				{TargetURL: "http://example.com/contact", IsFollow: true},
			},
		},
		{
			URL: "http://example.com/about",
			InternalLinks: []linter.SiteLink{
				{TargetURL: "http://example.com/", IsFollow: true},
				{TargetURL: "http://example.com/contact", IsFollow: true},
			},
		},
		{
			URL: "http://example.com/contact",
			InternalLinks: []linter.SiteLink{
				{TargetURL: "http://example.com/", IsFollow: true},
				{TargetURL: "http://example.com/about", IsFollow: true}, // Also link back to /about
			},
		},
	}

	results := linter.RunSiteRules(pages)

	// /about is linked from / and /contact - should not trigger
	assert.Empty(t, findSiteLint(results["http://example.com/about"], "single-incoming-link"))

	// /contact is linked from / and /about - should not trigger
	assert.Empty(t, findSiteLint(results["http://example.com/contact"], "single-incoming-link"))

	// / is linked from /about and /contact - should not trigger
	assert.Empty(t, findSiteLint(results["http://example.com/"], "single-incoming-link"))
}

func TestSiteLint_SingleIncomingLink_OnlyOne(t *testing.T) {
	pages := []linter.SiteLintInput{
		{
			URL: "http://example.com/",
			InternalLinks: []linter.SiteLink{
				{TargetURL: "http://example.com/orphan", IsFollow: true},
			},
		},
		{
			URL: "http://example.com/orphan",
			InternalLinks: []linter.SiteLink{
				{TargetURL: "http://example.com/", IsFollow: true},
			},
		},
	}

	results := linter.RunSiteRules(pages)

	// /orphan is only linked from / - should trigger
	lint := findSiteLint(results["http://example.com/orphan"], "single-incoming-link")
	assert.NotNil(t, lint, "expected single-incoming-link lint for orphan page")
	assert.Equal(t, linter.Medium, lint.Severity)
	assert.Equal(t, linter.Opportunity, lint.Tag)
	assert.Contains(t, lint.Evidence, "http://example.com/")
}

func TestSiteLint_SingleIncomingLink_NofollowIgnored(t *testing.T) {
	pages := []linter.SiteLintInput{
		{
			URL: "http://example.com/",
			InternalLinks: []linter.SiteLink{
				{TargetURL: "http://example.com/page", IsFollow: true},
			},
		},
		{
			URL: "http://example.com/other",
			InternalLinks: []linter.SiteLink{
				{TargetURL: "http://example.com/page", IsFollow: false}, // nofollow
			},
		},
		{
			URL:           "http://example.com/page",
			InternalLinks: []linter.SiteLink{},
		},
	}

	results := linter.RunSiteRules(pages)

	// /page has one followed link (from /) and one nofollow (from /other)
	// Only followed links count, so it should trigger
	lint := findSiteLint(results["http://example.com/page"], "single-incoming-link")
	assert.NotNil(t, lint, "nofollow links should not count")
}

func TestSiteLint_SingleIncomingLink_MultipleFromSameSource(t *testing.T) {
	pages := []linter.SiteLintInput{
		{
			URL: "http://example.com/",
			InternalLinks: []linter.SiteLink{
				{TargetURL: "http://example.com/page", IsFollow: true},
				{TargetURL: "http://example.com/page", IsFollow: true}, // Same link twice
				{TargetURL: "http://example.com/page", IsFollow: true}, // Three times
			},
		},
		{
			URL:           "http://example.com/page",
			InternalLinks: []linter.SiteLink{},
		},
	}

	results := linter.RunSiteRules(pages)

	// /page has multiple links but all from the same source - should trigger
	lint := findSiteLint(results["http://example.com/page"], "single-incoming-link")
	assert.NotNil(t, lint, "multiple links from same URL should still trigger")
}

func TestSiteLint_SingleIncomingLink_NoIncomingLinks(t *testing.T) {
	pages := []linter.SiteLintInput{
		{
			URL:           "http://example.com/",
			InternalLinks: []linter.SiteLink{},
		},
		{
			URL:           "http://example.com/orphan",
			InternalLinks: []linter.SiteLink{},
		},
	}

	results := linter.RunSiteRules(pages)

	// /orphan has NO incoming links - this is a different issue (orphan page)
	// The single-incoming-link rule only fires for pages with exactly 1 source
	assert.Empty(t, findSiteLint(results["http://example.com/orphan"], "single-incoming-link"))
}

func findSiteLint(lints []linter.Lint, ruleID string) *linter.Lint {
	for _, l := range lints {
		if l.Rule == ruleID {
			return &l
		}
	}
	return nil
}
