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

func TestLinter_MultipleTitles_IgnoresSVGTitle(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>Page Title</title>
</head>
<body>
<h1>Hello</h1>
<svg viewBox="0 0 100 100">
  <title>SVG Icon Title</title>
  <circle cx="50" cy="50" r="40"/>
</svg>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil)
	require.NoError(t, err)

	found := findLint(lints, "multiple-titles")
	assert.Nil(t, found, "should not flag SVG title as multiple titles")
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

func TestLinter_EmptyTitle(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title></title></head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil)
	require.NoError(t, err)

	found := findLint(lints, "empty-title")
	assert.NotNil(t, found, "expected empty-title lint")
	assert.Equal(t, linter.Critical, found.Severity)
}

func TestLinter_EmptyTitle_Whitespace(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>
  </title></head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil)
	require.NoError(t, err)

	found := findLint(lints, "empty-title")
	assert.NotNil(t, found, "whitespace-only title should trigger empty-title lint")
}

func TestLinter_TitleOutsideHead(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Proper Title</title></head>
<body>
<h1>Hello</h1>
<title>Wrong place for title</title>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil)
	require.NoError(t, err)

	found := findLint(lints, "title-outside-head")
	assert.NotNil(t, found, "expected title-outside-head lint")
	assert.Equal(t, linter.Critical, found.Severity)
}

func TestLinter_TitleOutsideHead_AllowsSVG(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page Title</title></head>
<body>
<h1>Hello</h1>
<svg viewBox="0 0 100 100">
  <title>This is valid inside SVG</title>
  <circle cx="50" cy="50" r="40"/>
</svg>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil)
	require.NoError(t, err)

	found := findLint(lints, "title-outside-head")
	assert.Nil(t, found, "title inside SVG should not trigger lint")
}

func TestLinter_EmptyHTML(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil)
	require.NoError(t, err)

	found := findLint(lints, "empty-html")
	assert.NotNil(t, found, "expected empty-html lint")
	assert.Equal(t, linter.Critical, found.Severity)
}

func TestLinter_EmptyHTML_WhitespaceOnly(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>

  </body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil)
	require.NoError(t, err)

	found := findLint(lints, "empty-html")
	assert.NotNil(t, found, "whitespace-only body should trigger empty-html lint")
}

func TestLinter_MetaDescriptionOutsideHead(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<meta name="description" content="This is in the wrong place">
<h1>Hello</h1>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil)
	require.NoError(t, err)

	found := findLint(lints, "meta-description-outside-head")
	assert.NotNil(t, found, "expected meta-description-outside-head lint")
	assert.Equal(t, linter.High, found.Severity)
}

func TestLinter_MetaDescriptionInHead_Valid(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>Page</title>
<meta name="description" content="This is correct">
</head>
<body>
<h1>Hello</h1>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil)
	require.NoError(t, err)

	found := findLint(lints, "meta-description-outside-head")
	assert.Nil(t, found, "meta description in head should not trigger lint")
}

func TestLinter_MissingAlt(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<img src="test.png">
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil)
	require.NoError(t, err)

	found := findLint(lints, "missing-alt")
	assert.NotNil(t, found, "expected missing-alt lint")
	assert.Equal(t, linter.Medium, found.Severity)
	assert.Equal(t, linter.Opportunity, found.Tag)
	assert.Contains(t, found.Evidence, "test.png")
}

func TestLinter_MissingAlt_EmptyAlt(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<img src="test.png" alt="">
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil)
	require.NoError(t, err)

	found := findLint(lints, "missing-alt")
	assert.NotNil(t, found, "empty alt should trigger missing-alt lint")
}

func TestLinter_MissingAlt_WithAlt(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<img src="test.png" alt="A test image">
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil)
	require.NoError(t, err)

	found := findLint(lints, "missing-alt")
	assert.Nil(t, found, "image with alt text should not trigger lint")
}

func TestLinter_MissingAlt_RolePresentation(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<img src="decorative.png" role="presentation">
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil)
	require.NoError(t, err)

	found := findLint(lints, "missing-alt")
	assert.Nil(t, found, "decorative image with role=presentation should not trigger lint")
}

func TestLinter_LoremIpsum(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit.</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil)
	require.NoError(t, err)

	found := findLint(lints, "lorem-ipsum")
	assert.NotNil(t, found, "expected lorem-ipsum lint")
	assert.Equal(t, linter.Medium, found.Severity)
	assert.Equal(t, linter.Issue, found.Tag)
}

func TestLinter_LoremIpsum_CaseInsensitive(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<p>LOREM IPSUM dolor sit amet.</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil)
	require.NoError(t, err)

	found := findLint(lints, "lorem-ipsum")
	assert.NotNil(t, found, "should detect Lorem Ipsum case-insensitively")
}

func TestLinter_LoremIpsum_NotPresent(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<p>This is real content with no placeholder text.</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil)
	require.NoError(t, err)

	found := findLint(lints, "lorem-ipsum")
	assert.Nil(t, found, "should not trigger without Lorem Ipsum text")
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
