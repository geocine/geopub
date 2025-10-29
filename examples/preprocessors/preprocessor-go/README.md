# GeoPub Go Preprocessor Example

This is an example preprocessor written in Go that demonstrates how to develop preprocessors using the GeoPub SDK.

## What it does

The Go example preprocessor reads a preprocessor context, calculates reading time for each chapter, and adds a note at the beginning of each chapter showing the estimated reading time.

## Setup

1. Make sure you have Go installed (1.16+)

2. Compile the preprocessor:
```bash
cd examples/preprocessor-go
go build -o geopub-go-example main.go
```

Or use it directly with `go run`:
```bash
go run examples/preprocessor-go/main.go
```

## Configuration

Add the preprocessor to your `book.toml`:

```toml
[preprocessor.go-example]
# Use compiled binary
command = "./examples/preprocessor-go/geopub-go-example"
# Or use go run directly
# command = "go run examples/preprocessor-go/main.go"
renderers = ["html"]
```

## How it works

The example demonstrates:

1. **Reading context**: Uses `sdk.ReadContext()` to parse JSON from stdin
2. **Processing**: Iterates through chapters, calculates reading time, adds note
3. **Writing back**: Uses `sdk.WriteContext()` to write modified context to stdout

## Using the SDK

The GeoPub preprocessor SDK provides helper functions:

```go
// Read context from stdin
ctx, err := sdk.ReadContext()
if err != nil {
    log.Fatal(err)
}

// Modify ctx.Book...

// Write modified context to stdout
if err := sdk.WriteContext(ctx); err != nil {
    log.Fatal(err)
}
```

### SDK Functions

- **`ReadContext() (*runner.PreprocessorContext, error)`** - Reads JSON from stdin
- **`WriteContext(ctx *runner.PreprocessorContext) error`** - Writes JSON to stdout
- **`ReplaceTokenInBook(book *runner.JsonBook, token, replacement string)`** - Helper to replace tokens

## Context Structure

The preprocessor receives a context with:

```go
type PreprocessorContext struct {
    Book     *JsonBook              // The book structure
    Config   map[string]interface{} // Configuration
    Renderer string                 // Renderer name ("html")
    Version  string                 // Protocol version
}

type JsonBook struct {
    Sections []JsonSection // Book sections
}

type JsonSection struct {
    Chapter     *JsonChapter // Chapter with content, subchapters, etc.
    IsSeparator bool         // True if this is a separator
}

type JsonChapter struct {
    Name     string        // Chapter title
    Content  string        // Chapter markdown content
    Number   []int         // Section number (e.g., [1, 2, 3])
    SubItems []JsonSection // Nested chapters
    Path     string        // File path
}
```

## Building as external binary

To make the preprocessor available system-wide:

```bash
# Build
go build -o geopub-go-example main.go

# Install to PATH
sudo mv geopub-go-example /usr/local/bin/

# Now use in book.toml as:
# [preprocessor.go-example]
# # No command needed, will resolve to geopub-go-example
```

## Error handling

If your preprocessor encounters an error:

```go
if err != nil {
    log.Fatalf("Error: %v", err)
}
```

The error will be:
1. Written to stderr
2. Cause non-zero exit code
3. GeoPub will report the error and halt the build

## Performance considerations

- Preprocessors run in subprocess, so there's JSON marshaling overhead
- For large books, optimize JSON processing
- Reading time calculation uses word count (~200 words/minute)

## Advanced examples

### Token replacement with templates

```go
import (
    "text/template"
)

// Use Go templates instead of simple token replacement
t := template.Must(template.New("").Parse(chapter.Content))
var buf bytes.Buffer
t.Execute(&buf, map[string]interface{}{
    "Author": "Jane Doe",
    "Date": time.Now().Format("2006-01-02"),
})
chapter.Content = buf.String()
```

### Adding metadata

```go
// Add YAML frontmatter to first chapter
firstChapter := ctx.Book.Sections[0].Chapter
if firstChapter != nil {
    metadata := `---
author: Jane Doe
generated: 2025-10-29
---

`
    firstChapter.Content = metadata + firstChapter.Content
}
```

### Dynamic content generation

```go
// Generate table of contents
var toc strings.Builder
for _, section := range ctx.Book.Sections {
    if section.Chapter != nil {
        toc.WriteString("- " + section.Chapter.Name + "\n")
    }
}
chapters[0].Content = "# Contents\n\n" + toc.String() + "\n\n" + chapters[0].Content
```

## See also

- [Node.js Token Replace Example](../preprocessor-token-replace/)
- [GeoPub Documentation](../../geopub/README.md)
