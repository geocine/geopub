package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/geocine/geopub/internal/utils"
)

// InitOptions captures options for initializing a new book
type InitOptions struct {
	Name          string
	CreateMissing bool   // whether to create missing files later (documented in book.toml)
	SrcDir        string // default: src
	BuildDir      string // default: book
	Title         string // optional book title; defaults to Name
}

// Init scaffolds a new book at the given directory
func Init(opts InitOptions) error {
	if opts.Name == "" {
		opts.Name = "my-book"
	}
	if opts.SrcDir == "" {
		opts.SrcDir = "src"
	}
	if opts.BuildDir == "" {
		opts.BuildDir = "book"
	}
	if opts.Title == "" {
		opts.Title = opts.Name
	}

	root := opts.Name

	// Create root directory
	if err := utils.CreateDirAll(root); err != nil {
		return err
	}

	// Create src directory
	srcPath := filepath.Join(root, opts.SrcDir)
	if err := utils.CreateDirAll(srcPath); err != nil {
		return err
	}

	// Write book.toml
	bookToml := []byte(fmt.Sprintf(`[book]
title = "%s"
src = "%s"

[build]
build-dir = "%s"
create-missing = %t
`, opts.Title, opts.SrcDir, opts.BuildDir, opts.CreateMissing))
	if err := utils.WriteFile(filepath.Join(root, "book.toml"), bookToml); err != nil {
		return err
	}

	// Write SUMMARY.md
	summary := []byte(`# Summary

[Introduction](README.md)

- [Chapter 1](ch1.md)
- [Conclusion](conclusion.md)
`)
	if err := utils.WriteFile(filepath.Join(srcPath, "SUMMARY.md"), summary); err != nil {
		return err
	}

	// Seed minimal chapters
	if err := utils.WriteFile(filepath.Join(srcPath, "README.md"), []byte("# Introduction\n\nWelcome to your new book!")); err != nil {
		return err
	}
	if err := utils.WriteFile(filepath.Join(srcPath, "ch1.md"), []byte("# Chapter 1\n\nStart writing here.")); err != nil {
		return err
	}
	if err := utils.WriteFile(filepath.Join(srcPath, "conclusion.md"), []byte("# Conclusion\n\nThanks for reading.")); err != nil {
		return err
	}

	// Create a .gitignore for the build dir
	gitignore := []byte(fmt.Sprintf("%s\n", opts.BuildDir))
	_ = utils.WriteFile(filepath.Join(root, ".gitignore"), gitignore)

	// Create a placeholder output directory (optional)
	if err := os.MkdirAll(filepath.Join(root, opts.BuildDir), 0o755); err != nil {
		return err
	}

	return nil
}
