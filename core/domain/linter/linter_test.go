package linter_test

import (
	"testing"

	"github.com/felixdorn/bare/core/domain/crawler"
	"github.com/felixdorn/bare/core/domain/linter"
	_ "github.com/felixdorn/bare/core/domain/linter/rules"
	"github.com/felixdorn/bare/core/domain/url"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
