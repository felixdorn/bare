# Adding Linting Rules

This guide explains how to add new SEO linting rules to dalin.

## Quick Start

Create a rule in the appropriate category file (or create a new one):

```go
package rules

import (
    "github.com/felixdorn/bare/core/domain/linter"
)

func init() {
    rule := &linter.Rule{
        ID:       "missing-meta-description",
        Name:     "Meta description is missing",
        Severity: linter.High,
        Category: linter.OnPage,
        Tag:      linter.Issue,
    }
    rule.Check = func(ctx *linter.Context) []linter.Lint {
        if ctx.Doc.Find(`meta[name="description"]`).Length() == 0 {
            return []linter.Lint{rule.Emit("")}
        }
        return nil
    }
    linter.Register(rule)
}
```

## Rule Structure

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `string` | Unique identifier, kebab-case (e.g., `missing-title`) |
| `Name` | `string` | Human-readable message shown in reports |
| `Severity` | `Severity` | `Critical`, `High`, `Medium`, or `Low` |
| `Category` | `Category` | See categories below |
| `Tag` | `Tag` | `Issue`, `Opportunity`, or `PotentialIssue` |
| `Check` | `func(*Context) []Lint` | The check function |

### Severity Levels

| Severity | When to Use |
|----------|-------------|
| `Critical` | Blocks indexing or causes major SEO problems |
| `High` | Significant issues that should be fixed |
| `Medium` | Notable issues, improvements recommended |
| `Low` | Minor issues or best practice suggestions |

### Categories

```go
Indexability     // Robots, noindex, crawlability
Links            // Broken links, redirects
OnPage           // Title, description, headings, content
Redirects        // Redirect chains, loops
Internal         // Internal linking structure
SearchTraffic    // Analytics, search console data
XMLSitemaps      // Sitemap issues
Security         // HTTPS, mixed content
International    // Hreflang, language tags
Accessibility    // Alt text, ARIA, contrast
AMP              // AMP validation
DuplicateContent // Canonical, duplicate pages
MobileFriendly   // Viewport, tap targets
Performance      // Page speed, resource size
Rendered         // JS rendering issues
```

### Tags

| Tag | When to Use |
|-----|-------------|
| `Issue` | Something is wrong and needs fixing |
| `Opportunity` | Something is missing that could help |
| `PotentialIssue` | Might be a problem depending on context |

## The Context Object

Your `Check` function receives a `*Context` with:

```go
type Context struct {
    Doc      *goquery.Document  // Parsed HTML (use CSS selectors)
    URL      *url.URL           // Page URL
    Body     []byte             // Raw HTML bytes
    Analysis *analyzer.Analysis // Pre-extracted metadata (may be nil)
}
```

### Using goquery Selectors

```go
// Count elements
count := ctx.Doc.Find("h1").Length()

// Check if element exists
exists := ctx.Doc.Find(`meta[name="robots"]`).Length() > 0

// Get attribute value
content, exists := ctx.Doc.Find(`meta[name="description"]`).Attr("content")

// Iterate over elements
ctx.Doc.Find("img").Each(func(i int, s *goquery.Selection) {
    src, _ := s.Attr("src")
    alt, _ := s.Attr("alt")
    // ...
})
```

## Emitting Lints

### Single Lint (Pass/Fail)

Most rules emit 0 or 1 lint:

```go
rule.Check = func(ctx *linter.Context) []linter.Lint {
    if somethingIsWrong {
        return []linter.Lint{rule.Emit("optional evidence")}
    }
    return nil
}
```

### Multiple Lints

Some rules check multiple items and emit a lint for each failure:

```go
rule.Check = func(ctx *linter.Context) []linter.Lint {
    var lints []linter.Lint
    ctx.Doc.Find("img").Each(func(i int, s *goquery.Selection) {
        alt, exists := s.Attr("alt")
        if !exists || alt == "" {
            src, _ := s.Attr("src")
            lints = append(lints, rule.Emit(src))
        }
    })
    return lints
}
```

### Evidence

The `Emit(evidence)` method accepts a string shown in reports to help identify the specific issue:

```go
rule.Emit("")                                    // No evidence
rule.Emit("Found 3 H1 tags")                     // Count
rule.Emit("/images/logo.png")                    // Specific element
rule.Emit(fmt.Sprintf("Title is %d chars", n))  // Formatted details
```

## File Organization

Rules are organized by category:

```
core/domain/linter/rules/
├── README.md        # This file
├── on_page.go       # Title, H1, meta tags
├── accessibility.go # Alt text, ARIA
├── security.go      # HTTPS, mixed content
├── performance.go   # Image size, resources
└── links.go         # Broken links, redirects
```

Create a new file for a new category, or add to an existing one.

## Complete Example

```go
// rules/accessibility.go
package rules

import (
    "fmt"
    "github.com/felixdorn/bare/core/domain/linter"
)

func init() {
    // Rule: Images must have alt text
    missingAlt := &linter.Rule{
        ID:       "missing-alt",
        Name:     "Image missing alt text",
        Severity: linter.High,
        Category: linter.Accessibility,
        Tag:      linter.Issue,
    }
    missingAlt.Check = func(ctx *linter.Context) []linter.Lint {
        var lints []linter.Lint
        ctx.Doc.Find("img").Each(func(i int, s *goquery.Selection) {
            alt, exists := s.Attr("alt")
            if !exists || alt == "" {
                src, _ := s.Attr("src")
                lints = append(lints, missingAlt.Emit(src))
            }
        })
        return lints
    }
    linter.Register(missingAlt)

    // Rule: Alt text should not be too long
    longAlt := &linter.Rule{
        ID:       "long-alt-text",
        Name:     "Alt text is too long",
        Severity: linter.Low,
        Category: linter.Accessibility,
        Tag:      linter.PotentialIssue,
    }
    longAlt.Check = func(ctx *linter.Context) []linter.Lint {
        var lints []linter.Lint
        ctx.Doc.Find("img[alt]").Each(func(i int, s *goquery.Selection) {
            alt, _ := s.Attr("alt")
            if len(alt) > 125 {
                lints = append(lints, longAlt.Emit(
                    fmt.Sprintf("%d chars: %.50s...", len(alt), alt),
                ))
            }
        })
        return lints
    }
    linter.Register(longAlt)
}
```

## Testing Your Rule

Add tests in `linter_test.go`:

```go
func TestLinter_MissingAlt(t *testing.T) {
    html := []byte(`<html><head><title>Test</title></head>
        <body><h1>Hi</h1><img src="test.png"></body></html>`)

    pageURL, _ := url.Parse("http://example.com/")
    lints, err := linter.Check(html, pageURL, nil)
    require.NoError(t, err)

    found := findLint(lints, "missing-alt")
    assert.NotNil(t, found)
    assert.Equal(t, linter.High, found.Severity)
    assert.Contains(t, found.Evidence, "test.png")
}
```

Run tests:

```bash
go test -v ./core/domain/linter/
```
