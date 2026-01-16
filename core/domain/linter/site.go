package linter

// SiteLintInput represents data needed for site-level lint analysis.
type SiteLintInput struct {
	URL           string
	StatusCode    int
	InSitemap     bool
	IsNoindex     bool
	InternalLinks []SiteLink
}

// SiteLink represents an internal link for site-level analysis.
type SiteLink struct {
	TargetURL string
	IsFollow  bool
}

// SiteLintResult contains lints to add to specific pages.
type SiteLintResult struct {
	URL   string
	Lints []Lint
}

// SiteRule defines a site-level linting rule that analyzes across all pages.
type SiteRule struct {
	ID       string
	Name     string
	Severity Severity
	Category Category
	Tag      Tag
	Check    func(pages []SiteLintInput) []SiteLintResult
}

// Emit creates a lint from this site rule with optional evidence.
func (r *SiteRule) Emit(evidence string) Lint {
	return Lint{
		Rule:     r.ID,
		Message:  r.Name,
		Severity: r.Severity,
		Category: r.Category,
		Tag:      r.Tag,
		Evidence: evidence,
	}
}

var siteRules []*SiteRule

// RegisterSiteRule adds a site-level rule to the registry.
func RegisterSiteRule(rule *SiteRule) {
	siteRules = append(siteRules, rule)
}

// AllSiteRules returns all registered site-level rules.
func AllSiteRules() []*SiteRule {
	return siteRules
}

// RunSiteRules executes all site-level rules and returns lints per URL.
func RunSiteRules(pages []SiteLintInput) map[string][]Lint {
	results := make(map[string][]Lint)

	for _, rule := range siteRules {
		if rule.Check != nil {
			for _, result := range rule.Check(pages) {
				results[result.URL] = append(results[result.URL], result.Lints...)
			}
		}
	}

	return results
}
