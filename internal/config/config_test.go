package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromStringAndGetters(t *testing.T) {
	toml := `
[book]
title = "My Book"
language = "en"
src = "src"

[build]
build-dir = "out"
create-missing = true

[output.html]
theme = "light"
highlight = "highlight.js"
`

	cfg, err := LoadFromString(toml)
	require.NoError(t, err)

	assert.Equal(t, "My Book", cfg.Book.Title)
	assert.Equal(t, "out", cfg.Build.BuildDir)
	assert.True(t, cfg.Build.CreateMissing)

	// Html config projection
	htmlCfg := cfg.GetHtmlConfig()
	require.NotNil(t, htmlCfg.Theme)
	assert.Equal(t, "light", *htmlCfg.Theme)
	assert.Equal(t, "highlight.js", htmlCfg.CodeHighlight)
}

func TestUpdateFromEnv(t *testing.T) {
	// set and ensure cleanup
	_ = os.Setenv("GEOPUB_BOOK__TITLE", "Env Title")
	_ = os.Setenv("GEOPUB_BUILD__BUILD-DIR", "env-book")
	t.Cleanup(func() {
		_ = os.Unsetenv("GEOPUB_BOOK__TITLE")
		_ = os.Unsetenv("GEOPUB_BUILD__BUILD-DIR")
	})

	cfg := NewDefaultConfig()
	cfg.UpdateFromEnv()

	assert.Equal(t, "Env Title", cfg.Book.Title)
	assert.Equal(t, "env-book", cfg.Build.BuildDir)
}

func TestNestedConfigKeys(t *testing.T) {
	toml := `
[book]
title = "Test Book"

[output.html]
default-theme = "rust"
preferred-dark-theme = "coal"
git-repository-url = "https://github.com/user/myproject"
edit-url-template = "https://github.com/user/myproject/edit/main/src/{path}"
git-repository-icon = "fa-gitlab"
`

	cfg, err := LoadFromString(toml)
	require.NoError(t, err)

	// Test nested key access for theme settings
	assert.Equal(t, "rust", cfg.GetString("output.html.default-theme", ""))
	assert.Equal(t, "coal", cfg.GetString("output.html.preferred-dark-theme", ""))

	// Test git repository settings
	assert.Equal(t, "https://github.com/user/myproject", cfg.GetString("output.html.git-repository-url", ""))
	assert.Equal(t, "https://github.com/user/myproject/edit/main/src/{path}", cfg.GetString("output.html.edit-url-template", ""))
	assert.Equal(t, "fa-gitlab", cfg.GetString("output.html.git-repository-icon", ""))

	// Test default values for missing keys
	assert.Equal(t, "default", cfg.GetString("output.html.nonexistent", "default"))
}

func TestGetWithDefaultValues(t *testing.T) {
	toml := `
[book]
title = "Test"

[output.html]
`

	cfg, err := LoadFromString(toml)
	require.NoError(t, err)

	// Test that defaults work when keys are missing
	assert.Equal(t, "light", cfg.GetString("output.html.default-theme", "light"))
	assert.Equal(t, "navy", cfg.GetString("output.html.preferred-dark-theme", "navy"))
	assert.Equal(t, "", cfg.GetString("output.html.git-repository-url", ""))
}
