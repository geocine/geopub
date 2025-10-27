package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// BookConfig contains metadata about the book
type BookConfig struct {
	Title       string   `toml:"title"`
	Authors     []string `toml:"authors"`
	Description string   `toml:"description"`
	Language    string   `toml:"language"`
	Src         string   `toml:"src"` // Source directory, defaults to "src"
}

// DefaultBookConfig returns a book config with defaults
func DefaultBookConfig() BookConfig {
	return BookConfig{
		Title:       "My Book",
		Authors:     []string{},
		Description: "",
		Language:    "en",
		Src:         "src",
	}
}

// BuildConfig contains build settings
type BuildConfig struct {
	BuildDir       string   `toml:"build-dir"`
	CreateMissing  bool     `toml:"create-missing"`
	ExtraWatchDirs []string `toml:"extra-watch-dirs"`
}

// DefaultBuildConfig returns a build config with defaults
func DefaultBuildConfig() BuildConfig {
	return BuildConfig{
		BuildDir:       "book",
		CreateMissing:  false,
		ExtraWatchDirs: []string{},
	}
}

// HtmlConfig contains HTML renderer settings
type HtmlConfig struct {
	Theme         *string `toml:"theme"`
	CodeHighlight string  `toml:"highlight"`
	SearchEnabled bool    `toml:"search"`
	PrintEnabled  bool    `toml:"print"`
}

// DefaultHtmlConfig returns HTML config with defaults
func DefaultHtmlConfig() HtmlConfig {
	return HtmlConfig{
		CodeHighlight: "highlight.js",
		SearchEnabled: true,
		PrintEnabled:  true,
	}
}

// Config is the top-level configuration
type Config struct {
	Book         BookConfig             `toml:"book"`
	Build        BuildConfig            `toml:"build"`
	Output       map[string]interface{} `toml:"output"`
	Preprocessor map[string]interface{} `toml:"preprocessor"`
	raw          map[string]interface{} // Raw TOML values
}

// NewDefaultConfig returns a config with sensible defaults
func NewDefaultConfig() *Config {
	return &Config{
		Book:         DefaultBookConfig(),
		Build:        DefaultBuildConfig(),
		Output:       make(map[string]interface{}),
		Preprocessor: make(map[string]interface{}),
		raw:          make(map[string]interface{}),
	}
}

// LoadFromFile loads configuration from a book.toml file
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := NewDefaultConfig()
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := toml.Unmarshal(data, &cfg.raw); err != nil {
		return nil, fmt.Errorf("failed to parse raw config: %w", err)
	}

	cfg.UpdateFromEnv()
	return cfg, nil
}

// LoadFromString loads configuration from a TOML string
func LoadFromString(content string) (*Config, error) {
	cfg := NewDefaultConfig()
	if err := toml.Unmarshal([]byte(content), cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := toml.Unmarshal([]byte(content), &cfg.raw); err != nil {
		return nil, fmt.Errorf("failed to parse raw config: %w", err)
	}

	cfg.UpdateFromEnv()
	return cfg, nil
}

// UpdateFromEnv updates config from environment variables
// Variables starting with GEOPUB_ are used
// GEOPUB_FOO_BAR -> foo-bar
// GEOPUB_FOO__BAR -> foo.bar
func (c *Config) UpdateFromEnv() {
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, "GEOPUB_") {
			continue
		}

		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimPrefix(parts[0], "GEOPUB_")
		value := parts[1]

		// Convert GEOPUB_KEY format to config key
		configKey := strings.ToLower(key)
		configKey = strings.ReplaceAll(configKey, "__", ".")
		configKey = strings.ReplaceAll(configKey, "_", "-")

		c.Set(configKey, value)
	}
}

// Set sets a configuration value using dot notation (e.g., "book.title", "output.renderer.theme")
func (c *Config) Set(key, value string) {
	parts := strings.Split(key, ".")

	switch parts[0] {
	case "book":
		if len(parts) >= 2 {
			c.setBookValue(parts[1:], value)
		}
	case "build":
		if len(parts) >= 2 {
			c.setBuildValue(parts[1:], value)
		}
	default:
		// Store in raw map
		c.setRawValue(parts, value)
	}
}

func (c *Config) setBookValue(parts []string, value string) {
	if len(parts) == 0 {
		return
	}

	key := parts[0]
	switch strings.ToLower(key) {
	case "title":
		c.Book.Title = value
	case "authors":
		// Parse as JSON array if possible, otherwise as single string
		c.Book.Authors = []string{value}
	case "description":
		c.Book.Description = value
	case "language":
		c.Book.Language = value
	case "src":
		c.Book.Src = value
	}
}

func (c *Config) setBuildValue(parts []string, value string) {
	if len(parts) == 0 {
		return
	}

	key := parts[0]
	switch strings.ToLower(key) {
	case "build-dir":
		c.Build.BuildDir = value
	case "create-missing":
		c.Build.CreateMissing = strings.ToLower(value) == "true"
	}
}

func (c *Config) setRawValue(parts []string, value string) {
	current := c.raw
	for i, part := range parts[:len(parts)-1] {
		if current[part] == nil {
			current[part] = make(map[string]interface{})
		}
		if m, ok := current[part].(map[string]interface{}); ok {
			current = m
		} else if i == len(parts)-2 {
			// Convert to map if needed
			current[part] = map[string]interface{}{}
			if m, ok := current[part].(map[string]interface{}); ok {
				current = m
			}
		}
	}

	if len(parts) > 0 {
		current[parts[len(parts)-1]] = value
	}
}

// Get retrieves a value from the config using dot notation
func (c *Config) Get(key string) (interface{}, bool) {
	parts := strings.Split(key, ".")

	if parts[0] == "output" && len(parts) > 1 {
		val, ok := c.Output[parts[1]]
		return val, ok
	} else if parts[0] == "preprocessor" && len(parts) > 1 {
		val, ok := c.Preprocessor[parts[1]]
		return val, ok
	}

	// Check raw values
	current := c.raw
	for _, part := range parts {
		if v, ok := current[part]; ok {
			if m, isMap := v.(map[string]interface{}); isMap {
				current = m
			} else {
				return v, true
			}
		} else {
			return nil, false
		}
	}

	return current, true
}

// GetString retrieves a string value from config
func (c *Config) GetString(key string, defaultVal string) string {
	val, ok := c.Get(key)
	if !ok {
		return defaultVal
	}
	if s, isStr := val.(string); isStr {
		return s
	}
	return defaultVal
}

// GetBool retrieves a bool value from config
func (c *Config) GetBool(key string, defaultVal bool) bool {
	val, ok := c.Get(key)
	if !ok {
		return defaultVal
	}
	if b, isBool := val.(bool); isBool {
		return b
	}
	return defaultVal
}

// GetHtmlConfig returns the HTML renderer configuration
func (c *Config) GetHtmlConfig() *HtmlConfig {
	htmlCfg := DefaultHtmlConfig()
	if output, ok := c.Output["html"]; ok {
		if m, isMap := output.(map[string]interface{}); isMap {
			if theme, ok := m["theme"]; ok {
				if s, isStr := theme.(string); isStr {
					htmlCfg.Theme = &s
				}
			}
			if highlight, ok := m["highlight"]; ok {
				if s, isStr := highlight.(string); isStr {
					htmlCfg.CodeHighlight = s
				}
			}
		}
	}
	return &htmlCfg
}
