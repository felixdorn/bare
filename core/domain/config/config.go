package config

import (
	"fmt"
	"os"
	"time"

	"github.com/felixdorn/bare/core/domain/url"
	"github.com/pelletier/go-toml/v2"
)

type Pages struct {
	Entrypoints url.Paths `toml:"entrypoints"`
	ExtractOnly url.Paths `toml:"extract_only"`
	Exclude     url.Paths `toml:"exclude"`
}

type JS struct {
	Enabled        bool          `toml:"enabled"`
	Wait           time.Duration `toml:"wait_for"`
	ExecutablePath string        `toml:"executable_path,omitempty"`
	Flags          []string      `toml:"flags,omitempty"`
}

type Config struct {
	URL          *url.URL `toml:"url"`
	Output       string   `toml:"output"`
	WorkersCount int      `toml:"workers_count"`

	JS    JS    `toml:"js"`
	Pages Pages `toml:"pages"`
}

// IsURLAllowed checks if a URL is allowed based on the exclude rules.
func (c *Config) IsURLAllowed(u *url.URL) bool {
	return !c.Pages.Exclude.MatchAny(u.Path)
}

// IsExtractOnly checks if a URL should only have its links extracted without saving its content.
func (c *Config) IsExtractOnly(u *url.URL) bool {
	return c.Pages.ExtractOnly.MatchAny(u.Path)
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
		URL:          defaultURL,
		Output:       "dist/",
		WorkersCount: 10,
		JS: JS{
			Enabled:        false,
			Wait:           2000,
			ExecutablePath: "",
			Flags:          []string{},
		},
		Pages: Pages{
			Entrypoints: url.Paths{"/"},
			ExtractOnly: url.Paths{},
			Exclude:     url.Paths{},
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

	if c.WorkersCount == 0 {
		c.WorkersCount = 10
	}

	if len(c.Pages.Entrypoints) == 0 {
		c.Pages.Entrypoints = []url.Path{"/"}
	}

	return &c, nil
}
