package frontmatter

import (
	"regexp"

	"github.com/geocine/geopub/internal/models"
)

// FrontmatterPreprocessor strips YAML/TOML frontmatter from chapters
// This preprocessor is DISABLED BY DEFAULT
// Users must explicitly enable it in book.toml:
//
//	[preprocessor.frontmatter]
//	# No command needed - it's built-in
type FrontmatterPreprocessor struct{}

// NewFrontmatterPreprocessor creates a new frontmatter preprocessor
func NewFrontmatterPreprocessor() *FrontmatterPreprocessor {
	return &FrontmatterPreprocessor{}
}

// Name returns the preprocessor name
func (f *FrontmatterPreprocessor) Name() string {
	return "frontmatter"
}

// Process strips frontmatter from all chapters
func (f *FrontmatterPreprocessor) Process(book *models.Book) error {
	for _, item := range book.Items {
		if ch, ok := item.(*models.Chapter); ok {
			ch.Content = stripFrontmatter(ch.Content)
			// Recursively process sub-chapters
			f.processSubItems(ch.SubItems)
		}
	}
	return nil
}

// processSubItems recursively strips frontmatter from sub-items
func (f *FrontmatterPreprocessor) processSubItems(items []models.BookItem) {
	for _, item := range items {
		if ch, ok := item.(*models.Chapter); ok {
			ch.Content = stripFrontmatter(ch.Content)
			f.processSubItems(ch.SubItems)
		}
	}
}

// stripFrontmatter removes YAML or TOML frontmatter from content
// Frontmatter formats:
// - YAML: between --- delimiters
// - TOML: between +++ delimiters
// - Must be at the very start of the content
func stripFrontmatter(content string) string {
	// YAML frontmatter: --- ... --- (content can be empty)
	yamlPattern := regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n`)
	if yamlPattern.MatchString(content) {
		return yamlPattern.ReplaceAllString(content, "")
	}

	// TOML frontmatter: +++ ... +++ (content can be empty)
	tomlPattern := regexp.MustCompile(`(?s)^\+\+\+\s*\n(.*?)\n\+\+\+\s*\n`)
	if tomlPattern.MatchString(content) {
		return tomlPattern.ReplaceAllString(content, "")
	}

	// No frontmatter found
	return content
}
