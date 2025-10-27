package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ReadToString reads a file into a string with error context
func ReadToString(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read '%s': %w", path, err)
	}
	return string(data), nil
}

// WriteFile writes content to a file, creating parent directories if needed
func WriteFile(path string, content []byte) error {
	if parent := filepath.Dir(path); parent != "." {
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return fmt.Errorf("failed to create directory '%s': %w", parent, err)
		}
	}

	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("failed to write '%s': %w", path, err)
	}

	return nil
}

// CreateDirAll creates a directory with better error messages
func CreateDirAll(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("failed to create directory '%s': %w", path, err)
	}
	return nil
}

// RemoveAll removes a file or directory tree with error context
func RemoveAll(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove '%s': %w", path, err)
	}
	return nil
}

// CopyFile copies a single file
func CopyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open '%s': %w", src, err)
	}
	defer source.Close()

	// Create destination directory if needed
	if parent := filepath.Dir(dst); parent != "." {
		if err := CreateDirAll(parent); err != nil {
			return err
		}
	}

	destination, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create '%s': %w", dst, err)
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return fmt.Errorf("failed to copy '%s' to '%s': %w", src, dst, err)
	}

	return nil
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// PathToRoot calculates relative path from the given path back to root
// E.g., "some/relative/path" -> "../../"
func PathToRoot(path string) string {
	dir := filepath.Dir(path)
	if dir == "." {
		return ""
	}

	depth := 0
	for dir != "." && dir != "" {
		depth++
		dir = filepath.Dir(dir)
	}

	result := ""
	for i := 0; i < depth; i++ {
		result += "../"
	}
	return result
}

// RemoveDirContents removes all contents of a directory but not the directory itself
func RemoveDirContents(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory '%s': %w", dir, err)
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf("failed to remove '%s': %w", path, err)
			}
		} else {
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove '%s': %w", path, err)
			}
		}
	}

	return nil
}
