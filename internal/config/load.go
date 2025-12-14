package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type Social struct {
	Title string `yaml:"title"`
	Link  string `yaml:"link"`
}

type Config struct {
	Site struct {
		Name    string `yaml:"name"`
		Suffix  string `yaml:"suffix"`
		BaseURL string `yaml:"base_url"`
	} `yaml:"site"`

	Build struct {
		Output string `yaml:"output"`
		Mode   string `yaml:"mode"`
	} `yaml:"build"`

	Theme string `yaml:"theme"`

	IgnorePatterns []string `yaml:"ignorePatterns"`

	Socials []Social `yaml:"socials"`
}

const ConfigFile = "geode.config.yaml"

const (
	ModeDraft    = "draft"
	ModeExplicit = "explicit"
)

func Load() (*Config, error) {
	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	if cfg.Theme == "" {
		cfg.Theme = "default"
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	if cfg.Site.Name == "" {
		return errors.New("site.name is required")
	}

	if cfg.Site.BaseURL == "" {
		return errors.New("site.base_url is required")
	}

	if cfg.Build.Output == "" {
		return errors.New("build.output is required")
	}

	switch cfg.Build.Mode {
	case ModeDraft, ModeExplicit:
	// valid
	default:
		return errors.New(`build.mode must be either "draft" or "explicit"`)
	}

	return nil
}
