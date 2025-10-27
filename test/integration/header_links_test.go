package integration

import (
    "os"
    "path/filepath"
    "testing"

    th "github.com/geocine/geopub/test"
    "github.com/geocine/geopub/internal/config"
    "github.com/geocine/geopub/internal/loader"
    r "github.com/geocine/geopub/internal/renderer"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestHeaderLinksFixtureFullRender(t *testing.T) {
    root := th.GeoPubPath("rendering", "header_links")
    out := t.TempDir()

    cfg, err := config.LoadFromFile(filepath.Join(root, "book.toml"))
    require.NoError(t, err)
    bl := loader.NewBookLoader(root, cfg)
    book, err := bl.Load()
    require.NoError(t, err)

    rr := r.NewHtmlRenderer()
    ctx := &r.RenderContext{ Root: root, DestDir: out, Book: book, Config: cfg, SourceDir: filepath.Join(root, cfg.Book.Src), AssetsFS: os.DirFS(th.RepoRoot()) }
    require.NoError(t, rr.Render(ctx))

    b, err := os.ReadFile(filepath.Join(out, "header_links.html"))
    require.NoError(t, err)
    html := string(b)

    assert.Contains(t, html, `<h1 id="header-links"><a class="header" href="#header-links">Header Links</a>`)
    assert.Contains(t, html, `id="hï"`)
    assert.Contains(t, html, `id="repeat"`)
    assert.Contains(t, html, `id="repeat-1"`)
    assert.Contains(t, html, `id="repeat-2"`)
    assert.Contains(t, html, `id="repeat-1-1"`)
    assert.Contains(t, html, `id="with-emphasis-bold-bold_emphasis-code-escaped-html-link-httpsexamplecom"`)
}
