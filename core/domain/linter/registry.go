package linter

import (
	"sort"
)

var rules = make(map[string]*Rule)

// Register adds a rule to the registry
func Register(r *Rule) {
	rules[r.ID] = r
}

// All returns all registered rules sorted by category then severity
func All() []*Rule {
	result := make([]*Rule, 0, len(rules))
	for _, r := range rules {
		result = append(result, r)
	}

	severityOrder := map[Severity]int{
		Critical: 0,
		High:     1,
		Medium:   2,
		Low:      3,
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Category != result[j].Category {
			return result[i].Category < result[j].Category
		}
		return severityOrder[result[i].Severity] < severityOrder[result[j].Severity]
	})

	return result
}

// Run executes all registered rules against the context and returns all lints
func Run(ctx *Context) []Lint {
	var all []Lint
	for _, rule := range rules {
		all = append(all, rule.Check(ctx)...)
	}
	return all
}
