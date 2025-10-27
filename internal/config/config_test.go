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
    t.Cleanup(func(){
        _ = os.Unsetenv("GEOPUB_BOOK__TITLE")
        _ = os.Unsetenv("GEOPUB_BUILD__BUILD-DIR")
    })

    cfg := NewDefaultConfig()
    cfg.UpdateFromEnv()

    assert.Equal(t, "Env Title", cfg.Book.Title)
    assert.Equal(t, "env-book", cfg.Build.BuildDir)
}
