package config

import (
	"fmt"
	"os"
	"time"

	"github.com/felixdorn/bare/core/domain/url"
	"github.com/pelletier/go-toml/v2"
)

type Pages struct {
	Crawl   url.Paths `toml:"crawl"`
	Exclude url.Paths `toml:"exclude"`
	Include url.Paths `toml:"include"`
}

type Config struct {
	URL    *url.URL `toml:"url"`
	Output string   `toml:"output"`
	JSWait time.Duration `toml:"js_wait"`

	Pages Pages `toml:"pages"`
}

// IsURLAllowed checks if a URL is allowed based on the include and exclude rules.
func (c *Config) IsURLAllowed(u *url.URL) bool {
	// Exclude rules have precedence
	for _, p := range c.Pages.Exclude {
		if p.Matches(u.Path) {
			return false
		}
	}

	if len(c.Pages.Include) == 0 {
		return true
	}

	for _, p := range c.Pages.Include {
		if p.Matches(u.Path) {
			return true
		}
	}

	return false
}

func (c Config) Export() ([]byte, error) {
	byt, err := toml.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("could not export config: %w", err)
	}

	return byt, nil
}

func NewDefaultConfig() *Config {
	defaultURL, _ := url.Parse("http://127.0.0.1:8000")
	return &Config{
		URL:    defaultURL,
		Output: "dist/",
		JSWait: 2 * time.Second,
		Pages: Pages{
			Crawl:   url.Paths{"/"},
			Exclude: url.Paths{},
			Include: url.Paths{},
		},
	}
}

func Get() (*Config, error) {
	contents, err := os.ReadFile("bare.toml")
	if err != nil {
		return nil, fmt.Errorf("could not read bare.toml: %w", err)
	}

	var c Config
	err = toml.Unmarshal(contents, &c)
	if err != nil {
		return nil, fmt.Errorf("could not parse bare.toml: %w", err)
	}

	return &c, nil
}
