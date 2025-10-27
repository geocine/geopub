package loader

import (
    "path/filepath"
    "testing"

    "github.com/geocine/geopub/internal/config"
    "github.com/geocine/geopub/internal/models"
    "github.com/geocine/geopub/internal/testutil"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestLoadCreatesMissingAndLoadsChapters(t *testing.T) {
    root := testutil.TempBook(t, "book")

    // Write SUMMARY.md with one existing, one draft, one missing file
    testutil.WriteFile(t, root, filepath.Join("src", "SUMMARY.md"), `# Summary

- [Intro](intro.md)
- [Draft]()
- [Chapter 1](ch1/one.md)
`)

    // Create only intro.md
    testutil.WriteFile(t, root, filepath.Join("src", "intro.md"), "# Intro\nBody")

    // Config with CreateMissing enabled
    cfg := config.NewDefaultConfig()
    cfg.Build.CreateMissing = true

    bl := NewBookLoader(root, cfg)
    book, err := bl.Load()
    require.NoError(t, err)
    require.NotNil(t, book)

    // ch1/one.md should be created
    createdPath := filepath.Join(root, "src", "ch1", "one.md")
    require.FileExists(t, createdPath)
    content := testutil.ReadFile(t, root, filepath.Join("src", "ch1", "one.md"))
    assert.Contains(t, content, "# Chapter 1")

    // Validate items
    require.Len(t, book.Items, 3)

    // Intro chapter
    ch1, ok := book.Items[0].(*models.Chapter)
    require.True(t, ok)
    assert.False(t, ch1.IsDraft)
    require.NotNil(t, ch1.Path)
    assert.Equal(t, "intro.md", *ch1.Path)

    // Draft
    d, ok := book.Items[1].(*models.Chapter)
    require.True(t, ok)
    assert.True(t, d.IsDraft)
    assert.Nil(t, d.Path)

    // Chapter 1 created
    ch2, ok := book.Items[2].(*models.Chapter)
    require.True(t, ok)
    assert.False(t, ch2.IsDraft)
    require.NotNil(t, ch2.Path)
    assert.Equal(t, "ch1/one.md", filepath.ToSlash(*ch2.Path))
}
