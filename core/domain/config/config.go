package config

import (
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"os"
)

type Config struct {
	URL    string `toml:"url"`
	Output string `toml:"output"`
}

func (c Config) Export() ([]byte, error) {
	byt, err := toml.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("could not export config: %w", err)
	}

	return byt, nil
}

func NewDefaultConfig() *Config {
	return &Config{
		URL:    "http://localhost:8000",
		Output: "dist/",
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
