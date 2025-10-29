package runner

import (
	"testing"

	"github.com/geocine/geopub/internal/config"
	"github.com/geocine/geopub/internal/models"
)

func TestFrontmatterDisabledByDefault(t *testing.T) {
	// Create a config WITHOUT frontmatter preprocessor
	cfgStr := `
[book]
title = "Test"

[build]
use-default-preprocessors = true
`
	cfg, err := config.LoadFromString(cfgStr)
	if err != nil {
		t.Fatalf("LoadFromString() error: %v", err)
	}

	// Create a runner
	runner := NewRunner(cfg, "html")

	// Get execution order
	order, err := runner.GetExecutionOrder()
	if err != nil {
		t.Fatalf("GetExecutionOrder() error: %v", err)
	}

	// frontmatter should NOT be in the default execution order
	for _, name := range order {
		if name == "frontmatter" {
			t.Errorf("frontmatter should not be in default execution order, got: %v", order)
		}
	}

	// Only "index" should be in the default order
	if len(order) != 1 || order[0] != "index" {
		t.Errorf("default execution order should be [index], got: %v", order)
	}
}

func TestFrontmatterEnabledWhenConfigured(t *testing.T) {
	// Create a config WITH frontmatter preprocessor explicitly configured
	cfgStr := `
[book]
title = "Test"

[build]
use-default-preprocessors = true

[preprocessor.frontmatter]
# Explicitly enabled
`
	cfg, err := config.LoadFromString(cfgStr)
	if err != nil {
		t.Fatalf("LoadFromString() error: %v", err)
	}

	// Create a runner
	runner := NewRunner(cfg, "html")

	// Get execution order
	order, err := runner.GetExecutionOrder()
	if err != nil {
		t.Fatalf("GetExecutionOrder() error: %v", err)
	}

	// frontmatter SHOULD be in the execution order when configured
	found := false
	for _, name := range order {
		if name == "frontmatter" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("frontmatter should be in execution order when configured, got: %v", order)
	}

	// Should have both index and frontmatter
	if len(order) != 2 {
		t.Errorf("execution order should have 2 items, got: %v", order)
	}
}

func TestFrontmatterWithContentProcessing(t *testing.T) {
	// Integration test: verify frontmatter is actually stripped when enabled
	cfgStr := `
[book]
title = "Test"

[preprocessor.frontmatter]
`
	cfg, err := config.LoadFromString(cfgStr)
	if err != nil {
		t.Fatalf("LoadFromString() error: %v", err)
	}

	// Create a book with frontmatter content
	ch := models.NewChapter("Test", `---
author: Jane Doe
---

# Content

This is the content.`, "test.md", []string{})

	book := models.NewBook()
	book.PushItem(ch)

	// Process with runner (this should strip frontmatter if enabled)
	runner := NewRunner(cfg, "html")
	err = runner.Run(book)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Check that frontmatter was stripped
	result := book.Items[0].(*models.Chapter)
	if result.Content != "# Content\n\nThis is the content." {
		t.Errorf("frontmatter not stripped. Got: %q", result.Content)
	}
}
