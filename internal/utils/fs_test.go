package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathToRoot(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"file.md", ""},
		{filepath.Join("a", "file.md"), "../"},
		{filepath.Join("a", "b", "file.md"), "../../"},
		{filepath.Join("a", "b", "c", "file.md"), "../../../"},
	}

	for _, c := range cases {
		got := PathToRoot(c.in)
		assert.Equal(t, c.want, got, "input=%s", c.in)
	}
}

func TestRemoveDirContents(t *testing.T) {
	dir := t.TempDir()

	// Create nested structure
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sub", "nested"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("A"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sub", "b.txt"), []byte("B"), 0o644))

	// Sanity
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Greater(t, len(entries), 0)

	// Remove contents
	require.NoError(t, RemoveDirContents(dir))

	// Directory should exist but be empty
	after, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, after, 0)
}
