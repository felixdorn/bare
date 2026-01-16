package linter_test

import (
	"testing"

	"github.com/felixdorn/bare/core/domain/linter"
	_ "github.com/felixdorn/bare/core/domain/linter/rules"
	"github.com/felixdorn/bare/core/domain/url"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinter_MissingTitle(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head></head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil)
	require.NoError(t, err)

	found := findLint(lints, "missing-title")
	assert.NotNil(t, found, "expected missing-title lint")
	assert.Equal(t, linter.Critical, found.Severity)
	assert.Equal(t, linter.OnPage, found.Category)
	assert.Equal(t, linter.Issue, found.Tag)
}

func TestLinter_MultipleTitles(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>First Title</title>
<title>Second Title</title>
</head>
<body><h1>Hello</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil)
	require.NoError(t, err)

	found := findLint(lints, "multiple-titles")
	assert.NotNil(t, found, "expected multiple-titles lint")
	assert.Equal(t, linter.High, found.Severity)
	assert.Contains(t, found.Evidence, "2")
}

func TestLinter_MissingH1(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page Title</title></head>
<body><h2>Not an H1</h2></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil)
	require.NoError(t, err)

	found := findLint(lints, "missing-h1")
	assert.NotNil(t, found, "expected missing-h1 lint")
	assert.Equal(t, linter.Medium, found.Severity)
	assert.Equal(t, linter.Opportunity, found.Tag)
}

func TestLinter_MultipleH1(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page Title</title></head>
<body>
<h1>First H1</h1>
<h1>Second H1</h1>
<h1>Third H1</h1>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil)
	require.NoError(t, err)

	found := findLint(lints, "multiple-h1")
	assert.NotNil(t, found, "expected multiple-h1 lint")
	assert.Equal(t, linter.Low, found.Severity)
	assert.Equal(t, linter.PotentialIssue, found.Tag)
	assert.Contains(t, found.Evidence, "3")
}

func TestLinter_ValidPage(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Valid Page Title</title></head>
<body><h1>Single H1</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil)
	require.NoError(t, err)

	// A valid page should have no lints
	assert.Empty(t, lints, "expected no lints for valid page, got: %v", lints)
}

func TestLinter_AllRulesRegistered(t *testing.T) {
	rules := linter.All()
	assert.GreaterOrEqual(t, len(rules), 4, "expected at least 4 rules registered")

	// Verify our 4 starter rules exist
	ruleIDs := make(map[string]bool)
	for _, r := range rules {
		ruleIDs[r.ID] = true
	}

	assert.True(t, ruleIDs["missing-title"], "missing-title rule not registered")
	assert.True(t, ruleIDs["multiple-titles"], "multiple-titles rule not registered")
	assert.True(t, ruleIDs["missing-h1"], "missing-h1 rule not registered")
	assert.True(t, ruleIDs["multiple-h1"], "multiple-h1 rule not registered")
}

func findLint(lints []linter.Lint, ruleID string) *linter.Lint {
	for _, l := range lints {
		if l.Rule == ruleID {
			return &l
		}
	}
	return nil
}
