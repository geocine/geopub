package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/geocine/geopub/internal/preprocessor/runner"
	"github.com/geocine/geopub/internal/preprocessor/sdk"
)

// Example preprocessor in Go
// This demonstrates how to use the GeoPub preprocessor SDK.
// It reads a preprocessor context, adds a note to each chapter, and writes it back.
//
// Usage in book.toml:
// [preprocessor.go-example]
// command = "go run examples/preprocessor-go/main.go"
// renderers = ["html"]

func main() {
	// Read the preprocessor context from stdin
	ctx, err := sdk.ReadContext()
	if err != nil {
		log.Fatalf("Failed to read context: %v", err)
	}

	if ctx.Book == nil || ctx.Book.Sections == nil {
		log.Fatal("Invalid context: missing book or sections")
	}

	// Example: Add a note to the beginning of each chapter
	for _, section := range ctx.Book.Sections {
		if section.Chapter != nil {
			section.Chapter.Content = addNoteToChapter(section.Chapter)
		}
	}

	// Write the modified context back to stdout
	if err := sdk.WriteContext(ctx); err != nil {
		log.Fatalf("Failed to write context: %v", err)
	}
}

// addNoteToChapter adds a note to the beginning of chapter content
func addNoteToChapter(ch *runner.JsonChapter) string {
	if ch == nil {
		return ""
	}

	// Add a note at the top of the chapter
	note := "<!-- This chapter was processed by the Go example preprocessor -->\n\n"

	// Count words in the chapter (simple heuristic)
	words := len(strings.Fields(ch.Content))
	readingTime := (words + 199) / 200 // Assume 200 words per minute
	if readingTime == 0 {
		readingTime = 1
	}

	note += fmt.Sprintf("> **Reading time**: ~%d minute(s)\n\n", readingTime)

	return note + ch.Content
}
