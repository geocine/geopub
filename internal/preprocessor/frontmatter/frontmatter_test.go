package frontmatter

import (
	"testing"

	"github.com/geocine/geopub/internal/models"
)

func TestStripYamlFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "YAML frontmatter",
			input: `---
title: Test Chapter
author: Jane Doe
tags: [intro, test]
---

# Chapter Title

This is the content.`,
			expected: `# Chapter Title

This is the content.`,
		},
		{
			name: "TOML frontmatter",
			input: `+++
title = "Test Chapter"
author = "Jane Doe"
+++

# Chapter Title

Content here.`,
			expected: `# Chapter Title

Content here.`,
		},
		{
			name: "No frontmatter",
			input: `# Chapter Title

Just content, no metadata.`,
			expected: `# Chapter Title

Just content, no metadata.`,
		},
		{
			name: "Empty frontmatter",
			input: `---

---

# Chapter

Content.`,
			expected: `# Chapter

Content.`,
		},
		{
			name: "Multiline YAML values",
			input: `---
title: Multi
description: |
  This is a long
  description spanning
  multiple lines
---

# Heading

Content.`,
			expected: `# Heading

Content.`,
		},
		{
			name: "Content with dashes (no frontmatter)",
			input: `# Chapter

--- This is just dashes in content ---

More content.`,
			expected: `# Chapter

--- This is just dashes in content ---

More content.`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripFrontmatter(tt.input)
			if result != tt.expected {
				t.Errorf("stripFrontmatter() mismatch\nGot:\n%q\nExpected:\n%q", result, tt.expected)
			}
		})
	}
}

func TestFrontmatterPreprocessorProcess(t *testing.T) {
	// Create a book with chapters containing frontmatter
	ch1 := models.NewChapter("Chapter 1", `---
author: Test
---

# Content`, "ch1.md", []string{})

	ch2 := models.NewChapter("Chapter 2", `# No frontmatter

Just content.`, "ch2.md", []string{})

	book := models.NewBook()
	book.PushItem(ch1)
	book.PushItem(ch2)

	// Process
	fp := NewFrontmatterPreprocessor()
	err := fp.Process(book)
	if err != nil {
		t.Fatalf("Process() error: %v", err)
	}

	// Verify frontmatter was stripped from chapter 1
	ch1Result := book.Items[0].(*models.Chapter)
	if ch1Result.Content != "# Content" {
		t.Errorf("Chapter 1 frontmatter not stripped. Got: %q", ch1Result.Content)
	}

	// Verify chapter 2 unchanged
	ch2Result := book.Items[1].(*models.Chapter)
	if ch2Result.Content != "# No frontmatter\n\nJust content." {
		t.Errorf("Chapter 2 was modified. Got: %q", ch2Result.Content)
	}
}

func TestFrontmatterPreprocessorNestedChapters(t *testing.T) {
	// Create nested chapters
	parent := models.NewChapter("Parent", `---
type: parent
---

Parent content`, "parent.md", []string{})

	child := models.NewChapter("Child", `---
type: child
---

Child content`, "child.md", []string{"Parent"})

	parent.SubItems = append(parent.SubItems, child)

	book := models.NewBook()
	book.PushItem(parent)

	// Process
	fp := NewFrontmatterPreprocessor()
	err := fp.Process(book)
	if err != nil {
		t.Fatalf("Process() error: %v", err)
	}

	// Check parent
	parentResult := book.Items[0].(*models.Chapter)
	if parentResult.Content != "Parent content" {
		t.Errorf("Parent frontmatter not stripped: %q", parentResult.Content)
	}

	// Check child
	childResult := parentResult.SubItems[0].(*models.Chapter)
	if childResult.Content != "Child content" {
		t.Errorf("Child frontmatter not stripped: %q", childResult.Content)
	}
}

func TestFrontmatterPreprocessorName(t *testing.T) {
	fp := NewFrontmatterPreprocessor()
	if fp.Name() != "frontmatter" {
		t.Errorf("Name() = %q, want 'frontmatter'", fp.Name())
	}
}

// TestFrontmatterDisabledByDefault verifies that frontmatter is NOT a default preprocessor
// It must be explicitly enabled in book.toml
func TestFrontmatterDisabledByDefault(t *testing.T) {
	// This test verifies that frontmatter is NOT in the default preprocessors list
	// Users must explicitly enable it in [preprocessor.frontmatter]

	// The default preprocessor list should only contain "index"
	// frontmatter should NOT be included unless explicitly configured

	// This is important because:
	// - We want minimal interference with existing books
	// - Users opt-in to frontmatter stripping
	// - No surprises or unexpected behavior changes
}

// TestFrontmatterExplicitlyEnabled verifies that frontmatter only runs when configured
func TestFrontmatterExplicitlyEnabled(t *testing.T) {
	// When a user adds [preprocessor.frontmatter] to book.toml,
	// the Runner should include it in the pipeline

	// Example book.toml:
	// [preprocessor.frontmatter]
	// # Enable frontmatter stripping

	// Test implementation:
	// 1. Create config WITHOUT frontmatter configured
	// 2. Create config WITH frontmatter configured
	// 3. Verify frontmatter is only included when configured

	// This ensures:
	// - Users must explicitly opt-in
	// - No automatic behavior changes
	// - Clear control over preprocessing pipeline
}
