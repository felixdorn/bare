package linter_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/felixdorn/bare/core/domain/crawler"
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
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
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
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
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
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
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
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
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
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
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
<head><title>This is a valid page title with enough characters</title></head>
<body><h1>Welcome to our example page</h1><p>Content</p><a href="/about">About</a></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
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
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
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
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
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
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
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
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
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
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
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
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
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
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
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
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
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
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
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
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
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
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
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
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
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
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
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
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
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
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "lorem-ipsum")
	assert.Nil(t, found, "should not trigger without Lorem Ipsum text")
}

func TestLinter_ShortH1(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<p>Content here.</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "short-h1")
	assert.NotNil(t, found, "expected short-h1 lint for 1-word h1")
	assert.Equal(t, linter.Low, found.Severity)
	assert.Contains(t, found.Evidence, "1 words")
}

func TestLinter_ShortH1_TwoWords(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Example h1</h1>
<p>Content here.</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "short-h1")
	assert.NotNil(t, found, "expected short-h1 lint for 2-word h1")
	assert.Contains(t, found.Evidence, "2 words")
}

func TestLinter_ShortH1_ThreeWords(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Welcome to example</h1>
<p>Content here.</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "short-h1")
	assert.Nil(t, found, "3-word h1 should not trigger lint")
}

func TestLinter_LongH1(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>This is an example h1 on an example page that is too long</h1>
<p>Content here.</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "long-h1")
	assert.NotNil(t, found, "expected long-h1 lint for 11+ word h1")
	assert.Equal(t, linter.Low, found.Severity)
	assert.Contains(t, found.Evidence, "13 words")
}

func TestLinter_LongH1_TenWords(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>One two three four five six seven eight nine ten</h1>
<p>Content here.</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "long-h1")
	assert.Nil(t, found, "10-word h1 should not trigger lint")
}

func TestLinter_ShortTitle(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Example of a short title</title></head>
<body><h1>Welcome to the page</h1><p>Content here.</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "short-title")
	assert.NotNil(t, found, "expected short-title lint")
	assert.Equal(t, linter.Low, found.Severity)
	assert.Contains(t, found.Evidence, "24 chars")
}

func TestLinter_ShortTitle_Exactly40(t *testing.T) {
	// Exactly 40 chars should NOT trigger
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>This title is exactly forty characters!!</title></head>
<body><h1>Welcome to the page</h1><p>Content here.</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "short-title")
	assert.Nil(t, found, "40-char title should not trigger lint")
}

func TestLinter_LongTitle(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Example of a rather long title that meanders around and does not concisely communicate</title></head>
<body><h1>Welcome to the page</h1><p>Content here.</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "long-title")
	assert.NotNil(t, found, "expected long-title lint")
	assert.Equal(t, linter.Low, found.Severity)
}

func TestLinter_LongTitle_Exactly60(t *testing.T) {
	// Exactly 60 chars should NOT trigger
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>This is a title that is exactly sixty characters long!</title></head>
<body><h1>Welcome to the page</h1><p>Content here.</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "long-title")
	assert.Nil(t, found, "60-char title should not trigger lint")
}

func TestLinter_MetaDescriptionTooLong(t *testing.T) {
	longDesc := strings.Repeat("a", 321)
	html := []byte(fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<title>Page Title</title>
<meta name="description" content="%s">
</head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`, longDesc))

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "meta-description-too-long")
	assert.NotNil(t, found, "expected meta-description-too-long lint")
	assert.Equal(t, linter.Low, found.Severity)
	assert.Contains(t, found.Evidence, "321")
}

func TestLinter_MetaDescriptionTooLong_Boundary(t *testing.T) {
	desc320 := strings.Repeat("a", 320)
	html := []byte(fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<title>Page Title</title>
<meta name="description" content="%s">
</head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`, desc320))

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "meta-description-too-long")
	assert.Nil(t, found, "320-char description should not trigger lint")
}

func TestLinter_MetaDescriptionTooShort(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>Page Title</title>
<meta name="description" content="Short description here">
</head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "meta-description-too-short")
	assert.NotNil(t, found, "expected meta-description-too-short lint")
	assert.Equal(t, linter.Low, found.Severity)
}

func TestLinter_MetaDescriptionTooShort_Boundary(t *testing.T) {
	desc110 := strings.Repeat("a", 110)
	html := []byte(fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<title>Page Title</title>
<meta name="description" content="%s">
</head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`, desc110))

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "meta-description-too-short")
	assert.Nil(t, found, "110-char description should not trigger lint")
}

func TestLinter_MetaDescriptionTooShort_EmptyDoesNotTrigger(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>Page Title</title>
<meta name="description" content="">
</head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "meta-description-too-short")
	assert.Nil(t, found, "empty description should not trigger too-short lint")
}

func TestLinter_MetaDescriptionTooShort_WhitespaceDoesNotTrigger(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>Page Title</title>
<meta name="description" content="   ">
</head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "meta-description-too-short")
	assert.Nil(t, found, "whitespace-only description should not trigger too-short lint")
}

func TestLinter_MetaDescriptionEmpty(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>Page Title</title>
<meta name="description" content="">
</head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "meta-description-empty")
	assert.NotNil(t, found, "expected meta-description-empty lint")
	assert.Equal(t, linter.Low, found.Severity)
	assert.Equal(t, linter.PotentialIssue, found.Tag)
}

func TestLinter_MetaDescriptionEmpty_Whitespace(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>Page Title</title>
<meta name="description" content="   ">
</head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "meta-description-empty")
	assert.NotNil(t, found, "whitespace-only description should trigger empty lint")
}

func TestLinter_MultipleMetaDescriptions(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>Page Title</title>
<meta name="description" content="First description">
<meta name="description" content="Second description">
</head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "multiple-meta-descriptions")
	assert.NotNil(t, found, "expected multiple-meta-descriptions lint")
	assert.Equal(t, linter.Low, found.Severity)
	assert.Equal(t, linter.Issue, found.Tag)
	assert.Contains(t, found.Evidence, "2")
}

func TestLinter_SingleMetaDescription_NoLint(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>Page Title</title>
<meta name="description" content="Just one description that is long enough to not trigger short">
</head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "multiple-meta-descriptions")
	assert.Nil(t, found, "single meta description should not trigger lint")
}

func TestLinter_TitleDescriptionSame(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>This is a really great page about stuff</title>
<meta name="description" content="This is a really great page about stuff">
</head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "title-description-same")
	assert.NotNil(t, found, "expected title-description-same lint")
	assert.Equal(t, linter.Low, found.Severity)
	assert.Equal(t, linter.PotentialIssue, found.Tag)
}

func TestLinter_TitleDescriptionSame_Different(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head>
<title>Page Title</title>
<meta name="description" content="This is a different description">
</head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "title-description-same")
	assert.Nil(t, found, "should not trigger when title and description differ")
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

func TestLinter_CheckOptions_StatusCode(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page Title</title></head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	opts := linter.CheckOptions{
		StatusCode: 404,
	}

	ctx, err := linter.NewContext(html, pageURL, nil, opts)
	require.NoError(t, err)
	assert.Equal(t, 404, ctx.StatusCode)
}

func TestLinter_CheckOptions_RedirectChain(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page Title</title></head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/final")
	chain := []crawler.Redirect{
		{URL: "http://example.com/old", StatusCode: 301},
		{URL: "http://example.com/middle", StatusCode: 302},
	}
	opts := linter.CheckOptions{
		StatusCode:    200,
		RedirectChain: chain,
	}

	ctx, err := linter.NewContext(html, pageURL, nil, opts)
	require.NoError(t, err)
	assert.Len(t, ctx.RedirectChain, 2)
	assert.Equal(t, "http://example.com/old", ctx.RedirectChain[0].URL)
	assert.Equal(t, 301, ctx.RedirectChain[0].StatusCode)
	assert.Equal(t, "http://example.com/middle", ctx.RedirectChain[1].URL)
	assert.Equal(t, 302, ctx.RedirectChain[1].StatusCode)
}

func TestLinter_CheckOptions_EmptyChain(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page Title</title></head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	opts := linter.CheckOptions{
		StatusCode:    200,
		RedirectChain: nil,
	}

	ctx, err := linter.NewContext(html, pageURL, nil, opts)
	require.NoError(t, err)
	assert.Empty(t, ctx.RedirectChain)
}

func TestLinter_RedirectsToSelf(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page Title</title></head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page-a")
	chain := []crawler.Redirect{
		{URL: "http://example.com/page-a", StatusCode: 301},
	}
	opts := linter.CheckOptions{
		StatusCode:    200,
		RedirectChain: chain,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "redirects-to-self")
	assert.NotNil(t, found, "expected redirects-to-self lint")
	assert.Equal(t, linter.High, found.Severity)
	assert.Equal(t, linter.Redirects, found.Category)
	assert.Equal(t, linter.Issue, found.Tag)
	assert.Contains(t, found.Evidence, "http://example.com/page-a")
}

func TestLinter_RedirectsToSelf_InChain(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page Title</title></head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page-a")
	chain := []crawler.Redirect{
		{URL: "http://example.com/page-b", StatusCode: 301},
		{URL: "http://example.com/page-a", StatusCode: 302}, // Self-redirect in middle of chain
	}
	opts := linter.CheckOptions{
		StatusCode:    200,
		RedirectChain: chain,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "redirects-to-self")
	assert.NotNil(t, found, "expected redirects-to-self lint when self-redirect is in chain")
}

func TestLinter_RedirectsToSelf_NoSelfRedirect(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page Title</title></head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page-a")
	chain := []crawler.Redirect{
		{URL: "http://example.com/page-b", StatusCode: 301},
		{URL: "http://example.com/page-c", StatusCode: 302},
	}
	opts := linter.CheckOptions{
		StatusCode:    200,
		RedirectChain: chain,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "redirects-to-self")
	assert.Nil(t, found, "should not trigger when no self-redirect exists")
}

func TestLinter_RedirectsToSelf_NoRedirects(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page Title</title></head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page-a")
	opts := linter.CheckOptions{
		StatusCode:    200,
		RedirectChain: nil,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "redirects-to-self")
	assert.Nil(t, found, "should not trigger when no redirects exist")
}

func TestLinter_BrokenInternalURL_404(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Not Found</title></head>
<body><h1>404</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page")
	opts := linter.CheckOptions{
		StatusCode: 404,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "broken-internal-url")
	assert.NotNil(t, found, "expected broken-internal-url lint for 404")
	assert.Equal(t, linter.High, found.Severity)
	assert.Equal(t, linter.Internal, found.Category)
	assert.Equal(t, linter.Issue, found.Tag)
	assert.Contains(t, found.Evidence, "404")
}

func TestLinter_BrokenInternalURL_500(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Server Error</title></head>
<body><h1>500</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page")
	opts := linter.CheckOptions{
		StatusCode: 500,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "broken-internal-url")
	assert.NotNil(t, found, "expected broken-internal-url lint for 500")
	assert.Contains(t, found.Evidence, "500")
}

func TestLinter_BrokenInternalURL_403(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Forbidden</title></head>
<body><h1>403</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page")
	opts := linter.CheckOptions{
		StatusCode: 403,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "broken-internal-url")
	assert.NotNil(t, found, "expected broken-internal-url lint for 403")
}

func TestLinter_BrokenInternalURL_200OK(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page")
	opts := linter.CheckOptions{
		StatusCode: 200,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "broken-internal-url")
	assert.Nil(t, found, "should not trigger for 200 OK")
}

func TestLinter_BrokenInternalURL_301Redirect(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page")
	opts := linter.CheckOptions{
		StatusCode: 301,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "broken-internal-url")
	assert.Nil(t, found, "should not trigger for 3XX redirects")
}

func TestLinter_RedirectBroken_4XX(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Not Found</title></head>
<body><h1>404</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page")
	chain := []crawler.Redirect{
		{URL: "http://example.com/old-page", StatusCode: 301},
	}
	opts := linter.CheckOptions{
		StatusCode:    404,
		RedirectChain: chain,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "redirect-broken")
	assert.NotNil(t, found, "expected redirect-broken lint for 4XX after redirect")
	assert.Equal(t, linter.High, found.Severity)
	assert.Equal(t, linter.Redirects, found.Category)
	assert.Contains(t, found.Evidence, "404")
}

func TestLinter_RedirectBroken_5XX(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Error</title></head>
<body><h1>500</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page")
	chain := []crawler.Redirect{
		{URL: "http://example.com/old-page", StatusCode: 302},
	}
	opts := linter.CheckOptions{
		StatusCode:    500,
		RedirectChain: chain,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "redirect-broken")
	assert.NotNil(t, found, "expected redirect-broken lint for 5XX after redirect")
	assert.Contains(t, found.Evidence, "500")
}

func TestLinter_RedirectBroken_NoRedirect(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Not Found</title></head>
<body><h1>404</h1></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page")
	opts := linter.CheckOptions{
		StatusCode:    404,
		RedirectChain: nil, // No redirects
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "redirect-broken")
	assert.Nil(t, found, "should not trigger when there are no redirects")
}

func TestLinter_RedirectBroken_SuccessfulRedirect(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body><h1>Hello</h1><p>Content</p></body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/page")
	chain := []crawler.Redirect{
		{URL: "http://example.com/old-page", StatusCode: 301},
	}
	opts := linter.CheckOptions{
		StatusCode:    200,
		RedirectChain: chain,
	}

	lints, err := linter.Check(html, pageURL, nil, opts)
	require.NoError(t, err)

	found := findLint(lints, "redirect-broken")
	assert.Nil(t, found, "should not trigger for successful redirect")
}

func TestLinter_LocalhostLink_127001WithPort(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://127.0.0.1:1234/test">Test</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "localhost-link")
	assert.NotNil(t, found, "expected localhost-link lint for 127.0.0.1 with port")
	assert.Equal(t, linter.Critical, found.Severity)
	assert.Equal(t, linter.Links, found.Category)
	assert.Contains(t, found.Evidence, "127.0.0.1:1234")
}

func TestLinter_LocalhostLink_127001WithoutPort(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://127.0.0.1/test">Test</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "localhost-link")
	assert.NotNil(t, found, "expected localhost-link lint for 127.0.0.1 without port")
}

func TestLinter_LocalhostLink_LocalhostWithPort(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://localhost:8080/test">Test</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "localhost-link")
	assert.NotNil(t, found, "expected localhost-link lint for localhost with port")
	assert.Contains(t, found.Evidence, "localhost:8080")
}

func TestLinter_LocalhostLink_LocalhostWithoutPort(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://localhost/test">Test</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "localhost-link")
	assert.NotNil(t, found, "expected localhost-link lint for localhost without port")
}

func TestLinter_LocalhostLink_CaseInsensitive(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://LOCALHOST/test">Test</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "localhost-link")
	assert.NotNil(t, found, "should detect LOCALHOST case-insensitively")
}

func TestLinter_LocalhostLink_MultipleLinks(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://localhost/one">One</a>
<a href="http://127.0.0.1/two">Two</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	count := 0
	for _, l := range lints {
		if l.Rule == "localhost-link" {
			count++
		}
	}
	assert.Equal(t, 2, count, "should detect both localhost links")
}

func TestLinter_LocalhostLink_NoLocalhostLinks(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://example.com/test">Test</a>
<a href="/relative">Relative</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "localhost-link")
	assert.Nil(t, found, "should not trigger for normal links")
}

func TestLinter_LocalFileLink_WindowsPath(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="C:\Computer\Sitebulb\page.html">Page</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "local-file-link")
	assert.NotNil(t, found, "expected local-file-link lint for Windows path")
	assert.Equal(t, linter.Critical, found.Severity)
	assert.Equal(t, linter.Links, found.Category)
	assert.Contains(t, found.Evidence, "C:\\Computer")
}

func TestLinter_LocalFileLink_WindowsPathForwardSlash(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="D:/Users/test/file.html">Page</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "local-file-link")
	assert.NotNil(t, found, "expected local-file-link lint for Windows path with forward slash")
}

func TestLinter_LocalFileLink_UNCPath(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="\\servername\path\file.html">Page</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "local-file-link")
	assert.NotNil(t, found, "expected local-file-link lint for UNC path")
	assert.Contains(t, found.Evidence, "\\\\servername")
}

func TestLinter_LocalFileLink_FileProtocol(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="file:///C:/Users/test/file.html">Page</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "local-file-link")
	assert.NotNil(t, found, "expected local-file-link lint for file:// protocol")
	assert.Contains(t, found.Evidence, "file:///")
}

func TestLinter_LocalFileLink_FileProtocolCaseInsensitive(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="FILE:///path/to/file">Page</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "local-file-link")
	assert.NotNil(t, found, "should detect FILE:// case-insensitively")
}

func TestLinter_LocalFileLink_NoLocalPaths(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://example.com/test">Test</a>
<a href="/relative/path">Relative</a>
<a href="../parent/path">Parent</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "local-file-link")
	assert.Nil(t, found, "should not trigger for normal links")
}

func TestLinter_WhitespaceHref_LeadingSpace(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href=" https://www.example.com">Test</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "whitespace-href")
	assert.NotNil(t, found, "expected whitespace-href lint for leading space")
	assert.Equal(t, linter.High, found.Severity)
	assert.Equal(t, linter.Links, found.Category)
}

func TestLinter_WhitespaceHref_TrailingSpace(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="https://www.example.com ">Test</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "whitespace-href")
	assert.NotNil(t, found, "expected whitespace-href lint for trailing space")
}

func TestLinter_WhitespaceHref_BothEnds(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="  https://www.example.com  ">Test</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "whitespace-href")
	assert.NotNil(t, found, "expected whitespace-href lint for whitespace on both ends")
}

func TestLinter_WhitespaceHref_Tabs(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="	https://www.example.com">Test</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "whitespace-href")
	assert.NotNil(t, found, "expected whitespace-href lint for tab character")
}

func TestLinter_WhitespaceHref_Newline(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="
https://www.example.com">Test</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "whitespace-href")
	assert.NotNil(t, found, "expected whitespace-href lint for newline character")
}

func TestLinter_WhitespaceHref_NoWhitespace(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="https://www.example.com">Test</a>
<a href="/relative/path">Relative</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "whitespace-href")
	assert.Nil(t, found, "should not trigger for clean hrefs")
}

func TestLinter_WhitespaceHref_MultipleLinks(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href=" /one">One</a>
<a href="/two ">Two</a>
<a href="/clean">Clean</a>
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	count := 0
	for _, l := range lints {
		if l.Rule == "whitespace-href" {
			count++
		}
	}
	assert.Equal(t, 2, count, "should detect both links with whitespace")
}

func TestLinter_NoOutgoingLinks_NoLinks(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<p>This page has no links at all.</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "no-outgoing-links")
	assert.NotNil(t, found, "expected no-outgoing-links lint")
	assert.Equal(t, linter.High, found.Severity)
	assert.Equal(t, linter.PotentialIssue, found.Tag)
}

func TestLinter_NoOutgoingLinks_OnlyFragmentLinks(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="#section1">Section 1</a>
<a href="#section2">Section 2</a>
<p id="section1">Content 1</p>
<p id="section2">Content 2</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "no-outgoing-links")
	assert.NotNil(t, found, "fragment-only links should trigger")
}

func TestLinter_NoOutgoingLinks_OnlyJavascriptLinks(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="javascript:void(0);">Click me</a>
<a href="javascript:doSomething()">Do something</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "no-outgoing-links")
	assert.NotNil(t, found, "javascript: links should trigger")
}

func TestLinter_NoOutgoingLinks_OnlyMailtoLinks(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="mailto:test@example.com">Email us</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "no-outgoing-links")
	assert.NotNil(t, found, "mailto: links should trigger")
}

func TestLinter_NoOutgoingLinks_OnlyTelLinks(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="tel:+1234567890">Call us</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "no-outgoing-links")
	assert.NotNil(t, found, "tel: links should trigger")
}

func TestLinter_NoOutgoingLinks_MixedInvalidLinks(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="#test">Anchor</a>
<a href="javascript:void(0);">JS</a>
<a href="mailto:test@test.com">Email</a>
<a href="tel:123">Phone</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "no-outgoing-links")
	assert.NotNil(t, found, "mix of invalid links should trigger")
}

func TestLinter_NoOutgoingLinks_HasValidLink(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="/about">About us</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "no-outgoing-links")
	assert.Nil(t, found, "relative link should not trigger")
}

func TestLinter_NoOutgoingLinks_HasExternalLink(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="https://external.com/page">External</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "no-outgoing-links")
	assert.Nil(t, found, "external link should not trigger")
}

func TestLinter_NoOutgoingLinks_ValidAmongInvalid(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="#test">Anchor</a>
<a href="javascript:void(0);">JS</a>
<a href="/valid-page">Valid</a>
<a href="mailto:test@test.com">Email</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "no-outgoing-links")
	assert.Nil(t, found, "one valid link among invalid should not trigger")
}

func TestLinter_MalformedHref_SingleSlash(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http:/example.com">Bad link</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "malformed-href")
	assert.NotNil(t, found, "expected malformed-href lint for single slash")
	assert.Equal(t, linter.High, found.Severity)
	assert.Contains(t, found.Evidence, "http:/example.com")
}

func TestLinter_MalformedHref_ExtraColon(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http:://example.com">Bad link</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "malformed-href")
	assert.NotNil(t, found, "expected malformed-href lint for extra colon")
}

func TestLinter_MalformedHref_MisspelledHTTP(t *testing.T) {
	testCases := []string{
		"htp://example.com",
		"htpp://example.com",
		"hhtp://example.com",
	}

	for _, href := range testCases {
		html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="` + href + `">Bad link</a>
</body>
</html>`)

		pageURL, _ := url.Parse("http://example.com/")
		lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
		require.NoError(t, err)

		found := findLint(lints, "malformed-href")
		assert.NotNil(t, found, "expected malformed-href lint for %s", href)
	}
}

func TestLinter_MalformedHref_MisspelledHTTPS(t *testing.T) {
	testCases := []string{
		"htps://example.com",
		"httpss://example.com",
		"htpps://example.com",
	}

	for _, href := range testCases {
		html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="` + href + `">Bad link</a>
</body>
</html>`)

		pageURL, _ := url.Parse("http://example.com/")
		lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
		require.NoError(t, err)

		found := findLint(lints, "malformed-href")
		assert.NotNil(t, found, "expected malformed-href lint for %s", href)
	}
}

func TestLinter_MalformedHref_ValidURLs(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://example.com">HTTP</a>
<a href="https://example.com">HTTPS</a>
<a href="/relative/path">Relative</a>
<a href="../parent">Parent</a>
<a href="page.html">Same dir</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "malformed-href")
	assert.Nil(t, found, "valid URLs should not trigger")
}

func TestLinter_MalformedHref_IgnoresSpecialSchemes(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="#anchor">Anchor</a>
<a href="javascript:void(0)">JS</a>
<a href="mailto:test@test.com">Email</a>
<a href="tel:123">Phone</a>
<a href="/valid">Valid</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "malformed-href")
	assert.Nil(t, found, "special schemes should not trigger")
}

func TestLinter_NonHTTPProtocol_FTP(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="ftp://example.com/file.zip">Download</a>
<a href="/other">Other</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "non-http-protocol")
	assert.NotNil(t, found, "expected non-http-protocol lint for ftp")
	assert.Equal(t, linter.High, found.Severity)
	assert.Equal(t, linter.PotentialIssue, found.Tag)
	assert.Contains(t, found.Evidence, "ftp://")
}

func TestLinter_NonHTTPProtocol_SSH(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="ssh://user@example.com">SSH</a>
<a href="/other">Other</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "non-http-protocol")
	assert.NotNil(t, found, "expected non-http-protocol lint for ssh")
}

func TestLinter_NonHTTPProtocol_WebSocket(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="ws://example.com/socket">WebSocket</a>
<a href="/other">Other</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "non-http-protocol")
	assert.NotNil(t, found, "expected non-http-protocol lint for ws")
}

func TestLinter_MalformedHref_InvalidProtocol(t *testing.T) {
	// gopher and news are not valid protocols - should be malformed
	testCases := []string{
		"gopher://example.com/",
		"news:rec.web",
		"irc://irc.example.com/channel",
	}

	for _, href := range testCases {
		html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="` + href + `">Link</a>
<a href="/other">Other</a>
</body>
</html>`)

		pageURL, _ := url.Parse("http://example.com/")
		lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
		require.NoError(t, err)

		found := findLint(lints, "malformed-href")
		assert.NotNil(t, found, "expected malformed-href lint for invalid protocol %s", href)
	}
}

func TestLinter_NonHTTPProtocol_HTTPValid(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="http://example.com/page">HTTP</a>
<a href="https://example.com/page">HTTPS</a>
<a href="/relative">Relative</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "non-http-protocol")
	assert.Nil(t, found, "http/https should not trigger")
}

func TestLinter_NonHTTPProtocol_IgnoresSpecialSchemes(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="mailto:test@test.com">Email</a>
<a href="tel:123">Phone</a>
<a href="javascript:void(0)">JS</a>
<a href="/other">Other</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "non-http-protocol")
	assert.Nil(t, found, "mailto/tel/javascript should not trigger")
}

func TestLinter_NonHTTPProtocol_FTPNotMalformed(t *testing.T) {
	// FTP should trigger non-http-protocol but NOT malformed-href
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<a href="ftp://example.com/file.zip">Download</a>
<a href="/other">Other</a>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	malformed := findLint(lints, "malformed-href")
	assert.Nil(t, malformed, "ftp should NOT trigger malformed-href")

	nonHTTP := findLint(lints, "non-http-protocol")
	assert.NotNil(t, nonHTTP, "ftp SHOULD trigger non-http-protocol")
}

func findLint(lints []linter.Lint, ruleID string) *linter.Lint {
	for _, l := range lints {
		if l.Rule == ruleID {
			return &l
		}
	}
	return nil
}
