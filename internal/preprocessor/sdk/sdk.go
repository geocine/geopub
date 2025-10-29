// Package sdk provides helpers for developing GeoPub preprocessors in Go.
//
// Example usage:
//
//	package main
//
//	import (
//		"log"
//		"github.com/geocine/geopub/internal/preprocessor/sdk"
//		"github.com/geocine/geopub/internal/preprocessor/runner"
//	)
//
//	func main() {
//		ctx, err := sdk.ReadContext()
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		// Modify ctx.Book here
//		// For example, replace tokens in chapter content
//		for _, section := range ctx.Book.Sections {
//			if section.Chapter != nil {
//				// Transform the chapter...
//			}
//		}
//
//		if err := sdk.WriteContext(ctx); err != nil {
//			log.Fatal(err)
//		}
//	}
package sdk

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/geocine/geopub/internal/preprocessor/runner"
)

// ReadContext reads a preprocessor context from stdin
// Returns the parsed PreprocessorContext
func ReadContext() (*runner.PreprocessorContext, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read stdin: %w", err)
	}

	ctx, err := runner.UnmarshalContext(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal context: %w", err)
	}

	return ctx, nil
}

// WriteContext writes a preprocessor context to stdout
// This is what will be read by GeoPub to apply mutations
func WriteContext(ctx *runner.PreprocessorContext) error {
	data, err := json.Marshal(ctx)
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	if _, err := os.Stdout.Write(data); err != nil {
		return fmt.Errorf("failed to write stdout: %w", err)
	}

	return nil
}

// Helper function to replace tokens in all chapter content
// This is a common operation for preprocessors
func ReplaceTokenInBook(book *runner.JsonBook, token string, replacement string) {
	if book == nil {
		return
	}
	for _, section := range book.Sections {
		if section.Chapter != nil {
			replaceTokenInChapter(section.Chapter, token, replacement)
		}
	}
}

// replaceTokenInChapter recursively replaces tokens in a chapter and sub-chapters
func replaceTokenInChapter(ch *runner.JsonChapter, token string, replacement string) {
	if ch == nil {
		return
	}

	// Simple string replacement in content
	// For more complex token processing, use strings package or regex
	ch.Content = replaceTokensInString(ch.Content, token, replacement)

	// Recurse into sub-items
	for _, subSection := range ch.SubItems {
		if subSection.Chapter != nil {
			replaceTokenInChapter(subSection.Chapter, token, replacement)
		}
	}
}

// replaceTokensInString is a helper to replace all occurrences of a token
// This is a basic implementation; for production use regex or template engines
func replaceTokensInString(content, token, replacement string) string {
	// Simple approach using strings.ReplaceAll
	// For case-insensitive or regex matching, use the strings or regexp packages
	var result string
	for len(content) > 0 {
		idx := findToken(content, token)
		if idx == -1 {
			result += content
			break
		}
		result += content[:idx] + replacement
		content = content[idx+len(token):]
	}
	return result
}

// findToken finds the index of a token in content
func findToken(content, token string) int {
	for i := 0; i <= len(content)-len(token); i++ {
		if content[i:i+len(token)] == token {
			return i
		}
	}
	return -1
}
