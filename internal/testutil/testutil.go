package testutil

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TempBook creates a temporary book directory structure for testing
func TempBook(t *testing.T, name string) string {
	tmpDir := t.TempDir()
	bookDir := filepath.Join(tmpDir, name)

	srcDir := filepath.Join(bookDir, "src")
	require.NoError(t, os.MkdirAll(srcDir, 0755))

	return bookDir
}

// WriteFile writes content to a file in the test directory
func WriteFile(t *testing.T, dir, path, content string) {
	fullPath := filepath.Join(dir, path)
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
	require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
}

// ReadFile reads content from a test file
func ReadFile(t *testing.T, dir, path string) string {
	fullPath := filepath.Join(dir, path)
	content, err := os.ReadFile(fullPath)
	require.NoError(t, err)
	return string(content)
}

// NormalizeHTML normalizes HTML for comparison (whitespace, attrs, etc.)
func NormalizeHTML(html string) string {
	// Collapse multiple whitespace
	html = regexp.MustCompile(`\s+`).ReplaceAllString(html, " ")

	// Remove spaces around tags
	html = regexp.MustCompile(`>\s+<`).ReplaceAllString(html, "><")

	// Trim leading/trailing space
	html = strings.TrimSpace(html)

	return html
}

// FileExists checks if a file exists
func FileExists(t *testing.T, path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists checks if a directory exists
func DirExists(t *testing.T, path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
