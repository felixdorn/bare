package linter_test

import (
	"fmt"
	"strings"
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

func TestLinter_ShortAltText(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<img src="test.png" alt="Logo">
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "short-alt-text")
	assert.NotNil(t, found, "expected short-alt-text lint for 4-char alt")
	assert.Equal(t, linter.High, found.Severity)
	assert.Equal(t, linter.Issue, found.Tag)
	assert.Contains(t, found.Evidence, "test.png")
	assert.Contains(t, found.Evidence, "Logo")
}

func TestLinter_ShortAltText_NineChars(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<img src="test.png" alt="123456789">
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "short-alt-text")
	assert.NotNil(t, found, "expected short-alt-text lint for 9-char alt")
}

func TestLinter_ShortAltText_TenChars(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<img src="test.png" alt="1234567890">
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "short-alt-text")
	assert.Nil(t, found, "10-char alt should not trigger lint")
}

func TestLinter_ShortAltText_GoodAlt(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<img src="test.png" alt="A detailed description of the image">
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "short-alt-text")
	assert.Nil(t, found, "good alt text should not trigger lint")
}

func TestLinter_ShortAltText_RolePresentation(t *testing.T) {
	html := []byte(`<!DOCTYPE html>
<html>
<head><title>Page</title></head>
<body>
<h1>Hello</h1>
<img src="test.png" alt="X" role="presentation">
<p>Content</p>
</body>
</html>`)

	pageURL, _ := url.Parse("http://example.com/")
	lints, err := linter.Check(html, pageURL, nil, linter.CheckOptions{})
	require.NoError(t, err)

	found := findLint(lints, "short-alt-text")
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
