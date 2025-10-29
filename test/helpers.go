package testhelpers

import (
	"path/filepath"
	"runtime"
)

// RepoRoot returns the absolute path to the repository root.
func RepoRoot() string {
	// this file lives at <repo>/geopub/test/helpers.go
	_, file, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(file)
	repo := filepath.Dir(testDir)
	return repo
}

// IntegrationData joins under test/integration/testdata/...
func IntegrationData(parts ...string) string {
	base := []string{RepoRoot(), "test", "integration", "testdata"}
	return filepath.Join(append(base, parts...)...)
}

// GeoPubPath builds a path under test/integration/testdata/...
func GeoPubPath(parts ...string) string {
	return IntegrationData(parts...)
}

// GeoPubTestsuitePath builds a path under test/integration/testdata/testsuite/...
func GeoPubTestsuitePath(parts ...string) string {
	base := []string{"testsuite"}
	return IntegrationData(append(base, parts...)...)
}

// All tests should use GeoPubPath/GeoPubTestsuitePath for clarity
