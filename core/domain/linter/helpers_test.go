package linter_test

import "github.com/felixdorn/bare/core/domain/linter"

func findLint(lints []linter.Lint, ruleID string) *linter.Lint {
	for _, l := range lints {
		if l.Rule == ruleID {
			return &l
		}
	}
	return nil
}
