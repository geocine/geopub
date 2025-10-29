package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/geocine/geopub/internal/config"
	"github.com/geocine/geopub/internal/loader"
	r "github.com/geocine/geopub/internal/renderer"
	th "github.com/geocine/geopub/test"
	"github.com/stretchr/testify/require"
)

func TestSearchIndexFixtureGenerated(t *testing.T) {
	root := th.GeoPubPath("search", "reasonable_search_index")
	out := t.TempDir()

	cfg, err := config.LoadFromFile(filepath.Join(root, "book.toml"))
	require.NoError(t, err)
	bl := loader.NewBookLoader(root, cfg)
	book, err := bl.Load()
	require.NoError(t, err)

	rr := r.NewHtmlRenderer()
	ctx := &r.RenderContext{Root: root, DestDir: out, Book: book, Config: cfg, SourceDir: filepath.Join(root, cfg.Book.Src), AssetsFS: os.DirFS(th.RepoRoot())}
	require.NoError(t, rr.Render(ctx))

	_, err = os.ReadFile(filepath.Join(out, "searchindex.js"))
	require.NoError(t, err)
}
